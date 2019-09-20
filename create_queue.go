package alimns

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

func checkQueueName(name string) bool {
	regex := regexp.MustCompile(`^[0-9a-zA-Z]{1}[0-9a-zA-Z-]{0,255}$`)
	return regex.MatchString(name)
}

// CreateQueue 创建一个消息队列
func (c *Client) CreateQueue(name string, setters ...QueueAttributeSetter) (string, error) {
	var err error

	if !checkQueueName(name) {
		return "", errors.New("unqualified queue name")
	}

	attri := DefaultQueueAttri()
	for _, setter := range setters {
		err = setter(&attri)
		if err != nil {
			return "", err
		}
	}

	requestLine := fmt.Sprintf(mnsCreateQueue, name)
	req := c.ca.NewRequest().Put().WithPath(requestLine).WithXMLBody(&attri).WithTimeout(apiTimeout)

	resp, err := c.ca.Do(req)
	if err != nil {
		return "", err
	}

	switch resp.StatusCode() {
	case http.StatusCreated:
		return resp.Header().Get(location), nil
	default:
		var respErr RespErr
		if err := resp.DecodeFromXML(&respErr); err != nil {
			return "", nil
		}
		switch respErr.Code {
		case createQueueNoContentError.Error():
			return "", createQueueNoContentError
		case createQueueConflictError.Error():
			return "", createQueueConflictError
		case queueNumExceededLimitError.Error():
			return "", queueNumExceededLimitError
		default:
			return "", errors.New(respErr.Code)
		}
	}
}
