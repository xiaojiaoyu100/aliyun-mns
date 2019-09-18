package alimns

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// SendMessageResponse 发送消息回复
type SendMessageResponse struct {
	SendMessage
}

// SendBase64EncodedJSONMessage 发送base64编码的json消息
func (c *Client) SendBase64EncodedJSONMessage(name string, messageBody interface{},
	setters ...MessageSetter) (*SendMessageResponse, error) {
	var (
		resp *SendMessageResponse
		err  error
	)
	ended := make(chan struct{})
	go func() {
		for {
			resp, err = c.sendBase64EncodedJSONMessage(name, messageBody, setters...)
			switch {
			case shouldRetry(err):
				time.Sleep(time.Millisecond * 100)
				continue
			default:
				close(ended)
				return
			}
		}
	}()
	select {
	case <-time.After(time.Second * 60):
		return nil, sendMessageTimeoutError
	case <-ended:
		return resp, err
	}
}

func (c *Client) sendBase64EncodedJSONMessage(name string, body interface{}, setters ...MessageSetter) (*SendMessageResponse, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	c.log.WithField("queue", name).WithField("body", body).Info("sendBase64EncodedJSONMessage")
	b64Body := base64.StdEncoding.EncodeToString(b)
	c.log.WithField("queue", name).WithField("b64body", b64Body).Info("sendBase64EncodedJSONMessage")
	resp, err := c.sendMessage(name, b64Body, setters...)
	if err != nil {
		c.log.WithField("queue", name).WithError(err).WithField("body", body).Error("sendBase64EncodedJSONMessage")
	}
	return resp, err
}

func (c *Client) sendMessage(name, messageBody string, setters ...MessageSetter) (*SendMessageResponse, error) {
	var err error

	if len(messageBody) > maxMessageSize {
		return nil, messageBodyLimitError
	}

	attri := DefaultMessage()
	attri.MessageBody = messageBody
	for _, setter := range setters {
		err = setter(&attri)
		if err != nil {
			return nil, err
		}
	}

	requestLine := fmt.Sprintf(mnsSendMessage, name)
	req := c.ca.NewRequest().Post().WithPath(requestLine).WithXMLBody(&attri)

	resp, err := c.ca.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case http.StatusCreated:
		var sendMessageResponse SendMessageResponse
		if err := resp.DecodeFromXML(&sendMessageResponse); err != nil {
			return nil, err
		}
		return &sendMessageResponse, nil
	default:
		var respErr RespErr
		if err := resp.DecodeFromXML(&respErr); err != nil {
			return nil, err
		}
		switch respErr.Code {
		case queueNotExistError.Error():
			return nil, queueNotExistError
		case internalError.Error():
			return nil, internalError
		}
		return nil, errors.New(respErr.Message)
	}
}
