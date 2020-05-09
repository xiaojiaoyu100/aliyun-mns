package alimns

import (
	"context"
)

type contextKey int

const (
	aliyunMnsM          contextKey = 1
	aliyunMnsContextErr contextKey = 2
)

// MakeContext 生成一个context
type MakeContext func(m *M) (context.Context, error)

// MFrom 拿出message
func MFrom(ctx context.Context) (*M, error) {
	m, _ := ctx.Value(aliyunMnsM).(*M)
	err, _ := ctx.Value(aliyunMnsContextErr).(error)
	return m, err
}
