package mylog

import (
	"github.com/rs/zerolog"
	"sync/atomic"
)

type DecoLogger struct {
	name string
	lg   zerolog.Logger
	ll   atomic.Uint32
}

func (dl *DecoLogger) IsEmpty() bool {
	return zerolog.Level(dl.ll.Load()) == zerolog.NoLevel
}

func (dl *DecoLogger) GetLogger() zerolog.Logger {
	return dl.lvlLogger()
}

func (dl *DecoLogger) Trace() *zerolog.Event {
	lg := dl.lvlLogger()
	return lg.Trace()
}

func (dl *DecoLogger) Debug() *zerolog.Event {
	lg := dl.lvlLogger()
	return lg.Debug()
}

func (dl *DecoLogger) Info() *zerolog.Event {
	lg := dl.lvlLogger()
	return lg.Info()
}

func (dl *DecoLogger) Warn() *zerolog.Event {
	lg := dl.lvlLogger()
	return lg.Warn()
}

func (dl *DecoLogger) Error() *zerolog.Event {
	lg := dl.lvlLogger()
	return lg.Error()
}

func (dl *DecoLogger) Fatal() *zerolog.Event {
	lg := dl.lvlLogger()
	return lg.Fatal()
}

func (dl *DecoLogger) Panic() *zerolog.Event {
	lg := dl.lvlLogger()
	return lg.Panic()
}

func (dl *DecoLogger) lvlLogger() zerolog.Logger {
	return dl.lg.Level(zerolog.Level(dl.ll.Load()))
}

// return old level
func SetLoggerLevel(name string, ll zerolog.Level) zerolog.Level {
	rw.Lock()
	defer rw.Unlock()

	dl, ok := name2logger[name]
	if !ok {
		return zerolog.NoLevel
	}

	old := dl.ll.Swap(uint32(ll))
	return zerolog.Level(old)
}

func SetLoggersLevel(m map[string]zerolog.Level) map[string]zerolog.Level {
	rm := make(map[string]zerolog.Level, len(m))

	for name, ll := range m {
		oldLl := SetLoggerLevel(name, ll)
		rm[name] = oldLl
	}

	return rm
}

// reset all loggers level with the level in config
func ResetAllLevel(ll zerolog.Level) {
	rw.Lock()
	defer rw.Unlock()

	for _, dl := range name2logger {
		dl.ll.Store(uint32(ll))
	}
}
