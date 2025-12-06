package hhttp

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sweemingdow/gmicro_pkg/pkg/middleware/fibermw"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"github.com/sweemingdow/gmicro_pkg/pkg/server/srpc/rclient/rcfactory"
	"time"
)

const (
	OneTwoThreeModuleLogger = "oneTwoThreeLogger"
)

type OneTwoThreeHandler struct {
	rpcFactory rcfactory.ArpcClientFactory
}

func NewOneTwoThreeHandler(rpcFactory rcfactory.ArpcClientFactory) *OneTwoThreeHandler {
	mylog.AddModuleLogger(OneTwoThreeModuleLogger)

	return &OneTwoThreeHandler{
		rpcFactory: rpcFactory,
	}
}

func (oh *OneTwoThreeHandler) HandleOneTwo(c *fiber.Ctx) error {
	lg := fibermw.GetLoggerFromFiberCtx(c)
	lg.Debug().Msg("handle one two start")

	rpcCli := oh.rpcFactory.AcquireClient("two_service")

	req := "hello"
	rsp := ""
	err := rpcCli.Call("/two/test/12", &req, &rsp, time.Second*5)
	if err != nil {
		lg.Error().Stack().Err(err).Msg("call failed")
		return c.SendString(err.Error())
	}

	lg.Debug().Msgf("call response:%s", rsp)

	// biz logic
	return c.SendString(rsp)
}

func (oh *OneTwoThreeHandler) HandleOneTwoThree(c *fiber.Ctx) error {
	lg := fibermw.GetLoggerFromFiberCtx(c)
	lg.Debug().Msg("handle one two three start")

	// biz logic
	return c.SendString("pong")
}
