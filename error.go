package alimns

import (
	"encoding/xml"

	"github.com/xiaojiaoyu100/cast"
)

// MnsError 错误码
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
	batchSendMessageTryLimitError      = MnsError("BatchSendMessageTryLimit")
	batchSendMessageNumLimitError      = MnsError("BatchSendMessageNumLimit")
	handleCrashError                   = MnsError("handleCrash")
	queueNumExceededLimitError         = MnsError("QueueNumExceededLimit")
	subscriptionNameLengthError        = MnsError("SubscriptionNameLengthError")
	subscriptionNameInvalidError       = MnsError("SubscriptionNameInvalid")
	subscriptionAlreadyExistError      = MnsError("SubscriptionAlreadyExist")
	endpointInvalidError               = MnsError("EndpointInvalid")
	invalidArgumentError               = MnsError("InvalidArgument")
	topicAlreadyExistError             = MnsError("TopicAlreadyExist")
	topicNameLengthError               = MnsError("TopicNameLengthError")
	topicNotExistError                 = MnsError("TopicNotExist")
)

// aliyun mns error code
const (
	messageNotExistError = MnsError("MessageNotExist")
	queueNotExistError   = MnsError("QueueNotExist")
	internalError        = MnsError("InternalError")
)

// IsUnknown 是否是未知错误
func IsUnknown(err error) bool {
	return err == unknownError
}

// IsVisibilityTimeout 是否是可见时间不在范围
func IsVisibilityTimeout(err error) bool {
	return err == visibilityTimeoutError
}

// IsCreateQueueNoContent 是否是createQueueNoContentError
func IsCreateQueueNoContent(err error) bool {
	return err == createQueueNoContentError
}

// IsCreateQueueConflict 是否是createQueueConflictError
func IsCreateQueueConflict(err error) bool {
	return err == createQueueConflictError
}

// IsMessageBodyLimit 是否超出范围
func IsMessageBodyLimit(err error) bool {
	return err == messageBodyLimitError
}

// IsSendMessageTimeout 是否发送消息超时
func IsSendMessageTimeout(err error) bool {
	return err == sendMessageTimeoutError
}

// IsMessageDelaySecondsOutOfRange 延时时长是否合理
func IsMessageDelaySecondsOutOfRange(err error) bool {
	return err == messageDelaySecondsOutOfRangeError
}

// IsHandleCrash 是否是处理函数崩溃错误
func IsHandleCrash(err error) bool {
	return err == handleCrashError
}

// IsInternalError 是否内部错误
func IsInternalError(err error) bool {
	return err == internalError
}

// RespErr 阿里云回复错误
type RespErr struct {
	XMLName   xml.Name `xml:"Error"`
	XMLNs     string   `xml:"xmlns,attr"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	RequestID string   `xml:"RequestId"`
	HostID    string   `xml:"HostId"`
}

func shouldRetry(err error) bool {
	return IsInternalError(err) || cast.ShouldRetry(err)
}

type transientError interface {
	error
	// Transient returns whether the error is transient, if returns true, the handler error will not be reported in a
	// short period of time.
	Transient() bool
}

// TransientError implements the transientError interface.
type TransientError struct {
	Err error
}

// Transient means the error is transient.
func (err TransientError) Transient() bool {
	return true
}
