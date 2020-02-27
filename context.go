package alimns

import (
	"context"
)

type contextKey int

const (
	aliyunMnsM contextKey = 0
)

// MakeContext 生成一个context
type MakeContext func(m *M) context.Context

// MFrom 拿出message
func MFrom(ctx context.Context) (*M, bool) {
	m, ok := ctx.Value(aliyunMnsM).(*M)
	return m, ok
}
