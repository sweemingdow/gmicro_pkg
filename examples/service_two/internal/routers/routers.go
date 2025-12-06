package routers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lesismal/arpc"
	"github.com/sweemingdow/gmicro_pkg/examples/service_two/internal/handlers/hhttp"
	"github.com/sweemingdow/gmicro_pkg/examples/service_two/internal/handlers/hrpc"
	"github.com/sweemingdow/gmicro_pkg/pkg/app"
	"github.com/sweemingdow/gmicro_pkg/pkg/routebinder"
)

type twoServiceRouteBinder struct {
	testHandler    *hrpc.TestHandler
	apiTestHandler *hhttp.ApiTestHandler
}

func NewTwoServiceRouteBinder(testHandler *hrpc.TestHandler, apiTestHandler *hhttp.ApiTestHandler) routebinder.AppRouterBinder {
	return &twoServiceRouteBinder{
		testHandler:    testHandler,
		apiTestHandler: apiTestHandler,
	}
}

func (rb *twoServiceRouteBinder) BindFiber(fa *fiber.App) {
	fa.Get("/api/three/test", rb.apiTestHandler.HandleApiTestV1)
	fa.Get("/test/v2", func(c *fiber.Ctx) error {
		return c.SendString(fmt.Sprintf("ok for:[/test/v2] with port:%d", app.GetTheApp().GetHttpPort()))
	})
}

func (rb *twoServiceRouteBinder) BindArpc(srv *arpc.Server) {
	srv.Handler.Handle("/two/test/12", rb.testHandler.HandleTest)
	srv.Handler.Handle("/auth", rb.testHandler.HandleAuth)
}
