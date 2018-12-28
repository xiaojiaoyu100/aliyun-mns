package alimns

import (
	"context"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

// ListQueueRequest 获取队列列表参数
type ListQueueRequest struct {
	Marker    string
	RetNumber string
	Prefix    string
}

// QueueData 队列的属性
type QueueData struct {
	QueueURL string `xml:"QueueURL"`
}

// ListQueueResponse 获取队列的回复
type ListQueueResponse struct {
	XMLName    xml.Name     `xml:"Queues"`
	XMLNs      string       `xml:"xmlns,attr"`
	Queues     []*QueueData `xml:"Queue"`
	NextMarker string       `xml:"NextMarker"`
}

// ListQueue 请求队列列表
func (c *Client) ListQueue(request *ListQueueRequest) (*ListQueueResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.endpoint+mnsListQueue, nil)
	if err != nil {
		return nil, err
	}

	if request.Marker != "" {
		req.Header.Set(xMnsMarker, request.Marker)
	}
	if request.RetNumber != "" {
		req.Header.Set(xMnsRetNumber, request.RetNumber)
	}
	if request.Prefix != "" {
		req.Header.Set(xMnxPrefix, request.Prefix)
	}

	c.finalizeHeader(req, nil)

	contextLogger.
		WithField("method", req.Method).
		WithField("url", req.URL.String()).
		Info("获取队列列表请求")

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

	contextLogger.
		WithField("status", resp.Status).
		WithField("body", string(body)).
		WithField("url", req.URL.String()).
		Info("获取队列列表回复")

	switch resp.StatusCode {
	case http.StatusOK:
		response := new(ListQueueResponse)
		if err := xml.Unmarshal(body, &response); err != nil {
			return nil, err
		}
		return response, nil
	default:
		var respErr RespErr
		if err := xml.Unmarshal(body, &respErr); err != nil {
			return nil, err
		}
		return nil, errors.New(respErr.Code)
	}
}
