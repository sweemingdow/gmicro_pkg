package fibermw

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"github.com/sweemingdow/gmicro_pkg/pkg/response"
	"github.com/sweemingdow/gmicro_pkg/pkg/response/apiresp"
)

type FiberBizHandler[T, R any] func(req T) (R, error)

func getLogger() zerolog.Logger {
	return mylog.GetLoggerWrapMarker("fiberProcessor")
}

func BindAndWrite[T, R any](handler FiberBizHandler[T, R]) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req T
		if err := c.BodyParser(&req); err != nil {
			err = errors.Wrap(err, "parse api req failed")
			lg := getLogger()
			lg.Error().Stack().Err(err).Send()
			return c.JSON(apiresp.CodeMsgResp(response.CodecError, response.ApiCode2text(response.RpcCodecError)))
		}

		resp, err := handler(req)

		if err == nil {
			return c.JSON(apiresp.Ok(resp))
		}

		// give to fiber server error handler to process
		return err
	}
}

func CustomAndWrite[R any](handler FiberBizHandler[*fiber.Ctx, R]) fiber.Handler {
	return func(c *fiber.Ctx) error {
		resp, err := handler(c)

		if err == nil {
			return c.JSON(apiresp.Ok(resp))
		}

		// give to fiber server error handler to process
		return err
	}
}

func QueryAndWrite[R any](handler FiberBizHandler[map[string]string, R]) fiber.Handler {
	return func(c *fiber.Ctx) error {
		resp, err := handler(c.Queries())

		if err == nil {
			return c.JSON(apiresp.Ok(resp))
		}

		// give to fiber server error handler to process
		return err
	}
}
