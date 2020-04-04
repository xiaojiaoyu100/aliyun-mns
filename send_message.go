package alimns

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/vmihailenco/msgpack"
)

// SendMessageResponse 发送消息回复
type SendMessageResponse struct {
	SendMessage
}

// SendBase64EncodedJSONMessage 发送base64编码的json消息
func (c *Client) SendBase64EncodedJSONMessage(name string, messageBody interface{},
	setters ...MessageSetter) (*SendMessageResponse, error) {
	return c.sendBase64EncodedJSONMessage(name, messageBody, setters...)
}

func (c *Client) sendBase64EncodedJSONMessage(name string, body interface{}, setters ...MessageSetter) (*SendMessageResponse, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	c.log.WithField("queue", name).WithField("body", body).Info("sendBase64EncodedJSONMessage")
	b64Body := base64.StdEncoding.EncodeToString(b)
	c.log.WithField("queue", name).WithField("b64body", b64Body).Info("sendBase64EncodedJSONMessage")
	return c.sendMessage(name, b64Body, setters...)
}

func (c *Client) send(name string, message Message) (*SendMessageResponse, error) {
	var err error
	requestLine := fmt.Sprintf(mnsSendMessage, name)
	req := c.ca.NewRequest().Post().WithPath(requestLine).WithXMLBody(&message)
	resp, err := c.ca.Do(context.TODO(), req)
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

func (c *Client) pushRetryQueue(name string, message Message) {
	if c.config.Cmdable == nil {
		return
	}

	w := wrapper{
		QueueName: name,
		Message:   message,
	}

	b, err := msgpack.Marshal(w)
	if err != nil {
		c.log.WithError(err).Errorf("msgpack.Marshal: %v", w)
		return
	}

	_, err = c.config.RPush(aliyunMnsRetryQueue, string(b)).Result()
	if err != nil {
		c.log.WithError(err).Errorf("RPush: %s", b)
		return
	}
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

	resp, err := c.ca.Do(context.TODO(), req)
	if shouldRetry(err) {
		c.pushRetryQueue(name, attri)
	}

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
