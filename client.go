package alimns

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"io"
	"time"

	"go.uber.org/zap"

	"github.com/sirupsen/logrus"
	"github.com/xiaojiaoyu100/cast"
)

// Client 存储了阿里云的相关信息
type Client struct {
	config Config
	before Before
	after  After
	codec  Codec
	ca     *cast.Cast
	logger *zap.Logger
}

// NewClient 返回Client的实例
func NewClient(config Config) (*Client, error) {
	logger, err := NewLogger()
	if err != nil {
		return nil, err
	}

	c, err := cast.New(
		cast.WithHTTPClientTimeout(40*time.Second),
		cast.WithBaseURL(config.Endpoint),
		cast.AddRequestHook(withAuth(config.AccessKeyID, config.AccessKeySecret)),
		cast.WithRetry(3),
		cast.WithConstantBackoffStrategy(time.Millisecond*100),
		cast.WithLogLevel(logrus.WarnLevel),
	)
	if err != nil {
		return nil, err
	}

	cli := &Client{
		config: config,
		ca:     c,
		logger: logger,
	}

	cli.defaultCodec()
	cli.defaultMakeContext()
	cli.defaultClean()

	return cli, nil
}

// AddLogHook add a log reporter.
func (c *Client) AddLogHook(f func(entry Entry) error) {
	c.logger = c.logger.WithOptions(Hooks(f))
}

// SetBefore 设置环境
func (c *Client) SetBefore(before Before) {
	c.before = before
}

// DefaultMakeContextIfNone 保证makeContext不为空
func (c *Client) defaultMakeContext() {
	if c.before == nil {
		c.before = func(m *M) (context.Context, error) {
			return context.TODO(), nil
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
	c.logger.Core().Enabled(zap.DebugLevel)
}

// SetQueuePrefix sets the query param for ListQueue.
func (c *Client) SetQueuePrefix(prefix string) {
	c.config.QueuePrefix = prefix
}

// SetAfter 消息队里善后处理函数
func (c *Client) SetAfter(after After) {
	c.after = after
}

func (c *Client) defaultClean() {
	if c.after == nil {
		c.after = func(ctx context.Context) {}
	}
}

// Base64Md5 md5值用base64编码
func Base64Md5(s string) (string, error) {
	h, err := Md5(s)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h), nil
}

// Md5 md5
func Md5(s string) ([]byte, error) {
	hash := md5.New()
	_, err := io.WriteString(hash, s)
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}
