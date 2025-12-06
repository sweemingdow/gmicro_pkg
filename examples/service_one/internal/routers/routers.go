package routers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lesismal/arpc"
	"gmicro_pkg/examples/service_one/internal/handlers/hhttp"
	"gmicro_pkg/examples/service_one/internal/routers/rhttp"
	"gmicro_pkg/pkg/app"
	"gmicro_pkg/pkg/routebinder"
)

type oneServiceRouteBinder struct {
	oneHandler *hhttp.OneHandler
	ottHandler *hhttp.OneTwoThreeHandler
}

func NewOneServiceRouteBinder(
	oneHandler *hhttp.OneHandler,
	ottHandler *hhttp.OneTwoThreeHandler,
) routebinder.AppRouterBinder {
	return &oneServiceRouteBinder{
		oneHandler: oneHandler,
		ottHandler: ottHandler,
	}
}

func (rb *oneServiceRouteBinder) BindFiber(fa *fiber.App) {
	//fa.Use(fibermw.RequestTrace())

	fa.Get("/test/v1", func(c *fiber.Ctx) error {
		//lg := mylog.AppLogger()
		//lg.Debug().Msgf("header val:%+v", c.GetReqHeaders()["Info-From-Gateway"][0])
		return c.SendString(fmt.Sprintf("ok for:[/test/v1] with port:%d", app.GetTheApp().GetHttpPort()))
	})

	rhttp.ConfigureOneRouter(fa, rb.oneHandler)

	rhttp.ConfigureOneTwoThreeRouter(fa, rb.ottHandler)
}

func (rb *oneServiceRouteBinder) BindArpc(srv *arpc.Server) {

}
