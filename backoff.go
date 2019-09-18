package alimns

// BackoffFunc 回退重试函数，计算消息重试间隔
type BackoffFunc func(*ReceiveMessage) int

// ExponentialBackoff 指数回退重试
func ExponentialBackoff(base, max int) BackoffFunc {
	if base < minVisibilityTimeout {
		base = minVisibilityTimeout
	}
	if max > maxVisibilityTimeout {
		max = maxVisibilityTimeout
	}
	if base > max {
		return nil
	}

	return func(receiveMsg *ReceiveMessage) int {
		t := base
		p := receiveMsg.DequeueCount

		for {
			p--
			if p <= 0 {
				break
			}
			t <<= 1
			if t >= max {
				t = max
				break
			}
		}
		return t
	}
}
