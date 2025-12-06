package fibermw

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"time"
)

func RequestTrace() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		defer func() {
			lg := mylog.MonitorLogger()
			lg.Debug().
				Str("method", c.Method()).
				Str("path", c.Path()).
				Dur("took", time.Since(start)).
				Send()
		}()

		return c.Next()
	}
}
