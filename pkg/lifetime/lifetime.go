package lifetime

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

type Creatable interface {
	OnCreated(ec chan<- error)
}

type Disposable interface {
	OnDispose(context.Context) error
}

type LifeCycle interface {
	Creatable

	Disposable
}

type lifeCycleEntry struct {
	key string
	lc  LifeCycle
}

type AppFinalizer struct {
	// stack impl
	entries  []lifeCycleEntry
	mapIndex map[string]int

	mu       sync.Mutex
	ec       chan<- error
	released atomic.Bool
}

type LazyFunc func(chan<- error) (LifeCycle, error)

var (
	finalizer *AppFinalizer
	once      sync.Once
)

func GetAppFinalizer() *AppFinalizer {
	if finalizer == nil {
		panic("finalizer is nil, create firstly")
	}

	return finalizer
}

func NewFinalizer(ec chan<- error) *AppFinalizer {
	once.Do(func() {
		finalizer = &AppFinalizer{
			entries:  make([]lifeCycleEntry, 0),
			mapIndex: make(map[string]int),
			ec:       ec,
		}
	})

	return finalizer
}

func (cl *AppFinalizer) Collect(key string, lc LifeCycle) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if _, loaded := cl.mapIndex[key]; loaded {
		return
	}

	cl.entries = append(cl.entries, lifeCycleEntry{key: key, lc: lc})
	cl.mapIndex[key] = len(cl.entries) - 1 // 记录索引

	// 初始化
	lc.OnCreated(cl.ec)
}

func (cl *AppFinalizer) CollectLazy(key string, lf LazyFunc) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if _, loaded := cl.mapIndex[key]; loaded {
		return
	}

	lc, err := lf(cl.ec)
	if err != nil {
		return
	}

	cl.entries = append(cl.entries, lifeCycleEntry{key: key, lc: lc})
	cl.mapIndex[key] = len(cl.entries) - 1

	lc.OnCreated(cl.ec)
}

// warning: 释放是栈逻辑: LIFO
// 确保先释放的是最后搜集的
func (cl *AppFinalizer) Release(ctx context.Context) ([]error, bool) {
	if !cl.released.CompareAndSwap(false, true) {
		return nil, true
	}

	if ctx == nil {
		ctx = context.Background()
	}

	cl.mu.Lock()
	entries := cl.entries
	cl.mu.Unlock()

	if len(entries) == 0 {
		return nil, true
	}

	var errs []error
	var aborted bool

	// LIFO (后进先出) 清理顺序
	for i := len(entries) - 1; i >= 0; i-- {
		lc := entries[i].lc

		select {
		case <-ctx.Done():
			aborted = true
			errs = append(errs, ctx.Err())
			return errs, aborted
		default:
		}

		if err := lc.OnDispose(ctx); err != nil {
			if errors.Is(err, context.DeadlineExceeded) ||
				errors.Is(err, context.Canceled) {
				aborted = true
				errs = append(errs, err)
				return errs, aborted
			}

			errs = append(errs, err)
		}
	}

	return errs, aborted
}
