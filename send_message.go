package aliyun_mns

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type SendMessageResponse struct {
	XMLName        xml.Name `xml:"Message"`
	XmlNs          string   `xml:"xmlns,attr"`
	MessageId      string   `xml:"MessageId"`
	MessageBodyMD5 string   `xml:"MessageBodyMD5"`
	ReceiptHandle  string   `xml:"ReceiptHandle"` // 发送延时消息才有返回
}

func (c *Client) SendBase64EncodedJsonMessage(name string, messageBody interface{}, setters ...MessageSetter) (*SendMessageResponse, error) {
	var (
		resp *SendMessageResponse
		err  error
	)
	ended := make(chan struct{})
	go func() {
		for {
			resp, err = c.sendBase64EncodedJsonMessage(name, messageBody, setters...)
			switch {
			case isNetworkError(err):
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

func (c *Client) SendMessage(name string, messageBody string, setters ...MessageSetter) (*SendMessageResponse, error) {
	var (
		resp *SendMessageResponse
		err  error
	)
	ended := make(chan struct{})
	go func() {
		for {
			resp, err = c.sendMessage(name, messageBody, setters...)
			switch {
			case isNetworkError(err):
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

func (c *Client) sendBase64EncodedJsonMessage(name string, body interface{}, setters ...MessageSetter) (*SendMessageResponse, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
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

	globalLogger.printf("发送消息请求: %s %s %s", req.Method, req.URL.String(), string(body))

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

	globalLogger.printf("发送消息回复: %s %s", resp.Status, string(body))

	switch resp.StatusCode {
	case http.StatusCreated:
		var sendMessageResponse SendMessageResponse
		if err := xml.Unmarshal(body, &sendMessageResponse); err != nil {
			return nil, err
		}
		return &sendMessageResponse, nil
	default:
		return nil, unknownError
	}
}
