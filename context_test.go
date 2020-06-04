package alimns

import (
	"context"
	"testing"
)

func TestHandleErrFrom(t *testing.T) {
	t.Log(HandleErrFrom(context.TODO()))
}
