package alimns

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
)

type messageResponse struct {
	XMLName        xml.Name `xml:"Message"`
	MessageID      string   `xml:"MessageId"`
	MessageBodyMD5 string   `xml:"MessageBodyMD5"`
}

// PublishMessage 向指定的主题发布消息，消息发布到主题后随即会被推送给Endpoint消费
func (c *Client) PublishMessage(topic, message string, setters ...PublishMessageParamSetter) (string, error) {
	var err error

	if !checkTopicName(topic) {
		return "", errors.New("unqualified topic name")
	}

	if message == "" {
		return "", errors.New("empty message")
	}

	publishMessageParam := PublishMessageParam{}

	for _, setter := range setters {
		err = setter(&publishMessageParam)
		if err != nil {
			return "", err
		}
	}

	publishMessageParam.MessageBody = message
	requestLine := fmt.Sprintf(mnsPublishMessage, topic)
	req := c.ca.NewRequest().Post().WithPath(requestLine).WithXMLBody(&publishMessageParam).WithTimeout(apiTimeout)

	resp, err := c.ca.Do(context.TODO(), req)
	if err != nil {
		return "", err
	}

	switch resp.StatusCode() {
	case http.StatusCreated:
		var messageResp messageResponse
		if err := resp.DecodeFromXML(&messageResp); err != nil {
			return "", err
		}

		return messageResp.MessageID, nil

	default:
		var respErr RespErr
		if err := resp.DecodeFromXML(&respErr); err != nil {
			return "", nil
		}

		switch respErr.Code {
		case topicNotExistError.Error():
			return "", topicNotExistError
		default:
			return "", errors.New(respErr.Code)
		}
	}
}
