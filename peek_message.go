package alimns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// PeekMessage 查看消息
type PeekMessage struct {
	MessageID        string `xml:"MessageId"`
	MessageBody      string `xml:"MessageBody"`
	MessageBodyMD5   string `xml:"MessageBodyMD5"`
	EnqueueTime      int64  `xml:"EnqueueTime"`
	FirstDequeueTime int64  `xml:"FirstDequeueTime"`
	DequeueCount     int    `xml:"DequeueCount"`
	Priority         int    `xml:"Priority"`
}

// PeekMessageResponse 查看消息回复
type PeekMessageResponse struct {
	PeekMessage
}

// PeekMessage 查看消息
func (c *Client) PeekMessage(name string) (*PeekMessageResponse, error) {
	var err error

	requestLine := fmt.Sprintf(mnsPeekMessage, name)
	req := c.ca.NewRequest().Get().WithPath(requestLine).WithTimeout(apiTimeout)

	resp, err := c.ca.Do(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var peekMessageResponse PeekMessageResponse
		err = resp.DecodeFromXML(&peekMessageResponse)
		if err != nil {
			return nil, err
		}
		return &peekMessageResponse, nil
	default:
		var respErr RespErr
		err = resp.DecodeFromXML(&respErr)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(respErr.Code)
	}
}
