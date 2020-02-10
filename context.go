package alimns

import (
	"context"
)

// MakeContext 生成一个context
type MakeContext func(m *M) context.Context
