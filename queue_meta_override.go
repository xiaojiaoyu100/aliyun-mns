package alimns

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// QueueMetaOverride 修改队列属性
func (c *Client) QueueMetaOverride(name string, setters ...QueueAttributeSetter) error {
	attri := ModifiedAttribute{}
	for _, setter := range setters {
		if err := setter(&attri); err != nil {
			return err
		}
	}

	body, err := xml.Marshal(&attri)
	if err != nil {
		return err
	}

	requestLine := fmt.Sprintf(mnsQueueMetaOverride, name)
	req, err := http.NewRequest(http.MethodPut, c.endpoint+requestLine, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	c.finalizeHeader(req, body)

	contextLogger.
		WithField("method", req.Method).
		WithField("url", req.URL.String()).
		WithField("body", string(body)).
		Info("设置队列属性请求")

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

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	contextLogger.
		WithField("status", resp.Status).
		WithField("body", string(body)).
		WithField("url", req.URL.String()).
		Info("设置队列属性回复")

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
