package aliyun_mns

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type QueueAttributeResponse struct {
	XMLName                xml.Name `xml:"Queue"`
	QueueName              string   `xml:"QueueName"`
	CreateTime             string   `xml:"CreateTime"`
	LastModifyTime         string   `xml:"LastModifyTime"`
	DelaySeconds           int      `xml:"DelaySeconds"`
	MaximumMessageSize     int      `xml:"MaximumMessageSize"`
	MessageRetentionPeriod int      `xml:"MessageRetentionPeriod"`
	VisibilityTimeout      int      `xml:"VisibilityTimeout"`
	PollingWaitSeconds     int      `xml:"PollingWaitSeconds"`
	Activemessages         string   `xml:"Activemessages"`
	InactiveMessages       string   `xml:"InactiveMessages"`
	DelayMessages          string   `xml:"DelayMessages"`
	LoggingEnabled         bool     `xml:"LoggingEnabled"`
}

func (c *Client) GetQueueAttributes(name string) (*QueueAttributeResponse, error) {
	requestLine := fmt.Sprintf(mnsGetQueueAttributes, name)
	req, err := http.NewRequest(http.MethodGet, c.endpoint+requestLine, nil)
	if err != nil {
		return nil, err
	}
	c.finalizeHeader(req, nil)

	globalLogger.printf("获取队列属性请求: %s %s", req.Method, req.URL.String())

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

	globalLogger.printf("获取队列属性回复: %s %s", resp.Status, string(body))

	switch resp.StatusCode {
	case http.StatusOK:

		var queueAttributeResponse QueueAttributeResponse

		if err := xml.Unmarshal(body, &queueAttributeResponse); err != nil {
			return nil, err
		}

		return &queueAttributeResponse, nil

	default:
		return nil, unknownError
	}

}
