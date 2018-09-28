package aliyun_mns

import (
	"encoding/xml"
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

// aliyun mns error code
const (
	messageNotExistError = MnsError("MessageNotExist")
	queueNotExistError   = MnsError("QueueNotExist")
	internalError        = MnsError("InternalError")
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
