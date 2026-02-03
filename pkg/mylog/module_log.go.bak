package mylog

import (
	"github.com/rs/zerolog"
	"sync"
	"sync/atomic"
)

type loggerState struct {
	ll zerolog.Level

	lg zerolog.Logger
}

type moduleLogger struct {
	state atomic.Pointer[loggerState]
}

var (
	rw            sync.RWMutex
	module2logger = make(map[string]*moduleLogger, 4)
	defLevel      int32
)

func AddModuleLogger(module string) {
	addModuleLogger(module, NewLoggerWithMeta(module))
}

func AddModuleLoggerWithFrame(module string, skipFrame int) {
	addModuleLogger(module, NewFrameLoggerWithMeta(module, skipFrame))
}

func addModuleLogger(module string, lg zerolog.Logger) {
	ml := &moduleLogger{}

	defLv := getModuleDefaultLevel()

	ml.state.Store(&loggerState{
		ll: defLv,
		lg: lg.Level(defLv),
	})

	rw.Lock()
	module2logger[module] = ml
	rw.Unlock()
}

func SetLoggerLevel(module string, level string) bool {
	newLl, err := zerolog.ParseLevel(level)
	if err != nil {
		newLl = zerolog.WarnLevel
	}

	rw.RLock()
	ml, ok := module2logger[module]
	rw.RUnlock()

	if !ok {
		return false
	}

	for {
		oldLs := ml.state.Load()
		if oldLs == nil {
			return false
		}

		nl := oldLs.lg.Level(newLl)

		newLs := &loggerState{
			ll: newLl,
			lg: nl,
		}

		if ml.state.CompareAndSwap(oldLs, newLs) {
			return true
		}
	}
}

func SetLoggersLevel(module2level map[string]string) {
	for md, ll := range module2level {
		SetLoggerLevel(md, ll)
	}
}

func GetLogger(module string) zerolog.Logger {
	rw.RLock()
	ml, ok := module2logger[module]
	rw.RUnlock()

	if ok {
		ls := ml.state.Load()
		if ls == nil {
			return zerolog.Nop()
		}
		return ls.lg
	}

	return zerolog.Nop()
}

func setModuleDefaultLevel(ll zerolog.Level) {
	atomic.StoreInt32(&defLevel, int32(ll))
}

func getModuleDefaultLevel() zerolog.Level {
	return zerolog.Level(atomic.LoadInt32(&defLevel))
}
