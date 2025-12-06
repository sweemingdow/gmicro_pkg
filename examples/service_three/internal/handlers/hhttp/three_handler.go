package hhttp

import (
	"github.com/gofiber/fiber/v2"
	"gmicro_pkg/pkg/middleware/fibermw"
	"gmicro_pkg/pkg/mylog"
)

const (
	OneModuleLogger = "oneLogger"
)

type ThreeHandler struct {
}

func NewThreeHandler() *ThreeHandler {
	mylog.AddModuleLogger(OneModuleLogger)

	return &ThreeHandler{}
}

func (oh *ThreeHandler) HandlePing(c *fiber.Ctx) error {
	lg := fibermw.GetLoggerFromFiberCtx(c)
	lg.Debug().Msg("handle ping start")

	// biz logic
	return c.SendString("pong")
}
