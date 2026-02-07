package executor

import (
	"context"
	"errors"
)

var (
	ErrHadBeClosed = errors.New("executor had be was closed")
)

type Task func()

type Executor interface {
	Submit(task Task) error

	Shutdown(ctx context.Context) error
}
