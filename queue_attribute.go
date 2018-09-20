package aliyun_mns

import (
	"encoding/xml"
	"errors"
)

const (
	defaultDelaySeconds = 0
	minDelaySeconds     = 0
	maxDelaySeconds     = 604800 // 7天
)

const (
	defaultMessageSize = 65536
	minMessageSize     = 1024
	maxMessageSize     = 65536
)

const (
	defaultRetentionPeriod = 1296000 // 15天
	minRetentionPeriod     = 60      // 60秒
	maxRetentionPeriod     = 1296000 // 15天
)

const (
	defaultVisibilityTimeout = 30
	minVisibilityTimeout     = 1
	maxVisibilityTimeout     = 43200
)

const (
	defaultPollingWaitSeconds = 30
	minPollingWaitSeconds     = 0
	maxPollingWaitSeconds     = 30
)

type ModifiedAttribute struct {
	XMLName                xml.Name `xml:"Queue"`
	DelaySeconds           *int     `xml:"DelaySeconds,omitempty"`
	MaximumMessageSize     int      `xml:"MaximumMessageSize,omitempty"`
	MessageRetentionPeriod int      `xml:"MessageRetentionPeriod,omitempty"`
	VisibilityTimeout      int      `xml:"VisibilityTimeout,omitempty"`
	PollingWaitSeconds     *int     `xml:"PollingWaitSeconds,omitempty"`
	LoggingEnabled         *bool    `xml:"LoggingEnabled,omitempty"`
}

type QueueAttributeSetter func(attri *ModifiedAttribute) error

func DefaultQueueAttri() ModifiedAttribute {
	var (
		delaySeconds       = defaultDelaySeconds
		pollingWaitSeconds = defaultPollingWaitSeconds
		loggingEnabled     = true
	)
	return ModifiedAttribute{
		DelaySeconds:           &delaySeconds,
		MaximumMessageSize:     defaultMessageSize,
		MessageRetentionPeriod: defaultRetentionPeriod,
		VisibilityTimeout:      defaultVisibilityTimeout,
		PollingWaitSeconds:     &pollingWaitSeconds,
		LoggingEnabled:         &loggingEnabled,
	}
}

// 设置延时时间
func WithDelaySeconds(s int) QueueAttributeSetter {
	return func(attri *ModifiedAttribute) error {
		if s < minDelaySeconds || s > maxDelaySeconds {
			return errors.New("delay seconds out of range")
		}
		attri.DelaySeconds = &s
		return nil
	}
}

// 设置消息体长度
func WithMaximumMessageSize(size int) QueueAttributeSetter {
	return func(attri *ModifiedAttribute) error {
		if size < minMessageSize || size > maxMessageSize {
			return errors.New("maximum message size out of range")
		}
		attri.MaximumMessageSize = size
		return nil
	}
}

// 设置最长存活时间
func WithMessageRetentionPeriod(s int) QueueAttributeSetter {
	return func(attri *ModifiedAttribute) error {
		if s < minRetentionPeriod || s > maxRetentionPeriod {
			return errors.New("retention period out of range")
		}
		attri.MessageRetentionPeriod = s
		return nil
	}
}

// 设置可见时间
func WithVisibilityTimeout(s int) QueueAttributeSetter {
	return func(attri *ModifiedAttribute) error {
		if s < minVisibilityTimeout || s > maxVisibilityTimeout {
			return visibilityTimeoutError
		}
		attri.VisibilityTimeout = s
		return nil
	}
}

// 设置长轮询时间
func WithPollingWaitSeconds(s int) QueueAttributeSetter {
	return func(attri *ModifiedAttribute) error {
		if s < minPollingWaitSeconds || s > maxPollingWaitSeconds {
			return errors.New("polling wait seconds out of range")
		}
		attri.PollingWaitSeconds = &s
		return nil
	}
}

// 设置日志开启
func WithLoggingEnabled(flag bool) QueueAttributeSetter {
	return func(attri *ModifiedAttribute) error {
		attri.LoggingEnabled = &flag
		return nil
	}
}
