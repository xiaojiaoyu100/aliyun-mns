package alimns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// DeleteTopic 删除主题
func (c *Client) DeleteTopic(name string) error {
	var err error

	requestLine := fmt.Sprintf(mnsDeleteTopic, name)
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
