package alimns

const (
	mnsListQueue               = "/queues"
	mnsCreateQueue             = "/queues/%s"
	mnsQueueMetaOverride       = "/queues/%s?metaoverride=true"
	mnsGetQueueAttributes      = "/queues/%s"
	mnsDeleteQueue             = "/queues/%s"
	mnsSendMessage             = "/queues/%s/messages"
	mnsBatchSendMessage        = "/queues/%s/messages"
	mnsReceiveMessage          = "/queues/%s/messages"
	mnsBatchReceiveMessage     = "/queues/%s/messages"
	mnsDeleteMessage           = "/queues/%s/messages?ReceiptHandle=%s"
	mnsPeekMessage             = "/queues/%s/messages?peekonly=true"
	mnsBatchPeekMessage        = "/queues/%s/messages?peekonly=true&numOfMessages=%s"
	mnsChangeMessageVisibility = "/queues/%s/messages?receiptHandle=%s&visibilityTimeout=%s"
)

// 主题订阅管理
const (
	mnsCreateTopic               = "/topics/%s"
	mnsSetTopicAttributes        = "/topics/%s?metaoverride=true"
	mnsDeleteTopic               = "/topics/%s"
	mnsSubscribe                 = "/topics/%s/subscriptions/%s"
	mnsSetSubscriptionAttributes = "/topics/%s/subscriptions/%s?metaoverride=true"
	mnsGetSubscriptionAttributes = "/topics/%s/subscriptions/%s"
	mnsUnsubscribe               = "/topics/%s/subscriptions/%s"
	mnsListSubscriptionByTopic   = "/topics/%s/subscriptions"
	mnsPublishMessage            = "/topics/%s/messages"
)
