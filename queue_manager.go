package alimns

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

const (
	changeVisibilityInterval = 5
)

// AddQueue 添加一个消息队列
func (c *Client) AddQueue(q *Queue) {
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
	c.queues = append(c.queues, q)
}

// PeriodicallyFetchQueues 周期性拉取消息队列与内存的消息队列做比较
func (c *Client) PeriodicallyFetchQueues() chan struct{} {
	fetchQueueReady := make(chan struct{})
	ticker := time.Tick(time.Minute * 3)

	go func() {
		err := c.BatchListQueue()
		if err != nil {
			notifyAsync("BatchListQueue err:", err)
		} else {
			fetchQueueReady <- struct{}{}
		}

		for {
			select {
			case <-ticker:
				err := c.BatchListQueue()
				if err != nil {
					notifyAsync("BatchListQueue err:", err)
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
func (c *Client) CreateQueueList(fetchQueueReady chan struct{}) chan struct{} {
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
						notifyAsync("CreateQueue err:", err)
					}
				}

				createQueueReady <- struct{}{}
			}
		}
	}()

	return createQueueReady
}

// Schedule 使消息队列开始运作起来
func (c *Client) Schedule(createQueueReady chan struct{}) {
	go func() {
		for {
			select {
			case <-createQueueReady:
				for _, queue := range c.queues {
					if queue.isRunning {
						continue
					}

					if queue.Parallel <= 0 {
						continue
					}

					if queue.OnReceive == nil {
						continue
					}

					queue.isRunning = true

					c.LongPollQueueMessage(queue)

					for i := 1; i <= queue.Parallel; i++ {
						c.ConsumeQueueMessage(queue)
					}
				}
			}
		}
	}()
}

// Run 入口函数
func (c *Client) Run() {
	fetchQueueReady := c.PeriodicallyFetchQueues()
	createQueueReady := c.CreateQueueList(fetchQueueReady)
	c.Schedule(createQueueReady)
}

// LongPollQueueMessage 长轮询消息
func (c *Client) LongPollQueueMessage(queue *Queue) {
	go func() {
		for {
			select {
			case <-queue.longPollQuit:
				globalLogger.printf("%s long poll quit", queue.Name)
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
					notifyAsync("BatchReceiveMessage err:", err)
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
func (c *Client) OnReceive(queue *Queue, receiveMsg *ReceiveMessage) {
	errChan := make(chan error)
	ticker := time.NewTicker(time.Second * changeVisibilityInterval)
	tickerStop := make(chan struct{})

	rwLock := sync.RWMutex{}

	go func() {
		defer func() {
			if p := recover(); p != nil {
				var buf bytes.Buffer
				fmt.Fprintf(&buf, "%v\n", p)
				fmt.Fprintf(&buf, "%s", debug.Stack())
				notifyAsync(buf.String())
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
					notifyAsync("ChangeVisibilityTimeout err:", err)
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
			if err != nil {
				notifyAsync("OnReceive err:", err)
			} else {
				rwLock.RLock()
				err = c.DeleteMessage(queue.Name, receiveMsg.ReceiptHandle)
				rwLock.RUnlock()
				if err != nil {
					notifyAsync("DeleteMessage err:", err)
				}
			}
		}
	case <-time.After(5 * time.Hour):
		{
			close(tickerStop)
		}
	}
}

// ConsumeQueueMessage 消费消息
func (c *Client) ConsumeQueueMessage(queue *Queue) {
	go func() {
		for {
			select {
			case receiveMessage := <-queue.receiveMessageChan:
				{
					if receiveMessage.NextVisibleTime+1000*changeVisibilityInterval < TimestampInMs() {
						globalLogger.printf("queue=%s overstock message_id=%s message_body=%s.", queue.Name, receiveMessage.MessageID, receiveMessage.MessageBody)
						continue
					}

					if receiveMessage.DequeueCount > dequeueCount {
						content := fmt.Sprintf("dequeue count: %d, queue: %s, body: %s", receiveMessage.DequeueCount, queue.Name, receiveMessage.MessageBody)
						notifyAsync(content)
					}

					c.OnReceive(queue, receiveMessage)
				}
			case <-queue.consumeQuit:
				{
					globalLogger.printf("%s consume quit", queue.Name)
					return
				}
			}
		}
	}()
}
