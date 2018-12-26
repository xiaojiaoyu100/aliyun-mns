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

// BatchPeekMessageResponse 批量查看消息
type BatchPeekMessageResponse struct {
	XMLName      xml.Name       `xml:"Messages"`
	XMLNs        string         `xml:"xmlns,attr"`
	PeekMessages []*PeekMessage `xml:"Message"`
}

// BatchPeekMessage 批量查看消息
func (c *Client) BatchPeekMessage(name string) (*BatchPeekMessageResponse, error) {
	requestLine := fmt.Sprintf(mnsBatchPeekMessage, name, "16")
	req, err := http.NewRequest(http.MethodGet, c.endpoint+requestLine, nil)
	if err != nil {
		return nil, err
	}
	c.finalizeHeader(req, nil)

	globalLogger.printf("批量查看消息请求: %s %s", req.Method, req.URL.String())

	ctx, cancel := context.WithCancel(context.TODO())
	_ = time.AfterFunc(time.Second*timeout, func() {
		cancel()
	})
	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	globalLogger.printf("批量查看消息回复: %s %s", resp.Status, string(body))

	switch resp.StatusCode {
	case http.StatusOK:
		var batchPeekMessageResponse BatchPeekMessageResponse
		if err := xml.Unmarshal(body, &batchPeekMessageResponse); err != nil {
			return nil, err
		}
		return &batchPeekMessageResponse, nil
	default:
		var respErr RespErr
		if err := xml.Unmarshal(body, &respErr); err != nil {
			return nil, err
		}
		switch respErr.Code {
		case queueNotExistError.Error():
			return nil, queueNotExistError
		}
		return nil, errors.New(respErr.Code)
	}
}
