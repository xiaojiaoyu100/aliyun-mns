package alimns

import (
	"errors"
	"fmt"
	"strings"
)

// EndPoint 是用户订阅主题时，指定接收消息的终端地址；
// 当有消息发布到主题时，MNS会主动将消息推送到对应的 Endpoint； 多个Subscription可以指定同一个Endpoint。
type EndPointer interface {
	EndPoint() (string, error)
}

// HTTPEndPoint HTTP格式的Endpoint
type HTTPEndPoint string

// EndPoint 获取EndPoint的字符串
func (ep HTTPEndPoint) EndPoint() (string, error) {
	if !strings.HasPrefix(string(ep), "http") {
		return "", errors.New(`httpEndpoint，必须以"http://"为前缀`)
	}

	return string(ep), nil
}

// QueueEndPoint 队列格式的endpoint
type QueueEndPoint struct {
	AccountID string
	Region    string
	QueueName string
}

// EndPoint 获取EndPoint的字符串
func (ep QueueEndPoint) EndPoint() (string, error) {
	if ep.Region == "" || ep.AccountID == "" || ep.QueueName == "" {
		return "", errors.New("region, accountID, queueName 参数为空")
	}

	return fmt.Sprintf("acs:mns:%s:%s:queues/%s", ep.Region, ep.AccountID, ep.QueueName), nil
}
