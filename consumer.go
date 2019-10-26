package alimns

import (
	"context"
	"encoding/base64"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/xiaojiaoyu100/curlew"
)

const (
	changeVisibilityInterval = 5 * time.Second
)

// Consumer 消费者
type Consumer struct {
	*Client
	queues     []*Queue
	doneQueues map[string]struct{}
	shutdown   chan struct{}
	isClosed   bool
	runningNum int32
}

// NewConsumer 生成了一个消费者
func NewConsumer(client *Client) *Consumer {
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
	request.Prefix = c.queuePrefix
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

		request.Marker = resp.NextMarker

		resp, err = c.ListQueue(request)

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

func setParallel(parallel int) int {
	if parallel > maxReceiveMessage {
		return maxReceiveMessage
	}
	if parallel == 0 {
		p := Parallel()
		if p < maxReceiveMessage {
			return p
		}
		return maxReceiveMessage
	}
	return parallel
}

// AddQueue 添加一个消息队列
func (c *Consumer) AddQueue(q *Queue) error {
	var err error
	q.Parallel = setParallel(q.Parallel)
	q.receiveMessageChan = make(chan *ReceiveMessage)
	q.longPollQuit = make(chan struct{})
	q.consumeQuit = make(chan struct{})

	monitor := func(e error) {
		c.log.WithError(err).Warning("curlew")
	}

	q.dispatcher, err = curlew.New(
		curlew.WithMaxWorkerNum(q.Parallel),
		curlew.WithJobSize(q.Parallel),
		curlew.WithMonitor(monitor),
	)
	if err != nil {
		return err
	}
	c.queues = append(c.queues, q)
	return nil
}

// PeriodicallyFetchQueues 周期性拉取消息队列与内存的消息队列做比较
func (c *Consumer) PeriodicallyFetchQueues() chan struct{} {
	fetchQueueReady := make(chan struct{})
	ticker := time.NewTicker(time.Minute * 3)

	go func() {
		err := c.BatchListQueue()
		if err != nil {
			c.log.WithError(err).Warning("BatchListQueue")
		} else {
			fetchQueueReady <- struct{}{}
		}

		for range ticker.C {
			err := c.BatchListQueue()
			if err != nil {
				c.log.WithError(err).Warning("BatchListQueue")
				continue
			} else {
				fetchQueueReady <- struct{}{}
			}
		}
	}()

	return fetchQueueReady
}

// CreateQueueList 创建消息队列
func (c *Consumer) CreateQueueList(fetchQueueReady chan struct{}) chan struct{} {
	createQueueReady := make(chan struct{})
	go func() {
		for range fetchQueueReady {
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
					c.log.WithError(err).Warn("CreateQueue")
				}
			}
			createQueueReady <- struct{}{}
		}
	}()

	return createQueueReady
}

func randInRange(min, max int) int {
	return rand.Intn(max-min) + min
}

