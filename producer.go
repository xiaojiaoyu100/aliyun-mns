package alimns

// Producer 生产者
type Producer struct {
	*Client
}

// NewProducer 产生生产者
func NewProducer(client *Client) *Producer {
	producer := new(Producer)
	producer.Client = client
	return producer
}
