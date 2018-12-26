package alimns

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"runtime"
	"time"
)

// Base64Md5 md5值用base64编码
func Base64Md5(s string) string {
	hash := md5.New()
	io.WriteString(hash, s)
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

// TimestampInMs 毫秒时间戳
func TimestampInMs() int64 {
	return time.Now().UnixNano() / 1000000
}

// Parallel 返回并发数
func Parallel() int {
	return runtime.NumCPU()
}
