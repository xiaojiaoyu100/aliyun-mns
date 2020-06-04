# aliyun-mns

[aliyun-mns](https://www.aliyun.com/product/mns/)是对阿里云消息服务的封装

## 队列模型

具有以下特点：

* 动态创建队列
* 可以设置消费者数目
* 消息处理时长自适应
* 发送消息重试。目前基于网络错误、阿里云MNS错误码表InternalError重试
* 监控报警
* 优雅的关闭消费者
* 处理函数处理最大时间限制
* 队列消费者使用协程池，每一个消息队列独占自己的协程池
* 发送消息失败保存进入redis，尝试重新发送，提高发送成功率
* 业务需要自己做消息幂等，有可能出现同样消息内容发送多次，这种情况非常罕见

### 消费者

```go
package main

import (
	"context"
	"github.com/go-redis/redis/v7"
	"github.com/xiaojiaoyu100/aliyun-mns/v2"
)

type Builder struct {
}

func (b *Builder) Handle(ctx context.Context) error {
	//return alimns.BackoffError{
	//	Err:  err,
	//	N:  30,
	//}
	//return err 
	//return nil  
}

func Before(m *alimns.M) (context.Context, error) {
	return context.TODO(), nil
}

func After(ctx context.Context) {
}

func main() {
	option := &redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	}

	redisClient := redis.NewClient(option)
	client, err := alimns.NewClient(alimns.Config{
		Cmdable:         redisClient,
		Endpoint:        "",
		QueuePrefix:     "", // 可以留空，表示拉取全部消息队列
		AccessKeyID:     "",
		AccessKeySecret: "",
	})
	if err != nil {
		return
	}

	client.SetBefore(Before)
	client.SetAfter(After)

	consumer := alimns.NewConsumer(client)
	err = consumer.AddQueue(
		&alimns.Queue{
			Name:     "QueueTest1",
			Builder:  &Builder{},
		},
	)
	if err != nil {
		return
	}
	consumer.Run()
}
```

### 生产者

```go
producer := alimns.NewProducer(client)
producer.SendBase64EncodedJSONMessage()
```

## 主题模型

支持以下主题api:

* 支持主题的创建，删除
* 支持订阅主题，取消主题订阅
* 支持向主题发布消息

### 创建/订阅主题

```go
endpoint := QueueEndPoint{
	AccountID: "xxx",
	Region:    "xxx",
	QueueName: "xxx",
}

// 创建
err := client.CreateTopic("topicName")
if err != nil {
	return
}

// 订阅
err = client.Subscribe("topicName", "subscriptionName", endpoint)
if err != nil {
	return
}

```

### 发布消息

```go
messageID, err := client.PublishMessage("topicName", "hello world")
if err != nil {
	return
}
```
