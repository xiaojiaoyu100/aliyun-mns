package alimns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// DeleteQueue 删除队列
func (c *Client) DeleteQueue(name string) error {
	var err error

	requestLine := fmt.Sprintf(mnsDeleteQueue, name)
	req := c.ca.NewRequest().Delete().WithPath(requestLine).WithTimeout(apiTimeout)

	resp, err := c.ca.Do(context.TODO(), req)
	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil
	default:
		var respErr RespErr
		if err := resp.DecodeFromXML(&respErr); err != nil {
			return err
		}
		return errors.New(respErr.Code)
	}
}
