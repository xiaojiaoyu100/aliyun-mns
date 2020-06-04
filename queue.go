package alimns

import (
	"context"
	"runtime"

	"github.com/xiaojiaoyu100/curlew"
)

// M 消息内容，去掉其它字段是为了不要依赖消息其它字段，应该依赖数据库字段
type M struct {
	QueueName     string // 队列名
	MessageBody   string // 消息体
	EnqueueTime   int64  // 入队时间
	ReceiptHandle string // 消息句柄
	codec         Codec
}

// Decode 解析消息
func (m *M) Decode(v interface{}) error {
	return m.codec.Decode([]byte(m.MessageBody), &v)
}

// Handle 消息处理函数模板
type Handle func(ctx context.Context) error

// Queue 消息队列
type Queue struct {
	Name             string
	Parallel         int
	AttributeSetters []QueueAttributeSetter
	Builder
	PullWait           bool // 等消息消費完再去拉取消息
	codec              Codec
	before             Before
	after              After
	isScheduled        bool
	receiveMessageChan chan *ReceiveMessage
	longPollQuit       chan struct{}
	consumeQuit        chan struct{}
	dispatcher         *curlew.Dispatcher
	popCount           int32
}

// Stop 使消息队列拉取消息和消费消息停止
func (q *Queue) Stop() {
	if q.isScheduled {
		q.isScheduled = false
		close(q.longPollQuit)
		close(q.consumeQuit)
	}
}

func (q *Queue) safeParallel() int {
	if q.Parallel > maxReceiveMessage {
		return maxReceiveMessage
	}
	if q.Parallel >= 1 && q.Parallel <= maxReceiveMessage {
		return q.Parallel
	}
	n := runtime.NumCPU() * 2
	switch {
	case n > maxReceiveMessage:
		n = maxReceiveMessage
	case n <= 0:
		n = 1
	}
	return n
}

func (q *Queue) busy() bool {
	return q.safeParallel() <= int(q.popCount)
}

func (q *Queue) safePullNumOfMessages() int {
	r := q.safeParallel() - int(q.popCount)
	if r < 1 {
		return 1
	}
	return r
}
