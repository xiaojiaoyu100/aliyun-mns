package alimns

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// QueueAttributeResponse 消息队列属性回复
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

// GetQueueAttributes 获取消息队列属性
func (c *Client) GetQueueAttributes(name string) (*QueueAttributeResponse, error) {
	requestLine := fmt.Sprintf(mnsGetQueueAttributes, name)
	req, err := http.NewRequest(http.MethodGet, c.endpoint+requestLine, nil)
	if err != nil {
		return nil, err
	}
	c.finalizeHeader(req, nil)

	contextLogger.
		WithField("method", req.Method).
		WithField("url", req.URL.String()).
		Info("获取队列属性请求")

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

	contextLogger.
		WithField("status", resp.Status).
		WithField("body", string(body)).
		WithField("url", req.URL.String()).
		Info("获取队列属性回复")

	switch resp.StatusCode {
	case http.StatusOK:

		var queueAttributeResponse QueueAttributeResponse

		if err := xml.Unmarshal(body, &queueAttributeResponse); err != nil {
			return nil, err
		}

		return &queueAttributeResponse, nil

	default:
		var respErr RespErr
		if err := xml.Unmarshal(body, &respErr); err != nil {
			return nil, err
		}
		return nil, errors.New(respErr.Code)
	}

}
