package rhttp

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sweemingdow/gmicro_pkg/examples/service_one/internal/handlers/hhttp"
	"github.com/sweemingdow/gmicro_pkg/pkg/middleware/fibermw"
)

func ConfigureOneRouter(fa *fiber.App, handler *hhttp.OneHandler) {
	oneGrp := fa.Group("/one")
	oneGrp.
		// 将默认日志注入到fiber.Ctx
		Use(fibermw.ModuleLoggerInject(hhttp.OneModuleLogger)).
		Get("/ping", handler.HandlePing)
}
