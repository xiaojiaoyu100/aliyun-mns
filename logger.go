package aliyun_mns

import (
	"fmt"
	"log"
	"os"
)

type logger interface {
	Output(calldepth int, s string) error
}

type debugLogger struct {
	debug bool
	l     logger
}

var globalLogger debugLogger

func (dl debugLogger) printf(format string, v ...interface{}) {
	if dl.debug && dl.l != nil {
		dl.l.Output(2, fmt.Sprintf(format, v...))
	}
}

func SetDebug(l logger) {
	globalLogger.debug = true
	globalLogger.l = l
}

func QuickDebug() {
	globalLogger.debug = true
	globalLogger.l = log.New(os.Stderr, "[ALIYUN-MNS]", log.LstdFlags|log.Llongfile)
}
