package alimns

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"time"

	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"

	"context"

	"github.com/sirupsen/logrus"
	"github.com/xiaojiaoyu100/cast"
)

// Client 存储了阿里云的相关信息
type Client struct {
	config      Config
	makeContext MakeContext
	clean       Clean
	codec       Codec
	ca          *cast.Cast
	logger      *zap.Logger
}

// NewClient 返回Client的实例
func NewClient(config Config) (*Client, error) {
	zc := &zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.WarnLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := zc.Build()
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
func (c *Client) AddLogHook(f LogHook) {
	c.logger = c.logger.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
		f(entry)
		return nil
	}))
}

// SetMakeContext 设置环境
func (c *Client) SetMakeContext(makeContext MakeContext) {
	c.makeContext = makeContext
}

// DefaultMakeContextIfNone 保证makeContext不为空
func (c *Client) defaultMakeContext() {
	if c.makeContext == nil {
		c.makeContext = func(m *M) (context.Context, error) {
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

// SetClean 消息队里善后处理函数
func (c *Client) SetClean(clean Clean) {
	c.clean = clean
}

func (c *Client) defaultClean() {
	if c.clean == nil {
		c.clean = func(ctx context.Context) {}
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
