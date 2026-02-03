package mylog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type LogWriter interface {
	io.Writer
	Stop(ctx context.Context) error
}

var (
	ErrQueueClosed = errors.New("queue Had be closed")
)

const (
	defaultQuantitativeSize = 100
	defaultQueueSize        = 200
	defaultWorkers          = 2
	defaultTiming           = 100 * time.Millisecond
	defaultStopTimeout      = 2 * time.Second
)

const debugLogPrefix = "[log writer proxy]:"

type ErrHandler func(error)

type LogAsyncConfig struct {
	QueueSize        int
	QuantitativeSize int
	Timing           time.Duration
	StopTimeout      time.Duration
	FlushWorkers     int
	Debug            bool
	ErrHandler       ErrHandler
}

type taskEntry struct {
	taskCnt int
	data    []byte
}

type logWriterProxy struct {
	writers          []LogWriter
	flushWorkers     int
	queue            chan []byte
	batchBuffer      [][]byte
	quantitativeSize int
	mergeMu          sync.Mutex
	writeMu          sync.Mutex
	timingTicker     *time.Ticker
	monitorTicker    *time.Ticker
	done             chan struct{}
	tasks            chan taskEntry
	closed           atomic.Bool
	receiverExit     chan struct{}
	wg               sync.WaitGroup
	batchTiming      time.Duration
	stopTimeout      time.Duration
	debug            bool
	errHandler       ErrHandler
}

func NewLogWriterProxy(writers []LogWriter, config LogAsyncConfig) LogLifetimeWriter {
	w := &logWriterProxy{
		writers:      writers,
		done:         make(chan struct{}),
		receiverExit: make(chan struct{}),
		debug:        config.Debug,
		errHandler:   config.ErrHandler,
	}

	w.init(config)

	go w.receiveAndMergeLogEvent()

	w.writeInBackend()

	return w
}

func (w *logWriterProxy) OnCreated(_ chan<- error) {
}

func (w *logWriterProxy) OnDispose(ctx context.Context) error {
	return w.Stop(ctx)
}

func (w *logWriterProxy) Write(p []byte) (n int, err error) {
	if w.closed.Load() {
		return 0, ErrQueueClosed
	}

	// async handling, copy required
	np := make([]byte, len(p))
	copy(np, p)

	// enqueue
	select {
	case <-w.done:
		return 0, ErrQueueClosed
	case w.queue <- np:
		return len(p), nil
	default:
		select {
		case w.queue <- np:
			return len(p), nil
		default:
			return 0, errors.New(fmt.Sprintf("%s queue buffer fully, discard this event:%s", debugLogPrefix, string(p)))
		}
	}
}

func (w *logWriterProxy) Stop(ctx context.Context) error {
	if !w.closed.CompareAndSwap(false, true) {
		return nil
	}

	close(w.done)

	w.timingTicker.Stop()
	if w.monitorTicker != nil {
		w.monitorTicker.Stop()
	}

	flushed := make(chan error, 1)
	go func() {
		defer close(flushed)

		w.cleanWhenStop(ctx)

		errs := make([]error, 0, len(w.writers))
		for _, writer := range w.writers {
			if err := writer.Stop(ctx); err != nil {
				errs = append(errs, err)
			}
		}

		flushed <- errors.Join(errs...)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(w.stopTimeout):
		return errors.New(fmt.Sprintf("%s stopped timeout after:%v", debugLogPrefix, w.stopTimeout))
	case err := <-flushed:
		if err != nil {
			return err
		}
		log.Printf("%s stopped successfully, metrics:%s\n", debugLogPrefix, "")
		//mm := tw.Monitor()
		//infoBytes, _ := utils.FmtJson(&mm)
		return nil
	}
}

func (w *logWriterProxy) receiveAndMergeLogEvent() {
	defer close(w.receiverExit)

	for {
		select {
		case <-w.done:
			return
		case b, ok := <-w.queue:
			if !ok {
				return
			}

			func() {
				w.mergeMu.Lock()

				w.batchBuffer = append(w.batchBuffer, b)

				// quantitative trigger
				if len(w.batchBuffer) >= w.quantitativeSize {
					// stop ticker
					w.timingTicker.Stop()

					newBytes := w.copyFromBuffer()

					w.mergeMu.Unlock()

					if w.debug {
						log.Printf("%s merge events, triggered with quantitative, batchSize:%d\n", debugLogPrefix, len(newBytes))
					}

					w.submit(newBytes)

					// reset
					w.timingTicker.Reset(w.batchTiming)
				} else {
					w.mergeMu.Unlock()
				}
			}()
		case <-w.timingTicker.C:
			// timing trigger
			func() {
				w.mergeMu.Lock()

				if len(w.batchBuffer) > 0 {
					newBytes := w.copyFromBuffer()

					w.mergeMu.Unlock()

					if w.debug {
						log.Printf("%s merge events, triggered with timing, batchSize:%d\n", debugLogPrefix, len(newBytes))
					}

					w.submit(newBytes)
				} else {
					w.mergeMu.Unlock()
				}
			}()
		}
	}
}

