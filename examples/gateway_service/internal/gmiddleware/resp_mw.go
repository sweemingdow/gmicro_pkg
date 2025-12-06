package gmiddleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"gmicro_pkg/gateway/pkg/gserver"
	"gmicro_pkg/pkg/app"
	"gmicro_pkg/pkg/myerr"
	"gmicro_pkg/pkg/server/srpc/rpccall"
)

func RespAttach() fiber.Handler {
	return func(c *fiber.Ctx) error {
		mi := gserver.GetMetaInfoFromCtx(c)
		c.Response().Header.Add("X-Req-ID", mi.ReqId)

		return nil
	}
}

func RespInterceptWhenError() fiber.ErrorHandler {
	debug := app.GetTheApp().IsDevProfile()

	return func(c *fiber.Ctx, err error) error {
		rce, ok := myerr.DecodeRpcCallErr(err)
		if ok {
			if rce.Timeout() {
				return c.SendStatus(fasthttp.StatusGatewayTimeout)
			}

			if debug {
				c.Status(fasthttp.StatusBadGateway)
				return c.SendString(err.Error())
			}

			return c.SendStatus(fasthttp.StatusBadGateway)
		}

		rrp, ok := myerr.DecodeRpcRespError(err)
		if ok {
			var statusCode int
			if rrp.Code() == rpccall.ServerUnpredictableErr {
				statusCode = fasthttp.StatusBadGateway
			} else {
				statusCode = fasthttp.StatusUnauthorized
			}

			if debug {
				c.Status(statusCode)
				return c.SendString(err.Error())
			}

			return c.SendStatus(statusCode)
		}

		c.Status(fasthttp.StatusInternalServerError)
		if debug {
			return c.SendString(err.Error())
		}

		return c.SendStatus(fasthttp.StatusInternalServerError)
	}
}
