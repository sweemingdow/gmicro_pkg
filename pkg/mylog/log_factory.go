package mylog

import (
	"github.com/rs/zerolog"
	"gmicro_pkg/pkg/app"
)

func NewLogger(module string) zerolog.Logger {
	return newLogger(func(root zerolog.Logger) zerolog.Logger {
		return root.With().Caller().Str("logger", module).Logger()
	})
}

func NewFrameLogger(module string, frame int) zerolog.Logger {
	return newLogger(func(root zerolog.Logger) zerolog.Logger {
		return root.With().CallerWithSkipFrameCount(frame).Str("logger", module).Logger()
	})
}

func NewFrameLoggerWithMeta(module string, frame int) zerolog.Logger {
	return newLogger(func(root zerolog.Logger) zerolog.Logger {
		return wrapMeta(root.With().CallerWithSkipFrameCount(frame).Str("logger", module).Logger())
	})
}

func NewLoggerWithMeta(module string) zerolog.Logger {
	return newLogger(func(root zerolog.Logger) zerolog.Logger {
		return wrapMeta(root.With().Caller().Str("logger", module).Logger())
	})
}

func wrapMeta(lg zerolog.Logger) zerolog.Logger {
	ta := app.GetTheApp()
	return lg.With().
		Str("app_name", ta.GetAppName()).
		Str("app_id", ta.GetAppId()).
		Str("profile", ta.GetProfile()).
		Str("host", ta.GetLocalIp()).
		Logger()
}

func AppLogger() zerolog.Logger {
	return GetLogger("appLogger")
}

func MonitorLogger() zerolog.Logger {
	return GetLogger("monitorLogger")
}

const (
	markerKey = "marker"
)

func AttachMarker(marker string, lg zerolog.Logger) zerolog.Logger {
	return lg.With().Str(markerKey, marker).Logger()
}

func AppLoggerWithStop() zerolog.Logger {
	return AttachMarker("stop", AppLogger())
}

func AppLoggerWithInit() zerolog.Logger {
	return AttachMarker("init", AppLogger())
}

func AppLoggerWithListen() zerolog.Logger {
	return AttachMarker("listen", AppLogger())
}

func AppLoggerWithNotify() zerolog.Logger {
	return AttachMarker("notify", AppLogger())
}

func AppLoggerWithRpc() zerolog.Logger {
	return AttachMarker("rpc_call", AppLogger())
}

func AppLoggerWithBind() zerolog.Logger {
	return AttachMarker("bind", AppLogger())
}

func AppLoggerWithWriteBack() zerolog.Logger {
	return AttachMarker("write_back", AppLogger())
}
