package mylog

import (
	"context"
	"github.com/sweemingdow/gmicro_pkg/pkg/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

type fileLogWriter struct {
	logger *lumberjack.Logger
}

func NewFileLogWriter(fileLogCfg config.FileLogConfig, filename string) LogWriter {
	return &fileLogWriter{
		logger: &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    fileLogCfg.MaxFileSize,
			MaxAge:     fileLogCfg.HistoryDays,
			MaxBackups: fileLogCfg.MaxBackup,
			Compress:   fileLogCfg.Compress,
			LocalTime:  true,
		},
	}
}

func (w *fileLogWriter) Write(p []byte) (n int, err error) {
	return w.logger.Write(p)
}

func (w *fileLogWriter) Stop(_ context.Context) error {
	return w.logger.Close()
}
