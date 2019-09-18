package alimns

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
)

// BatchReceiveMessageResponse 批量消費消息
type BatchReceiveMessageResponse struct {
	XMLName         xml.Name          `xml:"Messages"`
	XMLNs           string            `xml:"xmlns,attr"`
	ReceiveMessages []*ReceiveMessage `xml:"Message"`
}

// DefaultBatchReceiveMessage 返回默认的批量消费消息的参数
func DefaultBatchReceiveMessage() ReceiveMessageParam {
	return ReceiveMessageParam{
		NumOfMessages: defaultReceiveMessage,
	}
}

// BatchReceiveMessage 批量消费消息
func (c *Client) BatchReceiveMessage(name string, setters ...ReceiveMessageParamSetter) (*BatchReceiveMessageResponse, error) {
	var err error

	receiveMessage := DefaultBatchReceiveMessage()
	for _, setter := range setters {
		err = setter(&receiveMessage)
		if err != nil {
			return nil, err
		}
	}

	requestLine := fmt.Sprintf(mnsBatchReceiveMessage, name)
	req := c.ca.NewRequest().Get().WithPath(requestLine).WithQueryParam(&receiveMessage)

	resp, err := c.ca.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var batchReceiveMessageResponse BatchReceiveMessageResponse
		if err := resp.DecodeFromXML(&batchReceiveMessageResponse); err != nil {
			return nil, err
		}
		return &batchReceiveMessageResponse, nil
	default:
		var respErr RespErr
		if err := resp.DecodeFromXML(&respErr); err != nil {
			return nil, err
		}

		switch respErr.Code {
		case messageNotExistError.Error():
			return nil, messageNotExistError
		case queueNotExistError.Error():
			return nil, queueNotExistError
		}

		return nil, errors.New(respErr.Code)
	}
}
