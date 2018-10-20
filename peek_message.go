package aliyun_mns

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"errors"
)

type PeekMessage struct {
	MessageId        string `xml:"MessageId"`
	MessageBody      string `xml:"MessageBody"`
	MessageBodyMD5   string `xml:"MessageBodyMD5"`
	EnqueueTime      int64  `xml:"EnqueueTime"`
	FirstDequeueTime int64  `xml:"FirstDequeueTime"`
	DequeueCount     int    `xml:"DequeueCount"`
	Priority         int    `xml:"Priority"`
}

type PeekMessageResponse struct {
	PeekMessage
}

func (c *Client) PeekMessage(name string) (*PeekMessageResponse, error) {
	requestLine := fmt.Sprintf(mnsPeekMessage, name)
	req, err := http.NewRequest(http.MethodGet, c.endpoint+requestLine, nil)
	if err != nil {
		return nil, err
	}

	c.finalizeHeader(req, nil)

	globalLogger.printf("查看消息请求: %s %s", req.Method, req.URL.String())

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

	globalLogger.printf("查看消息回复: %s %s", resp.Status, string(body))

	switch resp.StatusCode {
	case http.StatusOK:
		var peekMessageResponse PeekMessageResponse
		if err := xml.Unmarshal(body, &peekMessageResponse); err != nil {
			return nil, err
		}
		return &peekMessageResponse, nil
	default:
		var respErr RespErr
		if err := xml.Unmarshal(body, &respErr); err != nil {
			return nil, err
		}
		return nil, errors.New(respErr.Code)
	}
}
