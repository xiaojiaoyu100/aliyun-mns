package alimns

import (
	"encoding/xml"
	"errors"
)

// PublishMessageParam 发布消息的参数
type PublishMessageParam struct {
	XMLName           xml.Name `xml:"Message"`
	MessageBody       string   `xml:"MessageBody"`
	MessageTag        string   `xml:"MessageTag,omitempty"`
	MessageAttributes string   `xml:"MessageAttributes,omitempty"`
}

// PublishMessageParamSetter 发布消息的参数设置函数模板
type PublishMessageParamSetter func(attri *PublishMessageParam) error

// WithMessageTag 设置消息标签（用于消息过滤）
func WithMessageTag(filterTag string) PublishMessageParamSetter {
	return func(attri *PublishMessageParam) error {
		if filterTag == "" && len(filterTag) > 16 {
			return errors.New("参数限制为不超过16个字符的字符串")
		}
		attri.MessageTag = filterTag
		return nil
	}
}
