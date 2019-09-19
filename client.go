package alimns

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"os"
	"time"

	"github.com/xiaojiaoyu100/cast"

	"github.com/sirupsen/logrus"
)

// Client 存储了阿里云的相关信息
type Client struct {
	ca              *cast.Cast
	log             *logrus.Logger
	queuePrefix     string
	accessKeyID     string
	accessKeySecret string
}

// NewClient 返回Client的实例
func NewClient(endpoint, accessKeyID, accessKeySecret string) (*Client, error) {
	log := logrus.New()
	log.WithFields(logrus.Fields{
		"source": "alimns",
	})
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetReportCaller(true)
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	c, err := cast.New(
		cast.WithHTTPClientTimeout(40*time.Second),
		cast.WithBaseURL(endpoint),
		cast.AddRequestHook(withAuth(accessKeyID, accessKeySecret)),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		ca:              c,
		log:             log,
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
	}, nil
}

// AddLogHook add a log reporter.
func (c *Client) AddLogHook(f LogHook) {
	m := NewMonitor(f)
	c.log.AddHook(m)
}

// SetQueuePrefix sets the query param for ListQueue.
func (c *Client) SetQueuePrefix(prefix string) {
	c.queuePrefix = prefix
}

// Base64Md5 md5值用base64编码
func Base64Md5(s string) (string, error) {
	hash := md5.New()
	_, err := io.WriteString(hash, s)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}
