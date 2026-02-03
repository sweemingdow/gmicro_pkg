package mylog

import (
	"github.com/rs/zerolog"
	"sync"
)

var (
	rw              sync.RWMutex
	name2logger     = make(map[string]*DecoLogger, 4)
	emptyDecoLogger = func() *DecoLogger {
		dl := &DecoLogger{
			name: "emptyLogger",
			lg:   zerolog.Nop(),
		}

		dl.ll.Store(uint32(zerolog.NoLevel))
		return dl
	}()
)

func NewDecoLogger(name string) *DecoLogger {
	rw.RLock()
	dl, ok := name2logger[name]
	rw.RUnlock()
	if ok {
		return dl
	}

	rw.Lock()
	defer rw.Unlock()

	dl, ok = name2logger[name]
	if ok {
		return dl
	}

	gll := zerolog.GlobalLevel()

	dl = &DecoLogger{
		name: name,
		lg: NewLogger(func(root zerolog.Logger) zerolog.Logger {
			return root.With().Str("logger", name).Logger()
		}),
	}
	dl.ll.Store(uint32(gll))

	name2logger[name] = dl
	return dl
}

func GetDecoLogger() *DecoLogger {
	rw.RLock()
	dl, ok := name2logger[defaultLoggerName]
	rw.RUnlock()

	if ok {
		return dl
	}

	return emptyDecoLogger
}

const (
	markerKey = "marker"
)

func GetLoggerWrapMarker(marker string) zerolog.Logger {
	return GetDecoLogger().GetLogger().With().Str(markerKey, marker).Logger()
}

func GetInitMarkerLogger() zerolog.Logger {
	return GetLoggerWrapMarker("init")
}

func GetMonitorMarkLogger() zerolog.Logger {
	return GetLoggerWrapMarker("monitor")
}

func GetStopMarkLogger() zerolog.Logger {
	return GetLoggerWrapMarker("stop")
}

func GetListenerMarkLogger() zerolog.Logger {
	return GetLoggerWrapMarker("listener")
}
