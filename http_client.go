package aliyun_mns

import "net/http"

// 这里不要设置timeout, long poll不需要设置timeout, 其余请求请在
var httpClient = &http.Client{}

