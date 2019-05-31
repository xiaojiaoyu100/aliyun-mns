package alimns

import (
	"github.com/willf/bitset"
)

// M 消息内容，去掉其它字段是为了不要依赖消息其它字段，应该依赖数据库字段
type M struct {
	MessageBody string
	EnqueueTime int64
}

// Handler 消息处理函数模板
type Handler func(m *M) error

// Queue 消息队列
type Queue struct {
	Name                  string
	Parallel              int
	QueueAttributeSetters []QueueAttributeSetter
	OnReceive             Handler
	Backoff               BackoffFunc
	isScheduled           bool
	receiveMessageChan    chan *ReceiveMessage
	longPollQuit          chan struct{}
	consumeQuit           chan struct{}
	statusBits            *bitset.BitSet
}

// Stop 使消息队列拉取消息和消费消息停止
func (q *Queue) Stop() {
	if q.isScheduled {
		q.isScheduled = false
		close(q.longPollQuit)
		close(q.consumeQuit)
	}
}
