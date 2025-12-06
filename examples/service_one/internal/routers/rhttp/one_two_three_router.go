package rhttp

import (
	"github.com/gofiber/fiber/v2"
	"gmicro_pkg/examples/service_one/internal/handlers/hhttp"
	"gmicro_pkg/pkg/middleware/fibermw"
)

func ConfigureOneTwoThreeRouter(fa *fiber.App, handler *hhttp.OneTwoThreeHandler) {
	oneTwoThreeGrp := fa.Group("/one_two_three")
	oneTwoThreeGrp.
		// 将默认日志注入到fiber.Ctx
		Use(fibermw.ModuleLoggerInject(hhttp.OneTwoThreeModuleLogger)).
		Get("/12", handler.HandleOneTwo).
		Get("/123", handler.HandleOneTwoThree)
}
