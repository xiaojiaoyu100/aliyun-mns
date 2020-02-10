package alimns

// Config 配置
type Config struct {
	Cmdable

	Endpoint        string
	QueuePrefix     string
	AccessKeyID     string
	AccessKeySecret string
}
