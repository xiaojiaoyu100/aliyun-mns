package alimns

import (
	"encoding/json"
	"time"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Entry 在zapcore.Entry基础上增加 error & []zap.Fields
type Entry struct {
	Level      zapcore.Level `json:"-"`
	LevelStr   string        `json:"level"`
	LoggerName string        `json:"logger_name"`
	Message    string        `json:"message"`
	Caller     string        `json:"caller"`
	Stack      string        `json:"stack"`
	Fields     []zap.Field   `json:"-"` // 不包含error字段
	FieldsJSON string        `json:"-"`
	Err        error         `json:"-"`
	ErrStr     string        `json:"error"`
}

// NewLogger 创建一个logger
func NewLogger() (logger *zap.Logger, err error) {
	conf := zap.NewProductionConfig()
	conf.Sampling = nil
	conf.EncoderConfig = newEncoderConfig()
	logger, err = conf.Build()
	return
}

// Hooks 回调钩子(warn级别以上回调)
func Hooks(hooks ...func(entry Entry) error) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return registerHooksWithErrField(core, hooks...)
	})
}

// DefaultHook 提供一个默认钩子
func DefaultHook(callback func(entryJson string, fieldsJson string)) func(entry Entry) error {
	return func(en Entry) error {
		if en.Level < zapcore.WarnLevel {
			return nil
		}
		b, err := json.Marshal(&en)
		if err != nil {
			return err
		}
		callback(string(b), en.FieldsJSON)
		return nil
	}

}

// NewEntryFromZapEntry zapcore.Entry转换到Entry
func NewEntryFromZapEntry(ent zapcore.Entry) Entry {
	return Entry{
		Level:      ent.Level,
		LevelStr:   ent.Level.String(),
		LoggerName: ent.LoggerName,
		Message:    ent.Message,
		Caller:     ent.Caller.String(),
		Stack:      ent.Stack,
	}
}

type hookedWithErrField struct {
	zapcore.Core
	enc    zapcore.Encoder
	funcs  []func(entry Entry) error
	fields []zap.Field
}

func (h *hookedWithErrField) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	// Let the wrapped Core decide whether to log this message or not. This
	// also gives the downstream a chance to register itself directly with the
	// CheckedEntry.
	if downstream := h.Core.Check(ent, ce); downstream != nil {
		return downstream.AddCore(ent, h)
	}
	return ce
}

func (h *hookedWithErrField) With(fields []zapcore.Field) zapcore.Core {
	fields = append(fields, h.fields...)
	return &hookedWithErrField{
		Core:   h.Core.With(fields),
		enc:    h.enc,
		funcs:  h.funcs,
		fields: fields,
	}
}

func (h *hookedWithErrField) Write(ent zapcore.Entry, fs []zapcore.Field) error {
	if ent.Level < zapcore.WarnLevel {
		return nil
	}

	var err error
	entry := NewEntryFromZapEntry(ent)
	fs = append(h.fields, fs...)
	for _, v := range fs {
		if v.Key == "error" && v.Type == zapcore.ErrorType {
			entry.Err = v.Interface.(error)
			entry.ErrStr = entry.Err.Error()
		} else {
			entry.Fields = append(entry.Fields, v)
		}
	}

	buf, _ := h.enc.EncodeEntry(ent, entry.Fields)
	entry.FieldsJSON = buf.String()
	buf.Free()
	for i := range h.funcs {
		err = multierr.Append(err, h.funcs[i](entry))
	}
	return err
}

func registerHooksWithErrField(core zapcore.Core, hooks ...func(entry Entry) error) zapcore.Core {
	funcList := append([]func(entry Entry) error{}, hooks...)
	ec := newEncoderConfig()
	ec.CallerKey = ""
	ec.LevelKey = ""
	ec.MessageKey = ""
	ec.NameKey = ""
	ec.TimeKey = ""
	ec.StacktraceKey = ""
	enc := zapcore.NewJSONEncoder(ec)
	return &hookedWithErrField{
		Core:  core,
		enc:   enc,
		funcs: funcList,
	}
}

func newEncoderConfig() zapcore.EncoderConfig {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = func(i time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(i.Format("2006-01-02 15:04:05"))
	}
	return ec
}
