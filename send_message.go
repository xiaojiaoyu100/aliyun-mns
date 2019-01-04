package alimns

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// SendMessageResponse 发送消息回复
type SendMessageResponse struct {
	SendMessage
}

// SendBase64EncodedJSONMessage 发送base64编码的json消息
func (c *Client) SendBase64EncodedJSONMessage(name string, messageBody interface{}, setters ...MessageSetter) (*SendMessageResponse, error) {
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
	contextLogger.WithField("queue", name).WithField("body", body).Info("sendBase64EncodedJSONMessage")
	b64Body := base64.StdEncoding.EncodeToString(b)
	return c.sendMessage(name, b64Body, setters...)
}

func (c *Client) sendMessage(name, messageBody string, setters ...MessageSetter) (*SendMessageResponse, error) {
	if len(messageBody) > maxMessageSize {
		return nil, messageBodyLimitError
	}
	attri := DefaultMessage()
	attri.MessageBody = messageBody
	for _, setter := range setters {
		if err := setter(&attri); err != nil {
			return nil, err
		}
	}

	body, err := xml.Marshal(&attri)
	if err != nil {
		return nil, err
	}

	requestLine := fmt.Sprintf(mnsSendMessage, name)
	req, err := http.NewRequest(http.MethodPost, c.endpoint+requestLine, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	c.finalizeHeader(req, body)

	contextLogger.
		WithField("method", req.Method).
		WithField("url", req.URL.String()).
		WithField("body", string(body)).
		Info("发送消息请求")

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

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	contextLogger.
		WithField("status", resp.Status).
		WithField("body", string(body)).
		WithField("url", req.URL.String()).
		Info("发送消息回复")

	switch resp.StatusCode {
	case http.StatusCreated:
		var sendMessageResponse SendMessageResponse
		if err := xml.Unmarshal(body, &sendMessageResponse); err != nil {
			return nil, err
		}
		return &sendMessageResponse, nil
	default:
		var respErr RespErr
		if err := xml.Unmarshal(body, &respErr); err != nil {
			return nil, err
		}
		switch respErr.Code {
		case queueNotExistError.Error():
			return nil, queueNotExistError
		}
		return nil, errors.New(respErr.Message)
	}
}
