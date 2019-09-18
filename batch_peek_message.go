package alimns

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
)

// BatchPeekMessageResponse 批量查看消息
type BatchPeekMessageResponse struct {
	XMLName      xml.Name       `xml:"Messages"`
	XMLNs        string         `xml:"xmlns,attr"`
	PeekMessages []*PeekMessage `xml:"Message"`
}

// BatchPeekMessage 批量查看消息
func (c *Client) BatchPeekMessage(name string) (*BatchPeekMessageResponse, error) {
	var err error

	requestLine := fmt.Sprintf(mnsBatchPeekMessage, name, "16")
	req := c.ca.NewRequest().Get().WithPath(requestLine).WithTimeout(apiTimeout)

	resp, err := c.ca.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var batchPeekMessageResponse BatchPeekMessageResponse
		if err := resp.DecodeFromXML(&batchPeekMessageResponse); err != nil {
			return nil, err
		}
		return &batchPeekMessageResponse, nil
	default:
		var respErr RespErr
		if err := resp.DecodeFromXML(&respErr); err != nil {
			return nil, err
		}
		if respErr.Code == queueNotExistError.Error() {
			return nil, queueNotExistError
		}
		return nil, errors.New(respErr.Code)
	}
}
