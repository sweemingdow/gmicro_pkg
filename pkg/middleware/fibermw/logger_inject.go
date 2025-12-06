package fibermw

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
)

type logCtxKey struct {
}

func ModuleLoggerInject(module string) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Locals(logCtxKey{}, mylog.GetLogger(module))

		return c.Next()
	}
}

func GetLoggerFromFiberCtx(c *fiber.Ctx) zerolog.Logger {
	if lg, ok := c.Locals(logCtxKey{}).(zerolog.Logger); ok {
		// 这里还可以继续添加 trace_id和span_id之类的
		return lg
	}

	// fallback
	return mylog.AppLogger()
}
