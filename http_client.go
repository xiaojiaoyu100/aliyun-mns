package aliyun_mns

import (
	"net/http"
	"time"
)

// 这里设置的timeout应该大于长轮询的最大时间30s，暂定为60s。
var httpClient = &http.Client{
	Timeout: 60 * time.Second,
}
