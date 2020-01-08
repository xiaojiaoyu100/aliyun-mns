package alimns

import (
	"encoding/xml"
	"errors"
)

var defaultSubscribeParam = SubscribeParam{
	NotifyStrategy:      "BACKOFF_RETRY",
	NotifyContentFormat: "XML",
}

// SubscribeParam 订阅主题需要的参数
type SubscribeParam struct {
	XMLName             xml.Name `xml:"Subscription"`
	EndPoint            string   `xml:"Endpoint"`
	FilterTag           string   `xml:"FilterTag,omitempty"`
	NotifyStrategy      string   `xml:"NotifyStrategy"`
	NotifyContentFormat string   `xml:"NotifyContentFormat"`
}

// SubscribeParamSetter 订阅主题消息属性设置函数模板
type SubscribeParamSetter func(attri *SubscribeParam) error

// WithFilterTag 设置过滤标签
func WithFilterTag(filterTag string) SubscribeParamSetter {
	return func(attri *SubscribeParam) error {
		if filterTag == "" && len(filterTag) > 16 {
			return errors.New("参数限制为不超过16个字符的字符串")
		}
		attri.FilterTag = filterTag
		return nil
	}
}

// WithNotifyStrategy 设置推送消息出现错误时的重试策略
func WithNotifyStrategy(ns string) SubscribeParamSetter {
	return func(attri *SubscribeParam) error {
		if ns != "BACKOFF_RETRY" && ns != "EXPONENTIAL_DECAY_RETRY" {
			return errors.New("仅支持BACKOFF_RETRY 或者 EXPONENTIAL_DECAY_RETRY")
		}
		attri.NotifyStrategy = ns
		return nil
	}
}

// WithNotifyContentFormat 设置最长存活时间
func WithNotifyContentFormat(s string) SubscribeParamSetter {
	return func(attri *SubscribeParam) error {
		if s != "XML" && s != "JSON" && s != "SIMPLIFIED" {
			return errors.New("仅支持XML 、JSON 或者 SIMPLIFIED")
		}
		attri.NotifyContentFormat = s
		return nil
	}
}
