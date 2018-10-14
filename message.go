package aliyun_mns

import (
	"encoding/xml"
	"errors"
)

type Message struct {
	XMLName      xml.Name `xml:"Message"`
	XmlNs        string   `xml:"xmlns,attr"`
	MessageBody  string   `xml:"MessageBody,omitempty"` // 消息正文
	DelaySeconds int      `xml:"DelaySeconds,omitempty"`
	Priority     int      `xml:"Priority,omitempty"`
}

type MessageSetter func(msg *Message) error

func DefaultMessage() Message {
	return Message{}
}

func WithMessageDelaySeconds(delay int) MessageSetter {
	return func(msg *Message) error {
		if delay < minDelaySeconds || delay > maxDelaySeconds {
			return messageDelaySecondsOutOfRangeError
		}
		msg.DelaySeconds = delay
		return nil
	}
}

func WithMessagePriority(priority int) MessageSetter {
	return func(msg *Message) error {
		if priority < 1 || priority > 16 {
			return errors.New("message priority out of range")
		}
		msg.Priority = priority
		return nil
	}
}
