package alimns

import (
	"github.com/go-redis/redis"
)

// Cmdable represents redis op collection.
type Cmdable interface {
	Pipeline() redis.Pipeliner
	RPush(key string, values ...interface{}) *redis.IntCmd
	LRem(key string, count int64, value interface{}) *redis.IntCmd
}
