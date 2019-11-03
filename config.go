package alimns

type Config struct {
	Cmdable

	Endpoint        string
	QueuePrefix     string
	AccessKeyID     string
	AccessKeySecret string
}
