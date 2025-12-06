package srpc

import (
	"github.com/lesismal/arpc/log"
	"gmicro_pkg/pkg/mylog"
	"sync/atomic"
)

const (
	moduleLoggerName = "arpcAdaptLogger"
)

type arpcAdaptLogger struct {
}

var hadInit atomic.Bool

func InitArpcLogAdapter() {
	if !hadInit.CompareAndSwap(false, true) {
		return
	}

	mylog.AddModuleLogger(moduleLoggerName)
	log.SetLogger(arpcAdaptLogger{})
}

func (al arpcAdaptLogger) SetLevel(lvl int) {
	// Nothing to do
}

func (al arpcAdaptLogger) Debug(format string, v ...interface{}) {
	lg := mylog.GetLogger(moduleLoggerName)
	lg.Debug().Msgf(format, v...)
}

func (al arpcAdaptLogger) Info(format string, v ...interface{}) {
	lg := mylog.GetLogger(moduleLoggerName)
	lg.Info().Msgf(format, v...)
}

func (al arpcAdaptLogger) Warn(format string, v ...interface{}) {
	lg := mylog.GetLogger(moduleLoggerName)
	lg.Warn().Msgf(format, v...)
}

func (al arpcAdaptLogger) Error(format string, v ...interface{}) {
	lg := mylog.GetLogger(moduleLoggerName)
	lg.Error().Msgf(format, v...)
}
