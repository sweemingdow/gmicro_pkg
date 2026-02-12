package mylog

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/sweemingdow/gmicro_pkg/pkg/config"
	"github.com/sweemingdow/gmicro_pkg/pkg/lifetime"
	"github.com/sweemingdow/gmicro_pkg/pkg/utils"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	_root   zerolog.Logger
	hadInit atomic.Bool
)

const defaultLoggerName = "appLogger"

type stdLogWriter struct {
}

func (w stdLogWriter) Write(p []byte) (n int, err error) {
	return fmt.Fprintf(os.Stderr, "%s %s", utils.Fmt(time.Now(), utils.ProgramFmt), p)
}

func init() {
	//zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.ErrorStackMarshaler = MarshalStackLimited
	zerolog.TimeFieldFormat = utils.ProgramFmt

	log.SetFlags(0)
	log.SetOutput(stdLogWriter{})
}

type LogFileNameGenerator func() string

type LogLifetimeWriter interface {
	lifetime.LifeCycle
	LogWriter
}

func InitLogger(port int, cfg *config.Config, colorfulStdout bool, appName string, skipFrames, extractFrames int, nameGenFunc LogFileNameGenerator) LogLifetimeWriter {
	if !hadInit.CompareAndSwap(false, true) {
		panic("already initialized")
	}

	setSkipFrames(skipFrames)
	setExtractFrames(extractFrames)

	ll, err := zerolog.ParseLevel(cfg.LogCfg.Level)
	zerolog.SetGlobalLevel(ll)
	if err != nil {
		panic(fmt.Sprintf("parse log level failed:%v", err))
	}

	var (
		stdWriter = createStdoutWriter(colorfulStdout)
		writers   = make([]LogWriter, 0, 1)
		rootLog   zerolog.Logger
	)

	writers = append(
		writers,
		createFileWriter(port, cfg, appName, nameGenFunc),
		createTcpLogWriter(cfg.LogCfg.TcpLogWriterCfg),
	)

	lwProxy := createLogWriterProxy(writers, cfg.LogCfg.LogAsyncCfg)
	rootLog = zerolog.
		New(zerolog.MultiLevelWriter(stdWriter, lwProxy)).
		With().
		Caller().
		Timestamp().
		Int("pid", os.Getpid()).
		Str("app_name", cfg.AppCfg.AppName).
		Str("mip", utils.GetLocalIp()).
		Logger()

	_root = rootLog

	NewDecoLogger(defaultLoggerName)

	return lwProxy
}

func createStdoutWriter(colorfulStdout bool) io.Writer {
	if colorfulStdout {
		return zerolog.ConsoleWriter{
			Out:        os.Stdout,
			NoColor:    false,
			TimeFormat: utils.ProgramFmt,
		}
	} else {
		return os.Stdout
	}
}

func createFileWriter(port int, cfg *config.Config, appName string, nameGenFunc LogFileNameGenerator) LogWriter {
	fileCfg := cfg.LogCfg.FileLogCfg

	var logNamePaths []string
	if nameGenFunc != nil {
		logNamePaths = append(logNamePaths, fileCfg.FilePath, appName)
		logNamePaths = append(logNamePaths, nameGenFunc())
		logNamePaths = append(logNamePaths, "point.log")
	} else {
		logNamePaths = []string{
			fileCfg.FilePath,
			appName,
			strconv.Itoa(port),
			"point.log",
		}
	}

	return NewFileLogWriter(cfg.LogCfg.FileLogCfg, filepath.Join(logNamePaths...))
}

func createTcpLogWriter(cfg config.TcpLogWriterConfig) LogWriter {
	return NewTcpLogWriter(TcpLogWriterConfig{
		Host:                cfg.Host,
		Port:                cfg.Port,
		KeepAlive:           cfg.KeepAlive,
		ReconnectMaxDelay:   cfg.ReconnectMaxDelay,
		DialTimeout:         cfg.DialTimeout,
		WriteTimeout:        cfg.WriteTimeout,
		Debug:               cfg.Debug,
		MustConnectedInInit: cfg.MustConnectedInInit,
	})
}

func createLogWriterProxy(writers []LogWriter, asyncCfg config.LogAsyncConfig) LogLifetimeWriter {
	return NewLogWriterProxy(writers, LogAsyncConfig{
		QueueSize:        asyncCfg.QueueSize,
		QuantitativeSize: asyncCfg.QuantitativeSize,
		Timing:           asyncCfg.Timing,
		StopTimeout:      asyncCfg.StopTimeout,
		FlushWorkers:     asyncCfg.FlushWorkers,
		Debug:            asyncCfg.Debug,
		ErrHandler:       nil,
	})
}

type LogCreator func(root zerolog.Logger) zerolog.Logger

func NewLogger(lc LogCreator) zerolog.Logger {
	return lc(_root)
}

func Trace() *zerolog.Event {
	return GetDecoLogger().Trace()
}

func Debug() *zerolog.Event {
	return GetDecoLogger().Debug()
}

func Info() *zerolog.Event {
	return GetDecoLogger().Info()
}

func Warn() *zerolog.Event {
	return GetDecoLogger().Warn()
}

func Error() *zerolog.Event {
	return GetDecoLogger().Error()
}

func Fatal() *zerolog.Event {
	return GetDecoLogger().Fatal()
}

func Panic() *zerolog.Event {
	return GetDecoLogger().Panic()
}
