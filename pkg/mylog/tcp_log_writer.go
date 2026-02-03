package mylog

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type TcpLogWriterConfig struct {
	Host                string
	Port                int
	KeepAlive           time.Duration
	ReconnectMaxDelay   time.Duration
	DialTimeout         time.Duration
	WriteTimeout        time.Duration
	Debug               bool
	MustConnectedInInit bool
}

var (
	ErrHadBeStopped = errors.New("the tcp log writer had be stopped")
	ErrReconnecting = errors.New("connection invalid, try reconnecting")
)

const (
	tcpLogWriterPrefix = "[tcp log writer]:"
)

type tcpLogWriter struct {
	cfg          TcpLogWriterConfig
	mu           sync.Mutex
	reconnecting atomic.Bool
	closed       atomic.Bool
	conn         *connWrap
}

func NewTcpLogWriter(cfg TcpLogWriterConfig) LogWriter {
	w := &tcpLogWriter{
		cfg: cfg,
	}

	conn, err := w.createConnect()
	w.mu.Lock()
	if err != nil {
		w.conn = nil
	} else {
		w.conn = conn
	}
	w.mu.Unlock()

	if err != nil {
		if cfg.MustConnectedInInit {
			panic(fmt.Sprintf("%s create connection failed with address=%s:%d, err=%v", tcpLogWriterPrefix, cfg.Host, cfg.Port, err))
		} else {
			w.tryReconnect()
		}
	}

	return w
}

func (w *tcpLogWriter) Write(p []byte) (n int, err error) {
	if w.closed.Load() {
		return 0, ErrHadBeStopped
	}

	if w.reconnecting.Load() {
		return 0, ErrReconnecting
	}

	w.mu.Lock()
	curConn := w.conn
	w.mu.Unlock()

	if curConn == nil {
		w.tryReconnect()
		return 0, ErrReconnecting
	}

	start := time.Now()
	_, err = w.conn.write(p, w.cfg.WriteTimeout)

	if err != nil {
		if w.reconnecting.CompareAndSwap(false, true) {
			w.mu.Lock()
			_ = w.conn.close()
			w.conn = nil
			w.mu.Unlock()

			w.reconnect()
		}

		return 0, err
	}

	n = len(p)

	if w.cfg.Debug {
		log.Printf("%s send data bytes to remote, %s:%d, dataLen=%d, took=%v", tcpLogWriterPrefix, w.cfg.Host, w.cfg.Port, n, time.Since(start))
	}

	return
}

func (w *tcpLogWriter) Stop(_ context.Context) error {
	if !w.closed.CompareAndSwap(false, true) {
		return nil
	}

	return w.conn.close()
}

func (w *tcpLogWriter) tryReconnect() {
	if w.reconnecting.CompareAndSwap(false, true) {
		w.reconnect()
	}
}

func (w *tcpLogWriter) reconnect() {
	go func() {
		attempt := 0
		for {
			sleepMills := w.calcReconnectBackoff(attempt)
			if sleepMills > 0 {
				time.Sleep(time.Duration(sleepMills) * time.Millisecond)
			}

			conn, err := w.createConnect()
			if err == nil {
				w.mu.Lock()
				if w.closed.Load() {
					_ = conn.close() // if stopped, close it
					w.mu.Unlock()
					return
				}

				w.conn = conn
				w.reconnecting.Store(false)
				w.mu.Unlock()

				return
			}

			log.Printf("%s reconnect failed, attempt:%d, interval:%d, reason:%v\n", tcpLogWriterPrefix, attempt, sleepMills, err)
			attempt++
		}
	}()
}

func (w *tcpLogWriter) calcReconnectBackoff(attempt int) uint64 {
	if attempt <= 0 {
		return 0
	}

	var backoff uint64 = 1000 * (1 << attempt)
	if backoff > uint64(w.cfg.ReconnectMaxDelay.Milliseconds()) {
		return uint64(w.cfg.ReconnectMaxDelay.Milliseconds())
	}

	return backoff
}

func (w *tcpLogWriter) createConnect() (*connWrap, error) {
	conn, err := net.DialTimeout(
		"tcp",
		fmt.Sprintf("%s:%d", w.cfg.Host, w.cfg.Port),
		w.cfg.DialTimeout,
	)

	if err != nil {
		return nil, err
	}

	w.settingConn(conn)

	return &connWrap{conn: conn}, nil
}

func (w *tcpLogWriter) settingConn(conn net.Conn) {
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(w.cfg.KeepAlive)
		_ = tcpConn.SetNoDelay(true)
	}
}

type connWrap struct {
	conn   net.Conn
	closed atomic.Bool
}

func (cn *connWrap) write(p []byte, timeout time.Duration) (int, error) {
	err := cn.conn.SetWriteDeadline(time.Now().Add(timeout))

	if err != nil {
		return 0, err
	}

	return cn.conn.Write(p)
}

func (cn *connWrap) isClosed() bool {
	return cn != nil && cn.closed.Load()
}

func (cn *connWrap) close() error {
	if cn != nil && cn.closed.CompareAndSwap(false, true) {
		return cn.conn.Close()
	}

	return nil
}
