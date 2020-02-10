package alimns

import (
	"context"
	"github.com/xiaojiaoyu100/curlew"
)

// M 消息内容，去掉其它字段是为了不要依赖消息其它字段，应该依赖数据库字段
type M struct {
	QueueName   string // 队列名
	MessageBody string // 消息体
	EnqueueTime int64  // 入队时间
	codec       Codec
}

// Decode 解析消息
func (m *M) Decode(v interface{}) error {
	return m.codec.Decode([]byte(m.MessageBody), &v)
}

// Handle 消息处理函数模板
type Handle func(ctx context.Context, m *M) error

// MakeContext 生成一个context
type MakeContext func(m *M) context.Context

// Queue 消息队列
type Queue struct {
	Name                  string
	Parallel              int
	QueueAttributeSetters []QueueAttributeSetter
	OnReceive             Handle
	Backoff               BackoffFunc
	MakeContext           MakeContext
	Codec                 Codec
	isScheduled           bool
	receiveMessageChan    chan *ReceiveMessage
	longPollQuit          chan struct{}
	consumeQuit           chan struct{}
	dispatcher            *curlew.Dispatcher
	popCount              int32
}

// Stop 使消息队列拉取消息和消费消息停止
func (q *Queue) Stop() {
	if q.isScheduled {
		q.isScheduled = false
		close(q.longPollQuit)
		close(q.consumeQuit)
	}
}

// MakeDefaultCodecIfNone 保证codec不为空，默认是json编解码
func (q *Queue) MakeDefaultCodecIfNone() Codec {
	if q.Codec == nil {
		q.Codec = JSONCodec{}
	}
	return q.Codec
}

// MakeDefaultContextIfNone 保证context不为空
func (q *Queue) MakeDefaultContextIfNone(m *M) context.Context {
	ctx := q.MakeContext(m)
	if ctx == nil {
		return context.TODO()
	}
	return ctx
}
