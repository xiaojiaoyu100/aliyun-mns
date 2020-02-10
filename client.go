package alimns

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"os"
	"time"

	"context"
	"github.com/sirupsen/logrus"
	"github.com/xiaojiaoyu100/cast"
)

// Client 存储了阿里云的相关信息
type Client struct {
	config      Config
	makeContext MakeContext
	codec       Codec
	ca          *cast.Cast
	log         *logrus.Logger
}

// NewClient 返回Client的实例
func NewClient(config Config) (*Client, error) {

	level := logrus.WarnLevel

	log := logrus.New()
	log.WithFields(logrus.Fields{
		"source": "alimns",
	})
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetReportCaller(true)
	log.SetOutput(os.Stdout)
	log.SetLevel(level)

	c, err := cast.New(
		cast.WithHTTPClientTimeout(40*time.Second),
		cast.WithBaseURL(config.Endpoint),
		cast.AddRequestHook(withAuth(config.AccessKeyID, config.AccessKeySecret)),
		cast.WithRetry(3),
		cast.WithConstantBackoffStrategy(time.Millisecond*100),
		cast.WithLogLevel(level),
	)
	if err != nil {
		return nil, err
	}

	cli := &Client{
		config: config,
		ca:     c,
		log:    log,
	}

	cli.defaultCodec()
	cli.defaultMakeContext()

	return cli, nil
}

// AddLogHook add a log reporter.
func (c *Client) AddLogHook(f LogHook) {
	m := NewMonitor(f)
	c.log.AddHook(m)
}

// SetMakeContext 设置环境
func (c *Client) SetMakeContext(makeContext MakeContext) {
	c.makeContext = makeContext
}

// DefaultMakeContextIfNone 保证makeContext不为空
func (c *Client) defaultMakeContext() {
	if c.makeContext == nil {
		c.makeContext = func(m *M) context.Context {
			return context.TODO()
		}
	}
}

// defaultCodec 保证codec不为空，默认是json编解码
func (c *Client) defaultCodec() {
	if c.codec == nil {
		c.codec = JSONCodec{}
	}
}

// EnableDebug enables debug info.
func (c *Client) EnableDebug() {
	c.log.SetLevel(logrus.DebugLevel)
}

// SetQueuePrefix sets the query param for ListQueue.
func (c *Client) SetQueuePrefix(prefix string) {
	c.config.QueuePrefix = prefix
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
