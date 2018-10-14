package aliyun_mns

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type SendMessage struct {
	XMLName        xml.Name `xml:"Message"`
	XmlNs          string   `xml:"xmlns,attr"`
	ErrorCode      string   `xml:"ErrorCode"`
	ErrorMessage   string   `xml:"ErrorMessage"`
	MessageId      string   `xml:"MessageId"`
	MessageBodyMD5 string   `xml:"MessageBodyMD5"`
	ReceiptHandle  string   `xml:"ReceiptHandle"` // 发送延时消息才有返回
}

type BatchSendMessageResponse struct {
	XMLName      xml.Name       `xml:"Messages"`
	XmlNs        string         `xml:"xmlns,attr"`
	SendMessages []*SendMessage `xml:"Message"`
}

func (c *Client) BatchSendMessage(name string, messageList ...*Message) (*BatchSendMessageResponse, error) {
start:
	body, err := xml.Marshal(&messageList)
	if err != nil {
		return nil, err
	}
	requestLine := fmt.Sprintf(mnsBatchSendMessage, name)
	req, err := http.NewRequest(http.MethodPost, c.endpoint+requestLine, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	c.finalizeHeader(req, body)

	globalLogger.printf("批量发送消息请求: %s %s %s", req.Method, req.URL.String(), string(body))

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

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	globalLogger.printf("批量发送消息回复: %s %s", resp.Status, string(body))

	switch resp.StatusCode {
	case http.StatusCreated:
		var batchSendMessageResponse BatchSendMessageResponse
		if err := xml.Unmarshal(body, &batchSendMessageResponse); err != nil {
			return nil, err
		}
		return &batchSendMessageResponse, nil
	case http.StatusInternalServerError:
		var batchSendMessageResponse BatchSendMessageResponse
		if err := xml.Unmarshal(body, &batchSendMessageResponse); err != nil {
			return nil, err
		}
		var retryIdx []int
		for idx, sendMessage := range batchSendMessageResponse.SendMessages {
			if sendMessage.ErrorCode != internalError.Error() {
				continue
			}
			retryIdx = append(retryIdx, idx)
		}
		var retryMessageList []*Message
		for _, idx := range retryIdx {
			retryMessageList = append(retryMessageList, messageList[idx])
		}

		if len(retryMessageList) > 0 {
			return &batchSendMessageResponse, nil
		}

		messageList = retryMessageList
		goto start
	default:
		return nil, unknownError
	}
}
