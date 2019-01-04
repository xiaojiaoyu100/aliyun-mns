package alimns

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// LogHook log hook模板
type LogHook func(...interface{})

var contextLogger = log.WithFields(log.Fields{
	"source": "alimns",
})

func init() {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     true})
	log.SetReportCaller(true)
	log.SetOutput(os.Stderr)
	log.SetLevel(log.InfoLevel)
}

// AddLogHook add a log reporter.
func AddLogHook(f LogHook) {
	m := NewMonitor(f)
	log.AddHook(m)
}

// Monitor 信息监控
type Monitor struct {
	Callback LogHook
}

// NewMonitor 返回一个实例
func NewMonitor(l LogHook) *Monitor {
	m := new(Monitor)
	m.Callback = l
	return m
}

// Levels 这些级别的日志会被回调
func (m *Monitor) Levels() []log.Level {
	return []log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
		log.WarnLevel,
	}
}

// Fire 实际执行了回调
func (m *Monitor) Fire(entry *log.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	m.Callback(line)
	return nil
}
