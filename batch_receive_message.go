package aliyun_mns

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-querystring/query"
)

type BatchReceiveMessageResponse struct {
	XMLName         xml.Name          `xml:"Messages"`
	XmlNs           string            `xml:"xmlns,attr"`
	ReceiveMessages []*ReceiveMessage `xml:"Message"`
}

func DefaultBatchReceiveMessage() ReceiveMessageParam {
	return ReceiveMessageParam{
		NumOfMessages: defaultReceiveMessage,
	}
}

func (c *Client) BatchReceiveMessage(name string, setters ...ReceiveMessageParamSetter) (*BatchReceiveMessageResponse, error) {
	receiveMessage := DefaultBatchReceiveMessage()
	for _, setter := range setters {
		if err := setter(&receiveMessage); err != nil {
			return nil, err
		}
	}

	requestLine := fmt.Sprintf(mnsBatchReceiveMessage, name)
	req, err := http.NewRequest(http.MethodGet, c.endpoint+requestLine, nil)
	if err != nil {
		return nil, err
	}
	values, err := query.Values(&receiveMessage)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = values.Encode()

	c.finalizeHeader(req, nil)

	globalLogger.printf("批量消费消息请求: %s %s", req.Method, req.URL.String())

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	globalLogger.printf("批量消费消息回复: %s %s", resp.Status, string(body))

	switch resp.StatusCode {
	case http.StatusOK:
		var batchReceiveMessageResponse BatchReceiveMessageResponse
		if err := xml.Unmarshal(body, &batchReceiveMessageResponse); err != nil {
			return nil, err
		}
		return &batchReceiveMessageResponse, nil
	default:
		var respErr RespErr
		if err := xml.Unmarshal(body, &respErr); err != nil {
			return nil, err
		}

		switch respErr.Code {
		case messageNotExistError.Error():
			return nil, messageNotExistError
		case queueNotExistError.Error():
			return nil, queueNotExistError
		}

		return nil, unknownError
	}
}
