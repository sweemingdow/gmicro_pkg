package mylog

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/sweemingdow/log_remote_writer/pkg/writer"
	"github.com/sweemingdow/log_remote_writer/pkg/writer/tcpwriter"
	"gmicro_pkg/pkg/config"
	"gmicro_pkg/pkg/utils"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
)

var (
	_root   zerolog.Logger
	hadInit atomic.Bool
)

func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = utils.ProgramFmt
}

func InitLogger(logCfg config.LogConfig, colorfulStdout bool, appName string) writer.RemoteWriter {
	if !hadInit.CompareAndSwap(false, true) {
		panic("already initialized")
	}

	ll, err := zerolog.ParseLevel(logCfg.Level)
	if err != nil {
		panic(fmt.Sprintf("parse log level failed:%v", err))
	}

	setModuleDefaultLevel(ll)

	var (
		writers = make([]io.Writer, 0, 3)
		rootLog zerolog.Logger
	)

	writers = append(writers, createStdoutWriter(colorfulStdout))

	writers = append(writers, createFileWriter(logCfg.FileLogCfg, appName))

	remoteWriter := createRemoteWriter(logCfg.RemoteLogCfg)
	writers = append(writers, remoteWriter)

	rootLog = zerolog.
		New(zerolog.MultiLevelWriter(writers...)).
		Level(ll).
		With().
		Timestamp().
		Int("pid", os.Getpid()).
		Logger()

	_root = rootLog

	// add application main logger
	AddModuleLogger("appLogger")

	// add application monitor logger
	AddModuleLogger("monitorLogger")

	return remoteWriter
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

func createFileWriter(fileCfg config.FileLogConfig, appName string) io.Writer {
	return &lumberjack.Logger{
		Filename:   filepath.Join(fileCfg.FilePath, appName, "point.log"),
		MaxSize:    fileCfg.MaxFileSize,
		MaxAge:     fileCfg.HistoryDays,
		MaxBackups: fileCfg.MaxBackup,
		Compress:   fileCfg.Compress,
		LocalTime:  true,
	}
}

func createRemoteWriter(remoteCfg config.RemoteLogConfig) writer.RemoteWriter {
	return tcpwriter.New(tcpwriter.TcpRemoteConfig{
		Host:                   remoteCfg.Host,
		Port:                   remoteCfg.Port,
		ReconnectMaxDelayMills: remoteCfg.ReconnectMaxDelayMills,
		QueueSize:              remoteCfg.QueueSize,
		StopTimeoutMills:       remoteCfg.StopTimeoutMills,
		MustConnectedInInit:    remoteCfg.MustConnectedInInit,
		BatchQuantitativeSize:  remoteCfg.BatchQuantitativeSize,
		BatchTimingMills:       remoteCfg.BatchTimingMills,
		Debug:                  remoteCfg.Debug,
	})
}

type LogCreator func(root zerolog.Logger) zerolog.Logger

func newLogger(lc LogCreator) zerolog.Logger {
	return lc(_root)
}
