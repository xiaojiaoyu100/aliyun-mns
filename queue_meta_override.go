package alimns

import (
	"errors"
	"fmt"
	"net/http"
)

// QueueMetaOverride 修改队列属性
func (c *Client) QueueMetaOverride(name string, setters ...QueueAttributeSetter) error {
	var err error

	attri := ModifiedAttribute{}
	for _, setter := range setters {
		err = setter(&attri)
		if err != nil {
			return err
		}
	}

	requestLine := fmt.Sprintf(mnsQueueMetaOverride, name)
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
			return err
		}
		return errors.New(respErr.Code)
	}
}
