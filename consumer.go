package alimns

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/vmihailenco/msgpack"
	"github.com/xiaojiaoyu100/curlew"
	"github.com/xiaojiaoyu100/lizard/redispattern/lockguard"
	"go.uber.org/zap"
)

const (
	changeVisibilityInterval = 3 * time.Second
)

// Consumer 消费者
type Consumer struct {
	*Client
	queues     []*Queue
	doneQueues map[string]struct{}
	shutdown   chan struct{}
	isClosed   bool
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
	request.Prefix = c.config.QueuePrefix
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

// AddQueue 添加一个消息队列
func (c *Consumer) AddQueue(q *Queue) error {
	prefix := c.Client.config.QueuePrefix
	if prefix != "" && !strings.HasPrefix(q.Name, prefix) {
		return fmt.Errorf("queue name must start with %s", prefix)
	}

	var err error
	q.Parallel = q.safeParallel()
	q.codec = c.codec
	q.before = c.before
	q.after = c.after
	q.receiveMessageChan = make(chan *ReceiveMessage)
	q.longPollQuit = make(chan struct{})
	q.consumeQuit = make(chan struct{})

	monitor := func(e error) {

	}

	q.dispatcher, err = curlew.New(
		curlew.WithMaxWorkerNum(q.safeParallel()),
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
			c.logger.Warn("BatchListQueue", zap.Error(err))
		} else {
			fetchQueueReady <- struct{}{}
		}

		for range ticker.C {
			err := c.BatchListQueue()
			if err != nil {
				c.logger.Warn("BatchListQueue", zap.Error(err))
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

				h, err := Md5(queue.Name)
				if err != nil {
					continue
				}

				s := hex.EncodeToString(h)

				guard, err := lockguard.New(c.config.Cmdable, fmt.Sprintf("alimns:create:queue:%s", s))
				if err != nil {
					c.logger.Warn("generate lock", zap.Error(err))
					continue
				}

				err = guard.Run(context.TODO(), func(ctx context.Context) error {
					_, err := c.CreateQueue(queue.Name, queue.AttributeSetters...)
					return err
				})
				switch err {
				case nil:
				case createQueueConflictError, unknownError:
					c.logger.Warn("CreateQueue", zap.Error(err))
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

				if queue.safeParallel() <= 0 {
					continue
				}

				if queue.Builder == nil {
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
	c.retrySendMessage()
	c.gracefulShutdown()
	<-c.shutdown
	c.logger.Debug("Consumer is closed!")
}

// PopCount means the current number of running handlers.
func (c *Consumer) PopCount() int32 {
	var popCount int32
	for _, queue := range c.queues {
		popCount += queue.popCount
	}
	return popCount
}

func (c *Consumer) retrySendMessage() {
	go func() {
		for {
			time.Sleep(1 * time.Second)

			if c.config.Cmdable == nil {
				continue
			}

			pipe := c.config.Pipeline()

			strCmd := pipe.RPopLPush(aliyunMnsRetryQueue, aliyunMnsProcessingQueue)
			pipe.Expire(aliyunMnsProcessingQueue, time.Minute*5)

			cmders, err := pipe.Exec()
			if err != nil {
				continue
			}

			if len(cmders) != 2 {
				continue
			}

			strCmd, ok := cmders[0].(*redis.StringCmd)
			if !ok {
				continue
			}

			value, err := strCmd.Result()
			if err != nil {
				continue
			}

			if value == "" {
				continue
			}

			w := &wrapper{}

			err = msgpack.Unmarshal([]byte(value), w)
			if err != nil {
				c.logger.Error(fmt.Sprintf("msgpack.Unmarshal: %s", value), zap.Error(err))
				continue
			}

			_, err = c.send(w.QueueName, w.Message)
			if err != nil {
				c.logger.Error(fmt.Sprintf("send: %s, %v", w.QueueName, w.Message), zap.Error(err))
				continue
			}

			_, err = c.config.LRem(aliyunMnsProcessingQueue, 1, value).Result()
			if err != nil {
				c.logger.Error("LRem", zap.Error(err))
			}
		}
	}()
}

func (c *Consumer) gracefulShutdown() {
	gracefulStop := make(chan os.Signal)
	signal.Notify(gracefulStop, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-gracefulStop
		c.logger.Debug("Accepting an os signal...", zap.String("signal", sig.String()))

		c.isClosed = true
		for _, queue := range c.queues {
			queue.Stop()
		}

		doom := time.NewTimer(10 * time.Second)
		check := time.NewTicker(1 * time.Second)

		for {
			select {
			case <-doom.C:
				c.logger.Debug("timeout shutdown", zap.Int32("count", c.PopCount()))
				close(c.shutdown)
				return
			case <-check.C:
				popCount := c.PopCount()
				c.logger.Debug("check", zap.Int32("count", popCount))
				if popCount == 0 {
					c.logger.Debug("graceful shutdown")
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
				c.logger.Debug("long poll quit", zap.String("queue", queue.Name))
				return
			default:
				time.Sleep(time.Duration(randomIntInRange(10, 101)) * time.Millisecond)

				if queue.PullWait && queue.busy() {
					continue
				}

				num := queue.safePullNumOfMessages()

				resp, err := c.BatchReceiveMessage(queue.Name, WithReceiveMessageNumOfMessages(num))
				switch err {
				case messageNotExistError:
					continue
				case nil:
					break
				case queueNotExistError:
					queue.Stop()
					fallthrough
				default:
					c.logger.Warn("BatchReceiveMessage",
						zap.Error(err),
						zap.String("queue", queue.Name))
					continue
				}

				for _, receiveMessage := range resp.ReceiveMessages {
					queue.receiveMessageChan <- receiveMessage
				}
			}
		}
	}()
}

func (c *Consumer) periodicallyChangeVisibility(queue *Queue, receiveMsg *ReceiveMessage) chan struct{} {
	ticker := time.NewTicker(changeVisibilityInterval)
	tickerStop := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				resp, err := c.ChangeVisibilityTimeout(queue.Name, receiveMsg.ReceiptHandle, defaultVisibilityTimeout)
				switch {
				case err == nil:
					receiveMsg.ReceiptHandle = resp.ReceiptHandle
					receiveMsg.NextVisibleTime = resp.NextVisibleTime
				case err == messageNotExistError, err == queueNotExistError:
					ticker.Stop()
					return
				default:
					c.logger.Error("ChangeVisibilityTimeout",
						zap.Error(err),
						zap.String("queue", queue.Name),
						zap.String("body", receiveMsg.MessageBody),
						zap.String("message_id", receiveMsg.MessageID),
						zap.String("receipt_handle", receiveMsg.ReceiptHandle),
					)
				}
			case <-tickerStop:
				ticker.Stop()
				return
			}
		}
	}()

	return tickerStop
}

// OnReceive 消息队列处理函数
func (c *Consumer) OnReceive(queue *Queue, receiveMsg *ReceiveMessage) {
	errChan := make(chan error)

	tickerStop := c.periodicallyChangeVisibility(queue, receiveMsg)

	go func() {
		defer func() {
			if p := recover(); p != nil {
				e, _ := p.(error)
				c.logger.Error("消息处理函数崩溃", zap.Error(e),
					zap.String("queue", queue.Name),
					zap.String("body", receiveMsg.MessageBody),
					zap.String("message_id", receiveMsg.MessageID),
					zap.String("receipt_handle", receiveMsg.ReceiptHandle),
				)
				errChan <- handleCrashError
			}
		}()
		m := new(M)
		var body string
		if IsBase64(receiveMsg.MessageBody) {
			b64bytes, err := base64.StdEncoding.DecodeString(receiveMsg.MessageBody)
			if err != nil {
				c.logger.Error("尝试解析消息体失败(base64.StdEncoding)", zap.Error(err), zap.String("queue", queue.Name))
			}
			body = string(b64bytes)
		} else {
			body = receiveMsg.MessageBody
		}
		if receiveMsg.DequeueCount > dequeueCount {
			c.logger.Error("The message is dequeued many times.",
				zap.String("queue", queue.Name),
				zap.String("message_id", receiveMsg.MessageID),
				zap.String("receipt_handle", receiveMsg.ReceiptHandle),
				zap.String("body", body),
			)
		}
		m.QueueName = queue.Name
		m.MessageBody = body
		m.EnqueueTime = receiveMsg.EnqueueTime
		m.codec = queue.codec
		m.ReceiptHandle = receiveMsg.ReceiptHandle
		ctx, err := queue.before(m)
		ctx = context.WithValue(ctx, aliyunMnsM, m)
		ctx = context.WithValue(ctx, aliyunMnsContextErr, err)
		err = queue.Handle(ctx)
		defer func() {
			if queue.after == nil {
				return
			}
			ctx = context.WithValue(ctx, aliyunMnsHandleErr, err)
			queue.after(ctx)
		}()
		errChan <- err
	}()

	select {
	case err := <-errChan:
		close(tickerStop)
		switch {
		case IsHandleCrash(err):
			// 这里不报警
		case err != nil:
			switch {
			case errors.As(err, &BackoffError{}):
				b, ok := err.(Backoff)
				if ok {
					resp, err := c.ChangeVisibilityTimeout(queue.Name, receiveMsg.ReceiptHandle, b.Backoff())
					if err != nil {
						c.logger.Error("ChangeVisibilityTimeout",
							zap.String("queue", queue.Name),
							zap.Error(err),
							zap.String("message_id", receiveMsg.MessageID),
							zap.String("receipt_handle", receiveMsg.ReceiptHandle),
							zap.String("body", receiveMsg.MessageBody),
						)
					} else {
						c.logger.Debug(fmt.Sprintf("CurrentVisibleTime = %d, NextVisibleTime=%d", receiveMsg.NextVisibleTime, resp.NextVisibleTime),
							zap.String("queue", queue.Name),
							zap.String("message_id", receiveMsg.MessageID),
							zap.String("receipt_handle", receiveMsg.ReceiptHandle),
							zap.String("body", receiveMsg.MessageBody),
						)
						receiveMsg.NextVisibleTime = resp.NextVisibleTime
						receiveMsg.ReceiptHandle = resp.ReceiptHandle
					}
				}
			}
		default:
			err = c.DeleteMessage(queue.Name, receiveMsg.ReceiptHandle)
			if err != nil {
				c.logger.Error("DeleteMessage",
					zap.String("queue", queue.Name),
					zap.Error(err),
					zap.String("message_id", receiveMsg.MessageID),
					zap.String("receipt_handle", receiveMsg.ReceiptHandle),
					zap.String("body", receiveMsg.MessageBody),
				)
			}
		}
	case <-time.After(24 * time.Hour):
		close(tickerStop)
	}
}

// TimestampInMs 毫秒时间戳
func TimestampInMs() int64 {
	return time.Now().UnixNano() / 1000000
}

// ConsumeQueueMessage 消费消息
func (c *Consumer) ConsumeQueueMessage(queue *Queue) {
	go func() {
		for {
			select {
			case receiveMessage := <-queue.receiveMessageChan:
				{
					if receiveMessage.NextVisibleTime < TimestampInMs() {
						c.logger.Warn("Messages are stacked.", zap.String("queue", queue.Name), zap.String("body", receiveMessage.MessageBody))
						continue
					}

					j := curlew.NewJob()
					j.Arg = receiveMessage
					j.Fn = func(ctx context.Context, arg interface{}) error {
						rm := j.Arg.(*ReceiveMessage)
						atomic.AddInt32(&queue.popCount, 1)
						c.OnReceive(queue, rm)
						atomic.AddInt32(&queue.popCount, -1)
						return nil
					}
					queue.dispatcher.Submit(j)
				}
			case <-queue.consumeQuit:
				c.logger.Debug("Consumer quit", zap.String("queue", queue.Name))
				return
			}
		}
	}()
}
