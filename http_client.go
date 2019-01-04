package alimns

import (
	"net/http"
	"time"
)

// 这里设置的timeout应该大于长轮询的最大时间30s，暂定为60s。
var httpClient = &http.Client{
	Timeout: 60 * time.Second,
}

func init() {
	roundTripper := http.DefaultTransport
	if transport, ok := roundTripper.(*http.Transport); ok {
		transport.MaxIdleConns = 2000
		transport.MaxIdleConnsPerHost = 2000
	}
}
