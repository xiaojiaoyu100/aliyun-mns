package alimns

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// DeleteMessage 删除消息
func (c *Client) DeleteMessage(name, receiptHandle string) error {
	requestLine := fmt.Sprintf(mnsDeleteMessage, name, receiptHandle)
	req, err := http.NewRequest(http.MethodDelete, c.endpoint+requestLine, nil)
	if err != nil {
		return err
	}
	c.finalizeHeader(req, nil)

	contextLogger.
		WithField("method", req.Method).
		WithField("url", req.URL.String()).
		Info("删除消息请求")

	ctx, cancel := context.WithCancel(context.TODO())
	_ = time.AfterFunc(time.Second*timeout, func() {
		cancel()
	})
	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	contextLogger.
		WithField("status", resp.Status).
		WithField("body", string(body)).
		WithField("url", req.URL.String()).
		Info("删除消息回复")

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	default:
		var respErr RespErr
		if err := xml.Unmarshal(body, &respErr); err != nil {
			return err
		}
		return errors.New(respErr.Code)
	}
}
