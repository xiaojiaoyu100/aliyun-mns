package alimns

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// ChangeVisibilityTimeoutResponse 修改消息可见时长回复
type ChangeVisibilityTimeoutResponse struct {
	XMLName         xml.Name `xml:"ChangeVisibility"`
	XMLNs           string   `xml:"xmlns,attr"`
	ReceiptHandle   string   `xml:"ReceiptHandle"`
	NextVisibleTime int64    `xml:"NextVisibleTime"`
}

// ChangeVisibilityTimeout 修改消息可见时长
func (c *Client) ChangeVisibilityTimeout(
	name,
	receiptHandle string,
	visibilityTimeout int) (*ChangeVisibilityTimeoutResponse, error) {

	var err error
	if visibilityTimeout < minVisibilityTimeout || visibilityTimeout > maxVisibilityTimeout {
		return nil, visibilityTimeoutError
	}

	requestLine := fmt.Sprintf(mnsChangeMessageVisibility, name, receiptHandle, strconv.Itoa(visibilityTimeout))
	req := c.ca.NewRequest().Put().WithPath(requestLine).WithTimeout(apiTimeout)

	resp, err := c.ca.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		var changeVisibilityTimeoutResponse ChangeVisibilityTimeoutResponse
		if err := resp.DecodeFromXML(&changeVisibilityTimeoutResponse); err != nil {
			return nil, err
		}
		return &changeVisibilityTimeoutResponse, nil
	default:
		var respErr RespErr
		if err := resp.DecodeFromXML(&respErr); err != nil {
			return nil, err
		}
		switch respErr.Code {
		case messageNotExistError.Error():
			return nil, messageNotExistError
		case queueNotExistError.Error():
			return nil, queueNotExistError
		}
		return nil, errors.New(respErr.Message)
	}
}
