package alimns

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
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
	var err error

	requestLine := fmt.Sprintf(mnsGetQueueAttributes, name)
	req := c.ca.NewRequest().Get().WithPath(requestLine).WithTimeout(apiTimeout)

	resp, err := c.ca.Do(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var queueAttributeResponse QueueAttributeResponse
		if err := resp.DecodeFromXML(&queueAttributeResponse); err != nil {
			return nil, err
		}
		return &queueAttributeResponse, nil
	default:
		var respErr RespErr
		if err := resp.DecodeFromXML(&respErr); err != nil {
			return nil, err
		}
		return nil, errors.New(respErr.Code)
	}

}
