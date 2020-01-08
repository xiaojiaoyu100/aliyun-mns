package alimns

import (
	"errors"
	"fmt"
	"net/http"
)

// Unsubscribe 取消订阅主题
func (c *Client) Unsubscribe(topic, subscriptions string) error {
	var err error

	if !checkTopicName(topic) {
		return errors.New("unqualified topic name")
	}

	if !checkSubscription(subscriptions) {
		return errors.New("unqualified Subscriptio nname")
	}

	requestLine := fmt.Sprintf(mnsUnsubscribe, topic, subscriptions)
	req := c.ca.NewRequest().Delete().WithPath(requestLine).WithTimeout(apiTimeout)

	resp, err := c.ca.Do(req)
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
