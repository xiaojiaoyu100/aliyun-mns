package alimns

import (
	"context"
	"errors"
	"testing"
)

func TestErrFrom(t *testing.T) {
	ctx := context.WithValue(context.TODO(), aliyunMnsContextErr, errors.New("ssss"))
	err := ErrFrom(ctx)
	if err == nil {
		t.Errorf("ErrFrom failed")
		t.FailNow()
	}
}
