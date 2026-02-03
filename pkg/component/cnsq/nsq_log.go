package cnsq

import (
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"strings"
)

const (
	consumerAdaptLoggerName = "nsqConsumerAdaptLogger"
	producerAdaptLoggerName = "nsqProducerAdaptLogger"
)

type nsqAdaptLogger struct {
	dl *mylog.DecoLogger
}

func (l *nsqAdaptLogger) Output(_ int, s string) error {
	if len(s) < 3 {
		l.dl.Info().Msg(s)
		return nil
	}

	level := s[:3]
	message := strings.TrimSpace(s[3:])

	switch level {
	case "DBG":
		l.dl.Debug().Msg(message)
	case "INF":
		l.dl.Info().Msg(message)
	case "WRN":
		l.dl.Warn().Msg(message)
	case "ERR":
		l.dl.Error().Msg(message)
	default:
		l.dl.Info().Msg(s)
	}

	return nil
}

func newProduceAdaptLogger() *nsqAdaptLogger {
	return &nsqAdaptLogger{
		dl: mylog.NewDecoLogger(producerAdaptLoggerName),
	}
}

func newConsumerAdaptLogger() *nsqAdaptLogger {
	return &nsqAdaptLogger{
		dl: mylog.NewDecoLogger(consumerAdaptLoggerName),
	}
}
