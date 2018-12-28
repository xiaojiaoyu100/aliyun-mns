package alimns

import (
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/willf/bitset"
)

const (
	changeVisibilityInterval = 5
)

// Consumer 消费者
type Consumer struct {
	Client
	queues     []*Queue
	doneQueues map[string]struct{}
	shutdown   chan struct{}
	isClosed   bool
}

// NewConsumer 生成了一个消费者
func NewConsumer(client Client) *Consumer {
	consumer := new(Consumer)
	consumer.Client = client
	consumer.queues = make([]*Queue, 0)
	consumer.doneQueues = make(map[string]struct{})
	consumer.shutdown = make(chan struct{})
	return consumer
}

// BatchListQueue 批量请求队列
func (c *Consumer) BatchListQueue() error {
	request := new(ListQueueRequest)
	request.RetNumber = "1000"
	resp, err := c.ListQueue(request)
	if err != nil {
		return err
	}

	c.doneQueues = make(map[string]struct{})

	for _, queue := range resp.Queues {
		idx := strings.LastIndex(queue.QueueURL, "/")

		name := queue.QueueURL[idx+1:]

		if _, ok := c.doneQueues[name]; !ok {
			c.doneQueues[name] = struct{}{}
		}
	}

	for {
		if resp.NextMarker == "" {
			return nil
		}

		resp, err = c.ListQueue(&ListQueueRequest{
			Marker: resp.NextMarker,
		})

		if err != nil {
			return err
		}

		for _, queue := range resp.Queues {
			idx := strings.LastIndex(queue.QueueURL, "/")

			name := queue.QueueURL[idx+1:]

			if _, ok := c.doneQueues[name]; !ok {
				c.doneQueues[name] = struct{}{}
			}
		}

		time.Sleep(1 * time.Second)
	}
}

// AddQueue 添加一个消息队列
func (c *Consumer) AddQueue(q *Queue) {
	if q.Parallel == 0 {
		q.Parallel = 1
	}
	l := q.Parallel
	if l > maxReceiveMessage {
		l = maxReceiveMessage
	}
	q.receiveMessageChan = make(chan *ReceiveMessage, l)
	q.longPollQuit = make(chan struct{})
	q.consumeQuit = make(chan struct{})
	q.statusBits = bitset.New(uint(l))
	c.queues = append(c.queues, q)
}

// PeriodicallyFetchQueues 周期性拉取消息队列与内存的消息队列做比较
func (c *Consumer) PeriodicallyFetchQueues() chan struct{} {
	fetchQueueReady := make(chan struct{})
	ticker := time.Tick(time.Minute * 3)

	go func() {
		err := c.BatchListQueue()
		if err != nil {
			contextLogger.WithField("err", err).Info("BatchListQueue")
		} else {
			fetchQueueReady <- struct{}{}
		}

		for {
			select {
			case <-ticker:
				err := c.BatchListQueue()
				if err != nil {
					contextLogger.WithField("err", err).Info("BatchListQueue")
					continue
				} else {
					fetchQueueReady <- struct{}{}
				}
			}
		}
	}()

	return fetchQueueReady
}

// CreateQueueList 创建消息队列
func (c *Consumer) CreateQueueList(fetchQueueReady chan struct{}) chan struct{} {
	createQueueReady := make(chan struct{})
	go func() {
		for {
			select {
			case <-fetchQueueReady:
				for _, queue := range c.queues {
					if _, ok := c.doneQueues[queue.Name]; ok {
						continue
					}

					queue.Stop()

					_, err := c.CreateQueue(queue.Name, queue.QueueAttributeSetters...)
					switch err {
					case nil:
						continue
					case createQueueConflictError, unknownError:
						contextLogger.WithField("err", err).Warn("CreateQueue")
					}
				}

				createQueueReady <- struct{}{}
			}
		}
	}()

	return createQueueReady
}

// Schedule 使消息队列开始运作起来
func (c *Consumer) Schedule(createQueueReady chan struct{}) {
	go func() {
		for {
			select {
			case <-createQueueReady:
				for _, queue := range c.queues {
					if c.isClosed {
						continue
					}

					if queue.IsScheduled {
						continue
					}

					if queue.Parallel <= 0 {
						continue
					}

					if queue.OnReceive == nil {
						continue
					}

					queue.IsScheduled = true

					c.LongPollQueueMessage(queue)

					for i := 1; i <= queue.Parallel; i++ {
						c.ConsumeQueueMessage(queue, i)
					}
				}
			}
		}
	}()
}

// Run 入口函数
func (c *Consumer) Run() {
	fetchQueueReady := c.PeriodicallyFetchQueues()
	createQueueReady := c.CreateQueueList(fetchQueueReady)
	c.Schedule(createQueueReady)
	c.gracefulShutdown()
	select {
	case <-c.shutdown:
		contextLogger.Info("Consumer is closed!")
		return
	}
}

