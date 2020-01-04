package alimns

import (
	"errors"
	"fmt"
	"strings"
)

// EndPoint 是用户订阅主题时，指定接收消息的终端地址；
// 当有消息发布到主题时，MNS会主动将消息推送到对应的 Endpoint； 多个Subscription可以指定同一个Endpoint。
type EndPoint interface {
	EndPoint() (string, error)
}

// HTTPEndPoint HTTP格式的Endpoint
type HTTPEndPoint string

func getAccountID(endpoint string) (string, error) {
	ret := strings.Split(endpoint, ".")
	if len(ret) != 5 {
		return "", errors.New("endpoint 格式正确")
	}

	accountID := ret[0]
	accountID = strings.Replace(accountID, "https://", "", 1)
	accountID = strings.Replace(accountID, "http://", "", 1)

	return accountID, nil
}

func getRegion(endpoint string) (string, error) {
	ret := strings.Split(endpoint, ".")
	if len(ret) != 5 {
		return "", errors.New("endpoint 格式正确")
	}

	return ret[2], nil
}

// EndPoint 获取EndPoint的字符串
func (ep HTTPEndPoint) EndPoint() (string, error) {
	if !strings.HasPrefix(string(ep), "http") {
		return "", errors.New(`httpEndpoint，必须以"http://"为前缀`)
	}

	return string(ep), nil
}

// QueueEndPoint 队列格式的endpoint
type QueueEndPoint struct {
	MnsEndPoint string
	QueueName   string
}

// EndPoint 获取EndPoint的字符串
func (ep QueueEndPoint) EndPoint() (string, error) {
	region, err := getRegion(ep.MnsEndPoint)
	if err != nil {
		return "", err
	}

	accountID, err := getAccountID(ep.MnsEndPoint)
	if err != nil {
		return "", err
	}

	if region == "" || accountID == "" || ep.QueueName == "" {
		return "", errors.New("region, accountID, queueName 参数为空")
	}

	return fmt.Sprintf("acs:mns:%s:%s:queues/%s", region, accountID, ep.QueueName), nil
}
