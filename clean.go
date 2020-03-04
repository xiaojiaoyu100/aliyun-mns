package alimns

import "context"

// Clean 处理函数善后处理
type Clean func(ctx context.Context)
