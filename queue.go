package alimns

import (
	"github.com/willf/bitset"
)

// OnReceiveFunc 消息处理函数模板
type OnReceiveFunc func(message *ReceiveMessage) error

// Queue 消息队列
type Queue struct {
	Name                  string
	Parallel              int
	QueueAttributeSetters []QueueAttributeSetter
	OnReceive             OnReceiveFunc
	IsScheduled           bool
	receiveMessageChan    chan *ReceiveMessage
	longPollQuit          chan struct{}
	consumeQuit           chan struct{}
	statusBits            *bitset.BitSet
}

// Stop 使消息队列拉取消息和消费消息停止
func (q *Queue) Stop() {
	if q.IsScheduled {
		q.IsScheduled = false
		close(q.longPollQuit)
		close(q.consumeQuit)
	}
}
