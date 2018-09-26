package aliyun_mns

import (
	"encoding/xml"
	"io"
)

type MnsError string

func (e MnsError) Error() string {
	return string(e)
}

// customize code
const (
	unknownError                       = MnsError("Unknown")
	visibilityTimeoutError             = MnsError("VisibilityTimeoutOutOfRange")
	createQueueNoContentError          = MnsError("CreateQueueNoContent")
	createQueueConflictError           = MnsError("CreateQueueConflict")
	messageBodyLimitError              = MnsError("MessageBodyLimit")
	sendMessageTimeoutError            = MnsError("SendMessageTimeout")
	messageDelaySecondsOutOfRangeError = MnsError("MessageDelaySecondsOutOfRange")
)

// alimns error code
const (
	messageNotExistError = MnsError("MessageNotExist")
	queueNotExistError   = MnsError("QueueNotExist")
)

func IsUnknown(err error) bool {
	return err == unknownError
}

func IsVisibilityTimeout(err error) bool {
	return err == visibilityTimeoutError
}

func IsCreateQueueNoContent(err error) bool {
	return err == createQueueNoContentError
}

func IsCreateQueueConflict(err error) bool {
	return err == createQueueConflictError
}

func IsMessageBodyLimit(err error) bool {
	return err == messageBodyLimitError
}

func IsSendMessageTimeout(err error) bool {
	return err == sendMessageTimeoutError
}

func IsMessageDelaySecondsOutOfRange(err error) bool {
	return err == messageDelaySecondsOutOfRangeError
}

type RespErr struct {
	XMLName   xml.Name `xml:"Error"`
	XmlNs     string   `xml:"xmlns,attr"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	RequestId string   `xml:"RequestId"`
	HostId    string   `xml:"HostId"`
}

func shouldRetry(err error) bool {
	// tcp读
	if err == io.EOF {
		return true
	}
	// 关于timeout的说明，timeout并不一定代表请求没成功
	// 但是在消息队列这种场景下好像也是无害的．
	switch err := err.(type) {
	case interface {
		Temporary() bool
	}:
		return err.Temporary()
	case interface {
		Timeout() bool
	}:
		return err.Timeout()
	default:
		return false
	}
}
