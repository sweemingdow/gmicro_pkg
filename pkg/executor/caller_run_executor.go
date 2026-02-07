package executor

import (
	"context"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"sync"
	"sync/atomic"
	"time"
)

type callerRunExecutor struct {
	pool         *ants.Pool // for fallback
	coreWorkerCh chan Task
	done         chan struct{}
	closed       atomic.Bool
	wg           sync.WaitGroup
}

type CallerRunOptions struct {
	CoreWorkers      int
	MaxWorkers       int
	MaxWaitQueueSize int
	MaxIdleTimeout   time.Duration
}

func NewCallerRunExecutor(options CallerRunOptions) Executor {
	h := &callerRunExecutor{
		coreWorkerCh: make(chan Task, max(1, options.CoreWorkers-1)),
		done:         make(chan struct{}),
	}

	h.newDefaultPool(options)

	h.acquireTask(options)

	return h
}

func (h *callerRunExecutor) Submit(task Task) error {
	if h.closed.Load() {
		return ErrHadBeClosed
	}

	select {
	case h.coreWorkerCh <- task:
		return nil
	default:
		// core workers were busy, fallback into pool
		err := h.pool.Submit(task)

		// pool fully
		if err == ants.ErrPoolOverload {
			task() // call by self
			return nil
		}

		return err
	}
}

func (h *callerRunExecutor) Shutdown(ctx context.Context) error {
	if !h.closed.CompareAndSwap(false, true) {
		return nil
	}

	close(h.done) // notify core workers to exit

	h.wg.Wait() // waiting for core workers completed

	close(h.coreWorkerCh)

	stopDone := make(chan struct{})
	go func() {
		defer close(stopDone)

		for task := range h.coreWorkerCh {
			task()
		}

		select {
		case <-ctx.Done():
		default:

		}

		h.pool.Release()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-stopDone:
		return nil
	}
}

func (h *callerRunExecutor) newDefaultPool(options CallerRunOptions) {
	p, err := ants.NewPool(
		options.MaxWorkers-options.CoreWorkers,
		ants.WithPreAlloc(false),
		ants.WithNonblocking(true),
		ants.WithMaxBlockingTasks(options.MaxWaitQueueSize),
		ants.WithExpiryDuration(options.MaxIdleTimeout),
	)

	if err != nil {
		panic(fmt.Sprintf("create default pool failed, err=%v", err))
	}

	h.pool = p
}

func (h *callerRunExecutor) acquireTask(options CallerRunOptions) {
	h.wg.Add(options.CoreWorkers)

	for i := 0; i < options.CoreWorkers; i++ {
		go func() {
			defer h.wg.Done()

			for {
				select {
				case <-h.done:
					return
				case task, ok := <-h.coreWorkerCh:
					if !ok {
						return
					}

					task()
				}
			}
		}()
	}
}
