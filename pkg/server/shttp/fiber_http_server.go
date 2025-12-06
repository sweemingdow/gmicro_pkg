package shttp

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"github.com/sweemingdow/gmicro_pkg/pkg/parser/json"
	"sync/atomic"
	"time"
)

var FiberStoppedTimeout = errors.New("fiber http server stopped timeout, continuing graceful exit for other components")

const UseClosedConnErrDesc = "use of closed network connection"

// 用于各个服务的fiber app
const (
	defaultIdleTimeoutMills  = 60_000
	defaultReadTimeoutMills  = 10_000
	defaultWriteTimeoutMills = 10_000
	defaultBodyLimit         = 2 * 1024 * 1024
)

type FiberServerConfig struct {
	Port              int
	IdleTimeoutMills  int
	ReadTimeoutMills  int
	WriteTimeoutMills int
	BodyLimit         int
}

func DefaultFiberServerConfig(port int) FiberServerConfig {
	return FiberServerConfig{
		Port: port,
	}
}

type FiberHttpServer struct {
	fa     *fiber.App
	port   int
	closed atomic.Bool
}

type HttpRouterBind func(fa *fiber.App)

func NewFiberHttpServer(cfg FiberServerConfig, errHandler fiber.ErrorHandler) *FiberHttpServer {
	var idleTimeoutMills = cfg.IdleTimeoutMills
	if idleTimeoutMills == 0 {
		idleTimeoutMills = defaultIdleTimeoutMills
	}

	var readTimeoutMills = cfg.ReadTimeoutMills
	if readTimeoutMills == 0 {
		readTimeoutMills = defaultReadTimeoutMills
	}

	var writeTimeoutMills = cfg.WriteTimeoutMills
	if writeTimeoutMills == 0 {
		writeTimeoutMills = defaultWriteTimeoutMills
	}

	var bodyLimit = cfg.BodyLimit
	if bodyLimit == 0 {
		bodyLimit = defaultBodyLimit
	}

	fa := fiber.New(fiber.Config{
		IdleTimeout:  time.Duration(idleTimeoutMills) * time.Millisecond,
		ReadTimeout:  time.Duration(readTimeoutMills) * time.Millisecond,
		WriteTimeout: time.Duration(writeTimeoutMills) * time.Millisecond,
		BodyLimit:    bodyLimit,
		ErrorHandler: errHandler,
		JSONEncoder:  json.Fmt,
		JSONDecoder:  json.Parse,
	})

	return &FiberHttpServer{
		port: cfg.Port,
		fa:   fa,
	}
}

func (fhs *FiberHttpServer) GetFiber() *fiber.App {
	return fhs.fa
}

func (fhs *FiberHttpServer) OnCreated(ec chan<- error) {
	lg := mylog.AppLoggerWithInit()
	lg.Debug().Msgf("fiber http server start now, port:%d", fhs.port)

	go func() {
		if err := fhs.fa.Listen(fmt.Sprintf(":%d", fhs.port)); err != nil {
			if fhs.closed.Load() && err.Error() == UseClosedConnErrDesc {
				return
			}

			ec <- err
		}
	}()
}

func (fhs *FiberHttpServer) OnDispose(ctx context.Context) error {
	if !fhs.closed.CompareAndSwap(false, true) {
		return nil
	}

	lg := mylog.AppLoggerWithStop()
	lg.Debug().Msg("fiber http server stop now")

	// 解决fiber停机经常超时, 卡住其他资源无法释放问题
	var (
		bufferTime time.Duration
		shortCtx   = ctx
		cancel     = func() {}
	)

	dl, ok := ctx.Deadline()
	if ok {
		// 计算剩余时间
		remain := time.Until(dl)

		bufferTime = remain / 2

		if bufferTime > 0 && (remain-bufferTime) > time.Millisecond {
			shorterTimeout := remain - bufferTime
			newCtx, newCancel := context.WithTimeout(ctx, shorterTimeout)
			shortCtx = newCtx
			cancel = newCancel
		}
	}
	defer cancel()

	start := time.Now()
	err := fhs.fa.ShutdownWithContext(shortCtx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			lg.Warn().Msgf("fiber http server stopped timeout, took:%v", time.Since(start))

			return FiberStoppedTimeout
		}

		return err
	}

	lg.Info().Msgf("fiber http server stopped successfully, took:%v", time.Since(start))

	return nil
}
