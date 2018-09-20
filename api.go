package aliyun_mns

const (
	mnsListQueue               = "/queues"
	mnsCreateQueue             = "/queues/%s"
	mnsQueueMetaOverride       = "/queues/%s?metaoverride=true"
	mnsGetQueueAttributes      = "/queues/%s"
	mnsDeleteQueue             = "/queues/%s"
	mnsSendMessage             = "/queues/%s/messages"
	mnsReceiveMessage          = "/queues/%s/messages"
	mnsBatchReceiveMessage     = "/queues/%s/messages"
	mnsDeleteMessage           = "/queues/%s/messages?ReceiptHandle=%s"
	mnsPeekMessage             = "/queues/%s/messages?peekonly=true"
	mnsBatchPeekMessage        = "/queues/%s/messages?peekonly=true&numOfMessages=%s"
	mnsChangeMessageVisibility = "/queues/%s/messages?receiptHandle=%s&visibilityTimeout=%s"
)
