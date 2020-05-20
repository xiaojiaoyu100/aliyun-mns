package alimns

import (
	"time"

	"github.com/go-redis/redis/v7"
)

// Cmdable 封装了redis相关操作
type Cmdable interface {
	Pipeline() redis.Pipeliner
	RPush(key string, values ...interface{}) *redis.IntCmd
	LRem(key string, count int64, value interface{}) *redis.IntCmd
	SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Eval(script string, keys []string, args ...interface{}) *redis.Cmd
	Expire(key string, expiration time.Duration) *redis.BoolCmd
}
