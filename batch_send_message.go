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

// SendMessage 发送消息
type SendMessage struct {
	XMLName        xml.Name `xml:"Message"`
	XMLNs          string   `xml:"xmlns,attr"`
	ErrorCode      string   `xml:"ErrorCode"`
	ErrorMessage   string   `xml:"ErrorMessage"`
	MessageID      string   `xml:"MessageId"`
	MessageBodyMD5 string   `xml:"MessageBodyMD5"`
	ReceiptHandle  string   `xml:"ReceiptHandle"` // 发送延时消息才有返回
}

// BatchSendMessageResponse 批量发送消息回复
type BatchSendMessageResponse struct {
	XMLName      xml.Name       `xml:"Messages"`
	XMLNs        string         `xml:"xmlns,attr"`
	SendMessages []*SendMessage `xml:"Message"`
}

// BatchSendMessage 批量发送消息
func (c *Client) BatchSendMessage(name string, messageList ...*Message) (*BatchSendMessageResponse, error) {
	if len(messageList) > 16 {
		return nil, batchSendMessageNumLimitError
	}

	var (
		try            = 0
		sendedMessages = make([]*SendMessage, len(messageList))
		roundIndexList []int
	)

	originalMessageList := messageList

	for idx := range originalMessageList {
		roundIndexList = append(roundIndexList, idx)
	}

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
	try++
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	globalLogger.printf("批量发送消息回复: %s %s", resp.Status, string(body))

	switch resp.StatusCode {
	case http.StatusCreated,
		http.StatusInternalServerError:
		var batchSendMessageResponse BatchSendMessageResponse
		if err := xml.Unmarshal(body, &batchSendMessageResponse); err != nil {
			return nil, err
		}
		var retryIdx []int
		for seq, sendMessage := range batchSendMessageResponse.SendMessages {
			idx := roundIndexList[seq]

			switch sendMessage.ErrorCode {
			case "":
				sendedMessages[idx] = sendMessage
			case internalError.Error():
				retryIdx = append(retryIdx, idx)
			default:
				notifyAsync("批量发送消息部分失败: ", batchSendMessageResponse, err)
			}
		}
		if len(retryIdx) == 0 {
			batchSendMessageResponse.SendMessages = sendedMessages
			return &batchSendMessageResponse, nil
		}

		roundIndexList = retryIdx

		var retryMessageList []*Message
		for _, idx := range retryIdx {
			retryMessageList = append(retryMessageList, originalMessageList[idx])
		}

		messageList = retryMessageList

		if try > 4 {
			return nil, batchSendMessageTryLimitError
		}

		time.Sleep(100 * time.Millisecond * time.Duration(try))
		goto start

	default:
		var respErr RespErr
		if err := xml.Unmarshal(body, &respErr); err != nil {
			return nil, err
		}
		return nil, errors.New(respErr.Message)
	}
}
