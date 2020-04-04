package alimns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// Unsubscribe 取消订阅主题
func (c *Client) Unsubscribe(topic, subscription string) error {
	var err error

	if !checkTopicName(topic) {
		return errors.New("unqualified topic name")
	}

	if !checkSubscription(subscription) {
		return errors.New("unqualified subscription nname")
	}

	requestLine := fmt.Sprintf(mnsUnsubscribe, topic, subscription)
	req := c.ca.NewRequest().Delete().WithPath(requestLine).WithTimeout(apiTimeout)

	resp, err := c.ca.Do(context.TODO(), req)
	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil
	default:
		return errors.New(resp.String())
	}
}
