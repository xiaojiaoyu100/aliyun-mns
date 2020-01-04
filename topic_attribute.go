package alimns

import (
	"encoding/xml"
	"errors"
)

const (
	defaultMaximumMessageSize = 65536
	defaultLoggingEnabled     = true
)

// TopicAttribute 修改主题属性
type TopicAttribute struct {
	XMLName            xml.Name `xml:"Topic"`
	MaximumMessageSize int      `xml:"MaximumMessageSize"`
	LoggingEnabled     bool     `xml:"LoggingEnabled"`
}

// TopicAttributeSetter 主题属性设置函数模板
type TopicAttributeSetter func(attri *TopicAttribute) error

// DefaultTopicAttr 返回默认的主题属性
func DefaultTopicAttr() TopicAttribute {
	return TopicAttribute{
		MaximumMessageSize: defaultMaximumMessageSize,
		LoggingEnabled:     defaultLoggingEnabled,
	}
}

// TopicWithMaximumMessageSize 设置消息体长度
func TopicWithMaximumMessageSize(size int) TopicAttributeSetter {
	return func(attri *TopicAttribute) error {
		if size < minMessageSize || size > maxMessageSize {
			return errors.New("maximum message size out of range")
		}
		attri.MaximumMessageSize = size
		return nil
	}
}

// TopicWithLoggingEnabled 设置日志开启
func TopicWithLoggingEnabled(flag bool) TopicAttributeSetter {
	return func(attri *TopicAttribute) error {
		attri.LoggingEnabled = flag
		return nil
	}
}
