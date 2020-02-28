package alimns

import (
	"context"
)

type contextKey int

const (
	aliyunMnsM          contextKey = 0
	aliyunMnsContextErr contextKey = 1
)

// MakeContext 生成一个context
type MakeContext func(m *M) (context.Context, error)

// MFrom 拿出message
func MFrom(ctx context.Context) (*M, bool) {
	m, ok := ctx.Value(aliyunMnsM).(*M)
	return m, ok
}

// ErrFrom 拿出context error
func ErrFrom(ctx context.Context) (*M, bool) {
	m, ok := ctx.Value(aliyunMnsM).(*M)
	return m, ok
}
