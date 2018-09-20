package aliyun_mns

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type ChangeVisibilityTimeoutResponse struct {
	XMLName         xml.Name `xml:"ChangeVisibility"`
	XmlNs           string   `xml:"xmlns,attr"`
	ReceiptHandle   string   `xml:"ReceiptHandle"`
	NextVisibleTime int64    `xml:"NextVisibleTime"`
}

func (c *Client) ChangeVisibilityTimeout(name string, receiptHandle string, visibilityTimeout int) (*ChangeVisibilityTimeoutResponse, error) {
	if visibilityTimeout < minVisibilityTimeout || visibilityTimeout > maxVisibilityTimeout {
		return nil, visibilityTimeoutError
	}

	requestLine := fmt.Sprintf(mnsChangeMessageVisibility, name, receiptHandle, strconv.Itoa(visibilityTimeout))
	req, err := http.NewRequest(http.MethodPut, c.endpoint+requestLine, nil)
	if err != nil {
		return nil, err
	}

	c.finalizeHeader(req, nil)

	globalLogger.printf("修改消息可见时间请求: %s %s", req.Method, req.URL.String())

	ctx, cancel := context.WithCancel(context.TODO())
	_ = time.AfterFunc(time.Second*timeout, func() {
		cancel()
	})
	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	globalLogger.printf("修改消息可见时间回复: %s %s", resp.Status, string(body))

	switch resp.StatusCode {
	case http.StatusOK:
		var changeVisibilityTimeoutResponse ChangeVisibilityTimeoutResponse
		if err := xml.Unmarshal(body, &changeVisibilityTimeoutResponse); err != nil {
			return nil, err
		}
		return &changeVisibilityTimeoutResponse, nil
	default:
		return nil, unknownError
	}
}
