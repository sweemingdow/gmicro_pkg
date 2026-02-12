package srpc

import (
	"github.com/lesismal/arpc/log"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"sync/atomic"
)

const (
	arpcAdaptLoggerName = "arpcAdapter"
)

type arpcAdaptLogger struct {
	dl *mylog.DecoLogger
}

var hadInit atomic.Bool

func InitArpcLogAdapter() {
	if !hadInit.CompareAndSwap(false, true) {
		return
	}

	log.SetLogger(arpcAdaptLogger{
		dl: mylog.NewDecoLogger(arpcAdaptLoggerName),
	})
}

func (al arpcAdaptLogger) SetLevel(lvl int) {
	// Nothing to do
}

func (al arpcAdaptLogger) Debug(format string, v ...interface{}) {
	al.dl.Debug().Msgf(format, v...)
}

func (al arpcAdaptLogger) Info(format string, v ...interface{}) {
	al.dl.Info().Msgf(format, v...)
}

func (al arpcAdaptLogger) Warn(format string, v ...interface{}) {
	al.dl.Warn().Msgf(format, v...)
}

func (al arpcAdaptLogger) Error(format string, v ...interface{}) {
	al.dl.Error().Msgf(format, v...)
}
