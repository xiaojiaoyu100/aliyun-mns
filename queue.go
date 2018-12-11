package aliyun_mns

type OnReceiveFunc func(message *ReceiveMessage, queneName string) error

type Queue struct {
	Name                  string
	Parallel              int
	QueueAttributeSetters []QueueAttributeSetter
	OnReceive             OnReceiveFunc
	isRunning             bool
	receiveMessageChan    chan *ReceiveMessage
	longPollQuit          chan struct{}
	consumeQuit           chan struct{}
}

func (q *Queue) Stop() {
	if q.isRunning {
		q.isRunning = false
		close(q.longPollQuit)
		close(q.consumeQuit)
	}
}
