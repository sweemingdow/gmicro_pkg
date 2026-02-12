package arpcmw

import (
	"github.com/lesismal/arpc"
	"github.com/pkg/errors"
	"github.com/sweemingdow/gmicro_pkg/pkg/myerr"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"github.com/sweemingdow/gmicro_pkg/pkg/server/srpc/rpccall"
)

type ArpcBizHandler[T, R any] func(req T) (R, error)

func BindAndWrite[T, R any](handler ArpcBizHandler[T, R]) arpc.HandlerFunc {
	lg := mylog.GetLoggerWrapMarker("arpcProcessor")

	return func(c *arpc.Context) {
		// parse request
		var req T
		if err := c.Bind(&req); err != nil {
			err = errors.Wrap(err, "parse rpc req failed")
			if err = c.Write(rpccall.NewCodecErrResp(err)); err != nil {
				err = errors.Wrap(err, "write rpc resp failed ")
				lg.Error().Stack().Err(err).Send()
			}

			return
		}

		// biz handle
		resp, err := handler(req)

		// then write back rpc response
		if err == nil {
			if err = c.Write(rpccall.Ok(resp)); err != nil {
				err = errors.Wrap(err, "write rpc resp failed")
				lg.Error().Stack().Err(err).Send()
			}

			return
		}

		e, ok := myerr.AsRpcRespError(err)
		if ok {
			if err = c.Write(rpccall.ErrAll[any](e.ErrCode(), e.ErrDesc(), e.ErrMsg(), e.Data())); err != nil {
				err = errors.Wrap(err, "write rpc expected error resp failed")
				lg.Error().Stack().Err(err).Send()
			}

			return
		}

		if err = c.Write(rpccall.NewUnpredictableErrResp(err)); err != nil {
			err = errors.Wrap(err, "write rpc unexpected error resp failed")
			lg.Error().Stack().Err(err).Send()
		}
	}
}
