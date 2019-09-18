package alimns

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
)

// ReceiveMessage 收到消息
type ReceiveMessage struct {
	XMLName          xml.Name `xml:"Message"`
	XMLNs            string   `xml:"xmlns,attr"`
	MessageID        string   `xml:"MessageId"`
	ReceiptHandle    string   `xml:"ReceiptHandle"`
	MessageBody      string   `xml:"MessageBody"`
	MessageBodyMD5   string   `xml:"MessageBodyMD5"`
	EnqueueTime      int64    `xml:"EnqueueTime"`
	NextVisibleTime  int64    `xml:"NextVisibleTime"`
	FirstDequeueTime int64    `xml:"FirstDequeueTime"`
	DequeueCount     int      `xml:"DequeueCount"`
	Priority         int      `xml:"Priority"`
}

// ReceiveMessageResponse 收到消息回复
type ReceiveMessageResponse struct {
	ReceiveMessage
}

// ReceiveMessageParam 收到消息请求
type ReceiveMessageParam struct {
	WaitSeconds   *int `url:"waitseconds,omitempty"`
	NumOfMessages int  `url:"numOfMessages"`
}

// DefaultReceiveMessage 默认的收到消息请求参数
func DefaultReceiveMessage() ReceiveMessageParam {
	return ReceiveMessageParam{}
}

// ReceiveMessageParamSetter 收到消息请求参数设置函数
type ReceiveMessageParamSetter func(*ReceiveMessageParam) error

// WithReceiveMessageWaitSeconds 设置收到消息的long poll等待时长
func WithReceiveMessageWaitSeconds(s int) ReceiveMessageParamSetter {
	return func(rm *ReceiveMessageParam) error {
		if s < minPollingWaitSeconds || s > maxPollingWaitSeconds {
			return errors.New("polling wait seconds out of range")
		}
		rm.WaitSeconds = &s
		return nil
	}
}

const (
	defaultReceiveMessage = 16
	minReceiveMessage     = 1
	maxReceiveMessage     = 16
)

// WithReceiveMessageNumOfMessages 设置请求消息数量
func WithReceiveMessageNumOfMessages(num int) ReceiveMessageParamSetter {
	return func(rm *ReceiveMessageParam) error {
		if num < minReceiveMessage || num > maxReceiveMessage {
			return errors.New("num of receive message out of range")
		}
		rm.NumOfMessages = num
		return nil
	}
}

// ReceiveMessage 接收消息
func (c *Client) ReceiveMessage(name string, setters ...ReceiveMessageParamSetter) (*ReceiveMessageResponse, error) {
	var err error
	receiveMessage := DefaultReceiveMessage()
	for _, setter := range setters {
		err = setter(&receiveMessage)
		if err != nil {
			return nil, err
		}
	}

	requestLine := fmt.Sprintf(mnsReceiveMessage, name)
	req := c.ca.NewRequest().Get().WithPath(requestLine).WithQueryParam(&receiveMessage)

	resp, err := c.ca.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var receiveMessageResponse ReceiveMessageResponse
		if err := resp.DecodeFromXML(&receiveMessageResponse); err != nil {
			return nil, err
		}
		return &receiveMessageResponse, nil
	default:
		var respErr RespErr
		if err := resp.DecodeFromXML(&respErr); err != nil {
			return nil, err
		}
		return nil, errors.New(respErr.Code)
	}
}
