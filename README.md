# aliyun-mns

aliyun-mns是对阿里云消息服务的封装，具有以下特点：

* 动态创建队列
* 可以设置消费者数目
* 长轮询
* 消息处理时长自适应
* 发送消息重试
* 监控报警
* 优雅的关闭消费者

# 消费者

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
	client := alimns.NewClient(endpoint, accessKeyId, accessKeySecret)
	consumer := alimns.NewConsumer(client)
	consumer.AddQueue(
		&alimns.Queue{
			Name: 	"QueueTest1",
			Handler: Handle1,
		},
		&alimns.Queue{
			Name: 	"QueueTest2",
			Handler: Handle2,
		},
		)
	consumer.Run()
}
```

# 生产者
```go
producer := alimns.NewProducer(client)
producer.SendBase64EncodedJSONMessage()
```
