package gserver

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"gmicro_pkg/pkg/mylog"
	"gmicro_pkg/pkg/parser/json"
	"gmicro_pkg/pkg/server/shttp"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// 用于网关自身的fiber app
const (
	hsDefaultReloadShutdownTimeoutMills = 30_000
	hsDefaultIdleTimeoutMills           = 60_000
	hsDefaultReadTimeoutMills           = 11_000
	hsDefaultWriteTimeoutMills          = 11_000
	hsDefaultBodyLimit                  = 2 * 1024 * 1024
	hsDefaultConcurrency                = 256 * 1024
)

type HotLoadServerConfig struct {
	Port                       int
	ReloadShutdownTimeoutMills int
	IdleTimeoutMills           int
	ReadTimeoutMills           int
	WriteTimeoutMills          int
	BodyLimit                  int
	Concurrency                int
}

type HotLoadServer struct {
	curFa      *fiber.App
	cfg        HotLoadServerConfig
	mu         sync.Mutex
	lh         *hotLoadHandler
	closed     atomic.Bool
	errHandler fiber.ErrorHandler
}

func NewHotLoadServer(ec chan<- error, cfg HotLoadServerConfig, path2handler map[string]fiber.Handler, errHandler fiber.ErrorHandler) *HotLoadServer {
	hs := &HotLoadServer{
		cfg:        cfg,
		errHandler: errHandler,
	}

	hs.mu.Lock()
	defer hs.mu.Unlock()
	fa := hs.createFiber()
	hs.curFa = fa

	for path, handler := range path2handler {
		fa.All(path, handler)
	}

	hs.lh = newHotLoadHandler(fa.Handler())

	go func() {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
		if err != nil {
			ec <- err
			return
		}

		server := &fasthttp.Server{
			Handler: hs.lh.serve,
		}

		if err = server.Serve(ln); err != nil {
			if hs.closed.Load() && err.Error() == shttp.UseClosedConnErrDesc {
				return
			}

			ec <- err
		}
	}()

	return hs
}

func (hs *HotLoadServer) Reload(path2handler map[string]fiber.Handler) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	oldFa := hs.curFa

	newFa := hs.createFiber()
	hs.curFa = newFa

	for path, handler := range path2handler {
		newFa.All(path, handler)
	}

	hs.lh.swap(newFa.Handler())

	if oldFa != nil {
		shutdownFiberAsync(oldFa, hs.cfg.ReloadShutdownTimeoutMills)
	}
}

func (hs *HotLoadServer) Shutdown(ctx context.Context) error {
	if !hs.closed.CompareAndSwap(false, true) {
		return nil
	}

	hs.mu.Lock()
	defer hs.mu.Unlock()

	if hs.curFa != nil {
		return hs.curFa.ShutdownWithContext(ctx)
	}

	return nil
}

func (hs *HotLoadServer) createFiber() *fiber.App {
	cfg := hs.cfg
	var idleTimeoutMills = cfg.IdleTimeoutMills
	if idleTimeoutMills == 0 {
		idleTimeoutMills = hsDefaultIdleTimeoutMills
	}

	var readTimeoutMills = cfg.ReadTimeoutMills
	if readTimeoutMills == 0 {
		readTimeoutMills = hsDefaultReadTimeoutMills
	}

	var writeTimeoutMills = cfg.WriteTimeoutMills
	if writeTimeoutMills == 0 {
		writeTimeoutMills = hsDefaultWriteTimeoutMills
	}

	var bodyLimit = cfg.BodyLimit
	if bodyLimit == 0 {
		bodyLimit = hsDefaultBodyLimit
	}

	var concurrency = cfg.Concurrency
	if concurrency == 0 {
		concurrency = hsDefaultConcurrency
	}

	return fiber.New(fiber.Config{
		IdleTimeout:  time.Duration(idleTimeoutMills) * time.Millisecond,
		ReadTimeout:  time.Duration(readTimeoutMills) * time.Millisecond,
		WriteTimeout: time.Duration(writeTimeoutMills) * time.Millisecond,
		BodyLimit:    bodyLimit,
		Concurrency:  concurrency,
		JSONEncoder:  json.Fmt,
		JSONDecoder:  json.Parse,
		ErrorHandler: hs.errHandler,
	})
}

func shutdownFiberAsync(fa *fiber.App, timeoutMills int) {
	if timeoutMills == 0 {
		timeoutMills = hsDefaultReloadShutdownTimeoutMills
	}

	go func() {
		lg := mylog.AppLoggerWithStop()

		defer func() {
			if r := recover(); r != nil {
				lg.Error().Stack().Err(fmt.Errorf("%v", r)).Msg("shutdown last fiber panic")
			}
		}()

		if err := fa.ShutdownWithTimeout(time.Duration(timeoutMills) * time.Millisecond); err != nil {
			lg.Error().Stack().Err(err).Msg("shutdown last fiber failed")
		}
	}()

}

type hotLoadHandler struct {
	handlerAtm atomic.Value
}

func newHotLoadHandler(handler fasthttp.RequestHandler) *hotLoadHandler {
	lh := &hotLoadHandler{}
	lh.handlerAtm.Store(handler)
	return lh
}

func (lh *hotLoadHandler) swap(newHandler fasthttp.RequestHandler) {
	lh.handlerAtm.Store(newHandler)
}

func (lh *hotLoadHandler) serve(c *fasthttp.RequestCtx) {
	h := lh.handlerAtm.Load()
	if h != nil {
		h.(fasthttp.RequestHandler)(c)
	} else {
		c.Error("Unavailable:Not ready", fasthttp.StatusServiceUnavailable)
	}
}
