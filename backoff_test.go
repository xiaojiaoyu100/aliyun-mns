package alimns

import "testing"

func TestExponentialBackoff(t *testing.T) {
	backoff := ExponentialBackoff(60, 1000000)
	expected := map[int]int{
		1:  60,
		2:  120,
		3:  240,
		10: 30720,
		11: maxVisibilityTimeout,
	}
	for d, v := range expected {
		b := backoff(&ReceiveMessage{DequeueCount: d})
		if b != v {
			t.Errorf("Timeout is %d seconds when dequeue count is %d, but expected %d seconds", b, d, v)
		}
	}
}