func (w *logWriterProxy) writeInBackend() {
	workers := w.flushWorkers
	w.wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer w.wg.Done()

			for {
				select {
				case <-w.done:
					return
				case entry, ok := <-w.tasks:
					if !ok {
						return
					}

					//start := time.Now()
					err := w.doFlush(entry, true)
					if err != nil {
						//tw.monitor.DeliverFailed(uint64(entry.taskCnt))
						w.handleError(err)
					} else {
						//tw.monitor.DeliverSuccess(uint64(entry.taskCnt))
					}

					//tw.monitor.UpdateTook(uint64(time.Since(start).Milliseconds()))
				}
			}
		}()
	}
}

func (w *logWriterProxy) copyFromBuffer() [][]byte {
	batches := w.batchBuffer

	w.batchBuffer = make([][]byte, 0, w.quantitativeSize)

	return batches
}

func (w *logWriterProxy) submit(entries [][]byte) {
	var batch bytes.Buffer
	for _, entry := range entries {
		batch.Write(entry)

		if len(entry) == 0 || entry[len(entry)-1] != '\n' {
			batch.WriteByte('\n')
		}
	}

	w.tasks <- taskEntry{
		taskCnt: len(entries),
		data:    batch.Bytes(),
	}
}

func (w *logWriterProxy) handleError(err error) {
	if w.errHandler != nil {
		w.errHandler(err)
	} else {
		log.Printf("%s handle failed:%v\n", debugLogPrefix, err)
	}
}

func (w *logWriterProxy) doFlush(entry taskEntry, requeue bool) error {
	if w.debug {
		log.Printf("%s write entries start, batchSize:%d, dataSize:%d\n", debugLogPrefix, entry.taskCnt, len(entry.data))
	}

	errs := make([]error, 0, len(w.writers))
	for _, writer := range w.writers {
		_, err := writer.Write(entry.data)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// both failed, try requeue
	if requeue && len(errs) == len(w.writers) {
		select {
		case w.tasks <- entry:
		default:

		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (w *logWriterProxy) cleanWhenStop(ctx context.Context) {
	// waiting receiver  exit completed, to avoid data race for use batchBuffer
	<-w.receiverExit

	// waiting backend worker done it work
	w.wg.Wait()

	select {
	case <-ctx.Done():
	default:
	}

	log.Printf("%s received stop signal, start flush remain log events\n", debugLogPrefix)

	flushStart := time.Now()

	close(w.queue)

	// take out all remain log event in queue
	var remains [][]byte
	for bt := range w.queue {
		remains = append(remains, bt)
	}

	w.mergeMu.Lock()

	if len(w.batchBuffer) > 0 {
		remains = append(remains, w.batchBuffer...)
		w.batchBuffer = nil
	}

	w.mergeMu.Unlock()

	if len(remains) > 0 {
		batchSize := w.quantitativeSize
		for i := 0; i < len(remains); i += batchSize {
			end := i + batchSize
			if end > len(remains) {
				end = len(remains)
			}

			w.submit(remains[i:end])
		}
	}

	close(w.tasks)

	// take out all remain log event in tasks
	var mergedRemains []taskEntry
	for entry := range w.tasks {
		mergedRemains = append(mergedRemains, entry)
	}

	if len(mergedRemains) > 0 {
		remainCnt := 0
		for _, entry := range mergedRemains {
			select {
			case <-ctx.Done():
			default:
			}

			remainCnt += entry.taskCnt

			if err := w.doFlush(entry, false); err != nil {
				//w.monitor.DeliverFailed(uint64(entry.taskCnt))
				w.handleError(err)
			} else {
				//tw.monitor.DeliverSuccess(uint64(entry.taskCnt))
			}
		}

		log.Printf("%s flush remain log events completed, remainCnt:%d, took:%v\n", debugLogPrefix, remainCnt, time.Since(flushStart))
	}
}

func (w *logWriterProxy) init(config LogAsyncConfig) {
	var queueSize = config.QueueSize
	if queueSize == 0 {
		queueSize = defaultQueueSize
	}
	w.queue = make(chan []byte, queueSize)

	var quantitativeSize = config.QuantitativeSize
	if quantitativeSize == 0 {
		quantitativeSize = defaultQuantitativeSize
	}
	w.batchBuffer = make([][]byte, 0, quantitativeSize)
	w.quantitativeSize = quantitativeSize

	w.tasks = make(chan taskEntry, queueSize/quantitativeSize+1)

	var timing = config.Timing
	if timing == 0 {
		timing = defaultTiming
	}
	w.batchTiming = timing
	w.timingTicker = time.NewTicker(timing)

	var workers = config.FlushWorkers
	if workers == 0 {
		workers = defaultWorkers
	}
	w.flushWorkers = workers

	var stopTimeout = config.StopTimeout
	if stopTimeout == 0 {
		stopTimeout = defaultStopTimeout
	}
	w.stopTimeout = stopTimeout
}