// Schedule 使消息队列开始运作起来
func (c *Consumer) Schedule(createQueueReady chan struct{}) {
	go func() {
		for range createQueueReady {
			for _, queue := range c.queues {
				time.Sleep(time.Duration(randInRange(20, 51)) * time.Millisecond)

				if c.isClosed {
					continue
				}

				if queue.isScheduled {
					continue
				}

				if queue.Parallel <= 0 {
					continue
				}

				if queue.OnReceive == nil {
					continue
				}

				queue.isScheduled = true

				c.LongPollQueueMessage(queue)
				c.ConsumeQueueMessage(queue)

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
	<-c.shutdown
	c.log.Debugln("Consumer is closed!")
}

func (c *Consumer) PopCount() int32 {
	return c.runningNum
}

func (c *Consumer) gracefulShutdown() {
	gracefulStop := make(chan os.Signal)
	signal.Notify(gracefulStop, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-gracefulStop
		c.log.WithField("signal", sig.String()).Debug("Accepting an os signal...")

		c.isClosed = true
		for _, queue := range c.queues {
			queue.Stop()
		}

		doom := time.NewTimer(10 * time.Second)
		check := time.NewTicker(1 * time.Second)

		for {
			select {
			case <-doom.C:
				c.log.WithField("count", c.PopCount()).Debugln("timeout shutdown")
				close(c.shutdown)
				return
			case <-check.C:
				popCount := c.PopCount()
				c.log.WithField("count", popCount).Debug("check")
				if popCount == 0 {
					c.log.Debugln("graceful shutdown")
					close(c.shutdown)
					return
				}
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
				c.log.WithField("queue", queue.Name).Debug("long poll quit")
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
					c.log.WithError(err).Warn("BatchReceiveMessage")
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
	ticker := time.NewTicker(changeVisibilityInterval)
	tickerStop := make(chan struct{})

	rwLock := sync.RWMutex{}

	go func() {
		defer func() {
			if p := recover(); p != nil {
				c.log.WithField("err", p).WithField("queue", queue.Name).Error("消息处理函数崩溃")
				errChan <- handleCrashError
			}
		}()
		m := new(M)
		var body string
		if IsBase54(receiveMsg.MessageBody) {
			b64bytes, err := base64.StdEncoding.DecodeString(receiveMsg.MessageBody)
			if err != nil {
				c.log.WithError(err).WithField("queue", queue.Name).Error("尝试解析消息体失败(base64.StdEncoding)")
			}
			body = string(b64bytes)
		} else {
			body = receiveMsg.MessageBody
		}
		if receiveMsg.DequeueCount > dequeueCount {
			c.log.
				WithField("queue", queue.Name).
				WithField("message_id", receiveMsg.MessageID).
				WithField("receipt_handle", receiveMsg.ReceiptHandle).
				WithField("body", body).
				WithField("count", receiveMsg.DequeueCount).
				Error("The message is dequeued many times.")
		}
		m.QueueName = queue.Name
		m.MessageBody = body
		m.EnqueueTime = receiveMsg.EnqueueTime
		errChan <- queue.OnReceive(m)
	}()

	go func() {
		for {
			select {
			case <-ticker.C:
				resp, err := c.ChangeVisibilityTimeout(queue.Name, receiveMsg.ReceiptHandle, defaultVisibilityTimeout)
				switch {
				case err == nil:
					rwLock.Lock()
					receiveMsg.ReceiptHandle = resp.ReceiptHandle
					receiveMsg.NextVisibleTime = resp.NextVisibleTime
					rwLock.Unlock()
				case err == messageNotExistError, err == queueNotExistError:
					ticker.Stop()
					return
				default:
					c.log.WithError(err).WithField("queue", queue.Name).Error("ChangeVisibilityTimeout")
				}
			case <-tickerStop:
				ticker.Stop()
				return
			}
		}
	}()

	select {
	case err := <-errChan:
		close(tickerStop)
		switch {
		case IsHandleCrash(err):
			// 这里不报警
		case err != nil:
			c.log.WithError(err).WithField("queue", queue.Name).Error("OnReceive")
			if queue.Backoff != nil {
				_, err = c.ChangeVisibilityTimeout(queue.Name, receiveMsg.ReceiptHandle, queue.Backoff(receiveMsg))
				if err != nil {
					c.log.WithError(err).WithField("queue", queue.Name).Error("ChangeVisibilityTimeout")
				}
			}
		default:
			rwLock.RLock()
			err = c.DeleteMessage(queue.Name, receiveMsg.ReceiptHandle)
			rwLock.RUnlock()
			if err != nil {
				c.log.WithError(err).WithField("queue", queue.Name).Error("DeleteMessage")
			}
		}
	case <-time.After(10 * time.Hour):
		close(tickerStop)
	}
}

// TimestampInMs 毫秒时间戳
func TimestampInMs() int64 {
	return time.Now().UnixNano() / 1000000
}

// Parallel 返回并发数
func Parallel() int {
	p := runtime.NumCPU() * 2
	if p > maxReceiveMessage {
		return maxReceiveMessage
	}
	return p
}

// ConsumeQueueMessage 消费消息
func (c *Consumer) ConsumeQueueMessage(queue *Queue) {
	go func() {
		for {
			select {
			case receiveMessage := <-queue.receiveMessageChan:
				{
					if receiveMessage.NextVisibleTime < TimestampInMs() {
						c.log.WithField("queue", queue.Name).WithField("body", receiveMessage.MessageBody).Warning("Messages are stacked.")
						continue
					}

					j := curlew.NewJob()
					j.Arg = receiveMessage
					j.Fn = func(ctx context.Context, arg interface{}) error {
						rm := arg.(*ReceiveMessage)
						atomic.AddInt32(&c.runningNum, 1)
						c.OnReceive(queue, rm)
						atomic.AddInt32(&c.runningNum, -1)
						return nil
					}
					queue.dispatcher.Submit(j)
				}
			case <-queue.consumeQuit:
				c.log.WithField("queue", queue.Name).Debug("Consumer quit")
				return
			}
		}
	}()
}
