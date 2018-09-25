package aliyun_mns

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"runtime"
	"time"
)

func Base64Md5(s string) string {
	hash := md5.New()
	io.WriteString(hash, s)
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func TimestampInMs() int64 {
	return time.Now().UnixNano() / 1000000
}

func Parallel() int {
	return runtime.NumCPU()
}
