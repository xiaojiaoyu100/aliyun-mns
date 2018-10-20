package aliyun_mns

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/go-querystring/query"
)

type ReceiveMessage struct {
	XMLName          xml.Name `xml:"Message"`
	XmlNs            string   `xml:"xmlns,attr"`
	MessageId        string   `xml:"MessageId"`
	ReceiptHandle    string   `xml:"ReceiptHandle"`
	MessageBody      string   `xml:"MessageBody"`
	MessageBodyMD5   string   `xml:"MessageBodyMD5"`
	EnqueueTime      int64    `xml:"EnqueueTime"`
	NextVisibleTime  int64    `xml:"NextVisibleTime"`
	FirstDequeueTime int64    `xml:"FirstDequeueTime"`
	DequeueCount     int      `xml:"DequeueCount"`
	Priority         int      `xml:"Priority"`
}

type ReceiveMessageResponse struct {
	ReceiveMessage
}

type ReceiveMessageParam struct {
	WaitSeconds   *int `url:"waitseconds,omitempty"`
	NumOfMessages int  `url:"numOfMessages"`
}

func DefaultReceiveMessage() ReceiveMessageParam {
	return ReceiveMessageParam{}
}

type ReceiveMessageParamSetter func(*ReceiveMessageParam) error

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

func WithReceiveMessageNumOfMessages(num int) ReceiveMessageParamSetter {
	return func(rm *ReceiveMessageParam) error {
		if num < minReceiveMessage || num > maxReceiveMessage {
			return errors.New("num of receive message out of range")
		}
		rm.NumOfMessages = num
		return nil
	}
}

func (c *Client) ReceiveMessage(name string, setters ...ReceiveMessageParamSetter) (*ReceiveMessageResponse, error) {

	receiveMessage := DefaultReceiveMessage()
	for _, setter := range setters {
		if err := setter(&receiveMessage); err != nil {
			return nil, err
		}
	}

	requestLine := fmt.Sprintf(mnsReceiveMessage, name)
	req, err := http.NewRequest(http.MethodGet, c.endpoint+requestLine, nil)
	if err != nil {
		return nil, err
	}
	values, err := query.Values(&receiveMessage)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()

	c.finalizeHeader(req, nil)

	globalLogger.printf("消费消息请求: %s %s", req.Method, req.URL.String())

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

	globalLogger.printf("消费消息回复: %s %s", resp.Status, string(body))

	switch resp.StatusCode {
	case http.StatusOK:
		var receiveMessageResponse ReceiveMessageResponse
		if err := xml.Unmarshal(body, &receiveMessageResponse); err != nil {
			return nil, err
		}
		return &receiveMessageResponse, nil
	default:
		var respErr RespErr
		if err := xml.Unmarshal(body, &respErr); err != nil {
			return nil, err
		}
		return nil, errors.New(respErr.Code)
	}
}
