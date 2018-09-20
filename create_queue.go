package aliyun_mns

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

func checkQueueName(name string) bool {
	regex := regexp.MustCompile(`^[0-9a-zA-Z]{1}[0-9a-zA-Z-]{0,255}$`)
	return regex.MatchString(name)
}

func (c *Client) CreateQueue(name string, setters ...QueueAttributeSetter) (string, error) {
	if !checkQueueName(name) {
		return "", errors.New("unqualified queue name")
	}

	attri := DefaultQueueAttri()
	for _, setter := range setters {
		if err := setter(&attri); err != nil {
			return "", err
		}
	}

	body, err := xml.Marshal(&attri)
	if err != nil {
		return "", err
	}

	requestLine := fmt.Sprintf(mnsCreateQueue, name)
	req, err := http.NewRequest(http.MethodPut, c.endpoint+requestLine, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	c.finalizeHeader(req, body)

	globalLogger.printf("创建队列请求: %s %s %s", req.Method, req.URL.String(), string(body))

	ctx, cancel := context.WithCancel(context.TODO())
	_ = time.AfterFunc(time.Second*timeout, func() {
		cancel()
	})
	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	globalLogger.printf("创建队列回复: %s %s", resp.Status, string(body))

	switch resp.StatusCode {
	case http.StatusCreated:
		return resp.Header.Get(location), nil
	case http.StatusNoContent:
		return "", createQueueNoContentError
	case http.StatusConflict:
		return "", createQueueConflictError
	default:
		return "", unknownError
	}
}
