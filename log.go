package alimns

import "go.uber.org/zap/zapcore"

// LogHook log hook模板
type LogHook func(entry zapcore.Entry)
