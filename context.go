package alimns

import (
	"context"
)

type contextKey int

const (
	aliyunMnsM          contextKey = 1
	aliyunMnsContextErr contextKey = 2
	aliyunMnsHandleErr  contextKey = 3
)

// Before 生成一个context
type Before func(m *M) (context.Context, error)

// MFrom 拿出message
func MFrom(ctx context.Context) (*M, error) {
	m, _ := ctx.Value(aliyunMnsM).(*M)
	err, _ := ctx.Value(aliyunMnsContextErr).(error)
	return m, err
}

// HandleErrFrom 返回handle的結果
func HandleErrFrom(ctx context.Context) error {
	err, _ := ctx.Value(aliyunMnsHandleErr).(error)
	return err
}
