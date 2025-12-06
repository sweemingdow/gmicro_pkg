package hhttp

import (
	"github.com/gofiber/fiber/v2"
	"gmicro_pkg/pkg/middleware/fibermw"
	"gmicro_pkg/pkg/mylog"
)

const (
	OneModuleLogger = "oneLogger"
)

type OneHandler struct {
}

func NewOneHandler() *OneHandler {
	mylog.AddModuleLogger(OneModuleLogger)

	return &OneHandler{}
}

func (oh *OneHandler) HandlePing(c *fiber.Ctx) error {
	lg := fibermw.GetLoggerFromFiberCtx(c)
	lg.Debug().Msg("handle ping start")

	// biz logic
	return c.SendString("pong")
}
