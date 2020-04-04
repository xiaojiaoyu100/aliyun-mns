package alimns

import (
	"context"
	"encoding/xml"
	"errors"
	"net/http"
)

// ListQueueRequest 获取队列列表参数
type ListQueueRequest struct {
	Marker    string
	RetNumber string
	Prefix    string
}

// QueueData 队列的属性
type QueueData struct {
	QueueURL string `xml:"QueueURL"`
}

// ListQueueResponse 获取队列的回复
type ListQueueResponse struct {
	XMLName    xml.Name     `xml:"Queues"`
	XMLNs      string       `xml:"xmlns,attr"`
	Queues     []*QueueData `xml:"Queue"`
	NextMarker string       `xml:"NextMarker"`
}

// ListQueue 请求队列列表
func (c *Client) ListQueue(request *ListQueueRequest) (*ListQueueResponse, error) {
	var err error

	req := c.ca.NewRequest().Get().WithPath(mnsListQueue).WithTimeout(apiTimeout)

	if request.Marker != "" {
		req.SetHeader(xMnsMarker, request.Marker)
	}
	if request.RetNumber != "" {
		req.SetHeader(xMnsRetNumber, request.RetNumber)
	}
	if request.Prefix != "" {
		req.SetHeader(xMnxPrefix, request.Prefix)
	}

	resp, err := c.ca.Do(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var response ListQueueResponse
		if err := resp.DecodeFromXML(&response); err != nil {
			return nil, err
		}
		return &response, nil
	default:
		var respErr RespErr
		if err := resp.DecodeFromXML(&respErr); err != nil {
			return nil, err
		}
		return nil, errors.New(respErr.Code)
	}
}
