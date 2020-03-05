package alimns

import (
	"context"
)

// Builder 构造者
type Builder interface {
	Handle(ctx context.Context) error
}
