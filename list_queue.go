package alimns

import (
	"context"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
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

// BatchListQueue 批量请求队列
func (c *Client) BatchListQueue() error {
	request := new(ListQueueRequest)
	request.RetNumber = "1000"
	resp, err := c.ListQueue(request)
	if err != nil {
		return err
	}

	c.doneQueues = make(map[string]struct{})

	for _, queue := range resp.Queues {
		idx := strings.LastIndex(queue.QueueURL, "/")

		name := queue.QueueURL[idx+1:]

		if _, ok := c.doneQueues[name]; !ok {
			c.doneQueues[name] = struct{}{}
		}
	}

	for {
		if resp.NextMarker == "" {
			return nil
		}

		resp, err = c.ListQueue(&ListQueueRequest{
			Marker: resp.NextMarker,
		})

		if err != nil {
			return err
		}

		for _, queue := range resp.Queues {
			idx := strings.LastIndex(queue.QueueURL, "/")

			name := queue.QueueURL[idx+1:]

			if _, ok := c.doneQueues[name]; !ok {
				c.doneQueues[name] = struct{}{}
			}
		}

		time.Sleep(1 * time.Second)
	}
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

	globalLogger.printf("获取队列列表请求: %s %s", req.Method, req.URL.String())

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

	globalLogger.printf("获取队列列表回复: %s %s", resp.Status, string(body))

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
