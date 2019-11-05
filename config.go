package alimns

// Config has client options.
type Config struct {
	Cmdable

	Endpoint        string
	QueuePrefix     string
	AccessKeyID     string
	AccessKeySecret string
}
