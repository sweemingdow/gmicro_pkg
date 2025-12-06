package dlog

import (
	"context"
	"github.com/sweemingdow/gmicro_pkg/pkg/lifetime"
	"github.com/sweemingdow/log_remote_writer/pkg/writer"
)

type remoteLogWriter struct {
	writer writer.RemoteWriter
}

func NewLogRemoteWriter(writer writer.RemoteWriter) lifetime.LifeCycle {
	return &remoteLogWriter{writer: writer}
}

func (rlw *remoteLogWriter) OnCreated(_ chan<- error) {
}

func (rlw *remoteLogWriter) OnDispose(ctx context.Context) error {
	return rlw.writer.Stop(ctx)
}