func (c *Consumer) popCount() uint {
	popCount := uint(0)
	for _, queue := range c.queues {
		popCount += queue.statusBits.Count()
	}
	return popCount
}

func (c *Consumer) gracefulShutdown() {
	gracefulStop := make(chan os.Signal)
	signal.Notify(gracefulStop, os.Kill, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-gracefulStop
		contextLogger.WithField("signal", sig.String()).Info("accepting an os signal")

		c.isClosed = true
		for _, queue := range c.queues {
			queue.Stop()
		}

		select {
		case <-time.After(10 * time.Second):
			contextLogger.WithField("count", c.popCount()).Info("timeout shutdown")
			close(c.shutdown)
		case <-time.Tick(time.Second):
			if c.popCount() == 0 {
				contextLogger.Info("graceful shutdown")
				close(c.shutdown)
			}
		}
	}()
}

// LongPollQueueMessage 长轮询消息
func (c *Consumer) LongPollQueueMessage(queue *Queue) {
	go func() {
		for {
			select {
			case <-queue.longPollQuit:
				contextLogger.WithField("queue", queue.Name).Info("long poll quit")
				return
			default:
				time.Sleep(50 * time.Millisecond)
				resp, err := c.BatchReceiveMessage(queue.Name, WithReceiveMessageNumOfMessages(queue.Parallel))
				switch err {
				case messageNotExistError:
					continue
				case nil:
					break
				case queueNotExistError:
					queue.Stop()
					fallthrough
				default:
					contextLogger.WithField("err", err).Warn("BatchReceiveMessage")
					continue
				}

				for _, receiveMessage := range resp.ReceiveMessages {
					queue.receiveMessageChan <- receiveMessage
				}
			}
		}
	}()
}

// OnReceive 消息队列处理函数
func (c *Consumer) OnReceive(queue *Queue, receiveMsg *ReceiveMessage) {
	errChan := make(chan error)
	ticker := time.NewTicker(time.Second * changeVisibilityInterval)
	tickerStop := make(chan struct{})

	rwLock := sync.RWMutex{}

	go func() {
		defer func() {
			if p := recover(); p != nil {
				contextLogger.WithField("err", p).WithField("queue", queue.Name).Error("消息处理函数崩溃")
				errChan <- handleCrashError
			}
		}()
		errChan <- queue.OnReceive(receiveMsg)
	}()

	go func() {
		for {
			select {
			case <-ticker.C:
				resp, err := c.ChangeVisibilityTimeout(queue.Name, receiveMsg.ReceiptHandle, defaultVisibilityTimeout)
				switch {
				case err == nil:
					{
						rwLock.Lock()
						receiveMsg.ReceiptHandle = resp.ReceiptHandle
						receiveMsg.NextVisibleTime = resp.NextVisibleTime
						rwLock.Unlock()
					}
				case err == messageNotExistError, err == queueNotExistError:
					ticker.Stop()
					return
				default:
					contextLogger.WithField("err", err).WithField("queue", queue.Name).Info("ChangeVisibilityTimeout")
				}
			case <-tickerStop:
				ticker.Stop()
				return
			}
		}
	}()

	select {
	case err := <-errChan:
		{
			close(tickerStop)
			if err != nil && !IsHandleCrash(err) {
				contextLogger.WithField("err", err).WithField("queue", queue.Name).Error("OnReceive")
			} else {
				rwLock.RLock()
				err = c.DeleteMessage(queue.Name, receiveMsg.ReceiptHandle)
				rwLock.RUnlock()
				if err != nil {
					contextLogger.WithField("err", err).WithField("queue", queue.Name).Error("DeleteMessage")
				}
			}
		}
	case <-time.After(5 * time.Hour):
		{
			close(tickerStop)
		}
	}
}

// TimestampInMs 毫秒时间戳
func TimestampInMs() int64 {
	return time.Now().UnixNano() / 1000000
}

// Parallel 返回并发数
func Parallel() int {
	return runtime.NumCPU()
}

// ConsumeQueueMessage 消费消息
func (c *Consumer) ConsumeQueueMessage(queue *Queue, idx int) {
	go func() {
		for {
			select {
			case receiveMessage := <-queue.receiveMessageChan:
				{
					if receiveMessage.NextVisibleTime < TimestampInMs() {
						contextLogger.WithField("queue", queue.Name).WithField("body", receiveMessage.MessageBody).Warning("Messages are stacked.")
						continue
					}

					if receiveMessage.DequeueCount > dequeueCount {
						contextLogger.
							WithField("queue", queue.Name).
							WithField("body", receiveMessage.MessageBody).
							WithField("count", receiveMessage.DequeueCount).
							Error("The message is dequeued many times.")
					}

					queue.statusBits.Set(uint(idx))
					c.OnReceive(queue, receiveMessage)
					queue.statusBits.Clear(uint(idx))
				}
			case <-queue.consumeQuit:
				{
					contextLogger.WithField("queue", queue.Name).Info("quit")
					return
				}
			}
		}
	}()
}
