package alimns

import (
	"encoding/xml"
	"errors"
)

// Message 代表了阿里云消息
type Message struct {
	XMLName      xml.Name `xml:"Message"`
	XMLNs        string   `xml:"xmlns,attr"`
	MessageBody  string   `xml:"MessageBody,omitempty"` // 消息正文
	DelaySeconds int      `xml:"DelaySeconds,omitempty"`
	Priority     int      `xml:"Priority,omitempty"`
}

// MessageSetter 设置消息的函数
type MessageSetter func(msg *Message) error

// DefaultMessage 给出了默认的消息
func DefaultMessage() Message {
	return Message{}
}

// WithMessageDelaySeconds 设置消息延时
func WithMessageDelaySeconds(delay int) MessageSetter {
	return func(msg *Message) error {
		if delay < minDelaySeconds || delay > maxDelaySeconds {
			return messageDelaySecondsOutOfRangeError
		}
		msg.DelaySeconds = delay
		return nil
	}
}

// WithMessagePriority 设置消息优先级
func WithMessagePriority(priority int) MessageSetter {
	return func(msg *Message) error {
		if priority < 1 || priority > 16 {
			return errors.New("message priority out of range")
		}
		msg.Priority = priority
		return nil
	}
}
