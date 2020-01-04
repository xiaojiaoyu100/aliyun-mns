package alimns

import (
	"errors"
	"fmt"
	"net/http"
)

// SetTopicAttributes 修改主题的属性
func (c *Client) SetTopicAttributes(topic string, setters ...TopicAttributeSetter) error {
	var err error

	if !checkTopicName(topic) {
		return errors.New("unqualified topic name")
	}

	attri := DefaultTopicAttr()

	for _, setter := range setters {
		err = setter(&attri)
		if err != nil {
			return err
		}
	}

	requestLine := fmt.Sprintf(mnsSetTopicAttributes, topic)
	req := c.ca.NewRequest().Put().WithPath(requestLine).WithXMLBody(&attri).WithTimeout(apiTimeout)

	resp, err := c.ca.Do(req)
	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil
	default:
		var respErr RespErr
		if err := resp.DecodeFromXML(&respErr); err != nil {
			return nil
		}

		switch respErr.Code {
		case topicNotExistError.Error():
			return topicNotExistError
		default:
			return errors.New(respErr.Code)
		}
	}
}
