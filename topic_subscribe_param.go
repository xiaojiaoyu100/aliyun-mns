package alimns

import (
	"encoding/xml"
	"errors"
	"fmt"
)

const (
	// BackOffRetryStrategy 重试3次，每次重试的间隔时间是10秒到 20秒之间的随机值
	BackOffRetryStrategy = "BACKOFF_RETRY"
	// ExponentialDecayRetryStrategy 重试176次，每次重试的间隔时间指数递增至512秒，总计重试时间为1天；每次重试的具体间隔为：1，2，4，8，16，32，64，128，256，512，512 ... 512 秒（共167个512）
	ExponentialDecayRetryStrategy = "EXPONENTIAL_DECAY_RETRY"
	// XMLNotifyFormat 消息体为XML格式，包含消息正文和消息属性
	XMLNotifyFormat = "XML"
	// JSONNotifyFormat 消息体为JSON格式，包含消息正文和消息属性
	JSONNotifyFormat = "JSON"
	// SimplifiedNotifyFormat 消息体即用户发布的消息，不包含任何属性信息
	SimplifiedNotifyFormat = "SIMPLIFIED"
)

var defaultSubscribeParam = SubscribeParam{
	NotifyStrategy:      BackOffRetryStrategy,
	NotifyContentFormat: XMLNotifyFormat,
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
		if ns != BackOffRetryStrategy && ns != ExponentialDecayRetryStrategy {
			return fmt.Errorf("notifyStrategy参数错误，仅支持%s或者%s",
				BackOffRetryStrategy, ExponentialDecayRetryStrategy)
		}
		attri.NotifyStrategy = ns
		return nil
	}
}

// WithNotifyContentFormat 设置最长存活时间
func WithNotifyContentFormat(s string) SubscribeParamSetter {
	return func(attri *SubscribeParam) error {
		if s != XMLNotifyFormat && s != JSONNotifyFormat && s != SimplifiedNotifyFormat {
			return fmt.Errorf("notifyContentFormat参数错误，仅支持%s、%s、%s",
				XMLNotifyFormat, JSONNotifyFormat, SimplifiedNotifyFormat)
		}
		attri.NotifyContentFormat = s
		return nil
	}
}
