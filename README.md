# aliyun-mns

aliyun-mns是对阿里云消息服务的封装，具有以下特点：

* 动态创建队列
* 可以设置消费者数目
* 长轮询
* 消息处理时长自适应
* 发送消息重试
* 监控报警

# 消费者

```go
package main

import (
	"log"

	"github.com/xiaojiaoyu100/aliyun-mns"
)

func HandleExample(rm *aliyun_mns.ReceiveMessage) error {
	log.Println(rm.MessageBody)
	return nil
}

func main() {
	aliyun_mns.QuickDebug()
	c := aliyun_mns.New(
		endpoint,
		accessKeyId,
		accessKeySecret)
	c.AddQueue(&aliyun_mns.Queue{
		Name:     "example",
		Parallel: 2,
		QueueAttributeSetters: []aliyun_mns.QueueAttributeSetter{
			aliyun_mns.WithDelaySeconds(10)},
		OnReceive: HandleExample,
	})
	c.Run()
}
```

# 生产者
```go
c.SendMessage("example", "test_data")
```
