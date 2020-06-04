package alimns

import "context"

// After 处理函数善后处理
type After func(ctx context.Context)
