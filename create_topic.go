package alimns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// CreateTopic 创建一个主题
func (c *Client) CreateTopic(name string, setters ...TopicAttributeSetter) error {
	var err error

	if !checkTopicName(name) {
		return errors.New("unqualified topic name")
	}

	attri := DefaultTopicAttr()

	for _, setter := range setters {
		err = setter(&attri)
		if err != nil {
			return err
		}
	}

	requestLine := fmt.Sprintf(mnsCreateTopic, name)
	req := c.ca.NewRequest().Put().WithPath(requestLine).WithXMLBody(&attri).WithTimeout(apiTimeout)

	resp, err := c.ca.Do(context.TODO(), req)
	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case http.StatusCreated:
		return nil
	default:
		var respErr RespErr
		if err := resp.DecodeFromXML(&respErr); err != nil {
			return nil
		}

		switch respErr.Code {
		case topicAlreadyExistError.Error():
			return topicAlreadyExistError
		case topicNameLengthError.Error():
			return topicNameLengthError
		default:
			return errors.New(respErr.Code)
		}
	}
}
