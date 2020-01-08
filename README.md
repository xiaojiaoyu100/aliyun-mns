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
	"github.com/xiaojiaoyu100/aliyun-mns"
)

func Handle1(m *alimns.M) error {
	return nil
}

func Handle2(m *alimns.M) error {
	return nil
}

func main() {
	client, err := alimns.NewClient(alimns.Config{
		Endpoint: "",
		QueuePrefix: "", // 可以留空，表示拉取全部消息队列
		AccessKeyID: "",
		AccessKeySecret: "",
	})
	if err != nil {
		return
	}
	consumer := alimns.NewConsumer(client)
	err = consumer.AddQueue(
		&alimns.Queue{
			Name:      "QueueTest1",
			OnReceive: Handle1,
		},
	)
	if err != nil {
		return
	}
	err = consumer.AddQueue(
		&alimns.Queue{
			Name:      "QueueTest2",
			OnReceive: Handle2,
            Backoff:   ExponentialBackoff(60, 3600), // 指数回退，1分钟起始，最长1小时
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
