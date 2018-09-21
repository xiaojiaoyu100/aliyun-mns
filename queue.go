package aliyun_mns

type OnReceiveFunc func(message *ReceiveMessage, err chan error)

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
		q.longPollQuit <- struct{}{}
		for i := 1; i <= q.Parallel; i++ {
			q.consumeQuit <- struct{}{}
		}
	}
}
