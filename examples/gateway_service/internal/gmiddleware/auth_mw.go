package gmiddleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sweemingdow/gmicro_pkg/external/call/crpc/cauth"
	"github.com/sweemingdow/gmicro_pkg/gateway/pkg/gserver"
	"github.com/sweemingdow/gmicro_pkg/pkg/myerr"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"github.com/sweemingdow/gmicro_pkg/pkg/parser/json"
	"github.com/sweemingdow/gmicro_pkg/pkg/server/srpc/rpccall"
)

func AuthWithRpc(provider cauth.AuthRpcProvider) fiber.Handler {

	return func(c *fiber.Ctx) error {
		lg := mylog.AppLoggerWithRpc()

		mi := gserver.GetMetaInfoFromCtx(c)

		req := cauth.AuthReq{
			Token: "woshidage",
		}

		resp, err := provider.Auth(rpccall.CreateIdReq(mi.ReqId, req))
		if err != nil {
			err = myerr.NewRpcCallError(err)
			lg.Error().Stack().Err(err).Str("req_id", mi.ReqId).Msg("call auth server failed")
			return err
		}

		val, err := resp.OkOrErr()
		if err != nil {
			return err
		}

		values, err := json.Fmt(&val)
		if err != nil {
			return err
		}

		// 写入鉴权信息到upstream server
		c.Request().Header.Add("Info-From-Gateway", string(values))

		return nil
	}
}
