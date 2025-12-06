package srpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/lesismal/arpc"
	"gmicro_pkg/pkg/app"
	"gmicro_pkg/pkg/myerr"
	"gmicro_pkg/pkg/mylog"
	"gmicro_pkg/pkg/server/srpc/rpccall"
	"net"
	"strings"
	"time"
)

var didaReply = []byte{'\n'}

func DiDa(c *arpc.Context) error {
	return c.Write(didaReply)
}

type ArpcServer struct {
	srv  *arpc.Server
	port int
}

func NewArpcServer(port int) *ArpcServer {
	mylog.AddModuleLoggerWithFrame("arpcServer", 3)

	rpcSrv := arpc.NewServer()

	rpcSrv.Codec = JsonIterCodec{}
	rpcSrv.Handler.SetLogTag(fmt.Sprintf("[%s]", app.GetTheApp().GetAppName()))

	return &ArpcServer{
		srv:  rpcSrv,
		port: port,
	}
}

func (ars *ArpcServer) GetArpcSrv() *arpc.Server {
	return ars.srv
}

func (ars *ArpcServer) OnCreated(ec chan<- error) {
	lg := mylog.AppLogger()
	lg.Debug().Msgf("arpc rpc server start now, port:%d", ars.port)

	go func() {
		if err := ars.srv.Run(fmt.Sprintf(":%d", ars.port)); err != nil {
			ilg := mylog.AppLogger()

			if strings.Contains(err.Error(), "use of closed network connection") {
				ilg.Debug().Msg("arpc rpc server run completed")
				return
			}

			var ope *net.OpError
			if errors.As(err, &ope) {
				ec <- err
				return
			}

			ilg.Error().Stack().Err(err).Msg("arpc rpc server run unexpected error")
		}
	}()
}

func (ars *ArpcServer) OnDispose(ctx context.Context) error {
	lg := mylog.AppLogger()
	lg.Debug().Msg("arpc rpc server stop now")

	start := time.Now()
	err := ars.srv.Shutdown(ctx)
	if err != nil {
		return err
	}

	lg.Info().Msgf("arpc rpc server stopped successfully, took:%v", time.Since(start))

	return nil
}

func BindAndWriteLoggedIfError(c *arpc.Context, val any) bool {
	if err := c.Bind(val); err != nil {
		err = myerr.NewRpcBindError(err)

		lg := mylog.AppLoggerWithBind()
		lg.Error().Stack().Err(err).Msg("bind data failed")

		return WriteLoggedIfError(c, rpccall.SimpleUnpredictableErr(err))
	}

	return true
}

func WriteLoggedIfError(c *arpc.Context, val any) bool {
	if err := c.Write(val); err != nil {
		lg := mylog.AppLoggerWithWriteBack()

		lg.Error().Stack().Err(err).Any("data", val).Msg("write back data failed")

		return false
	}

	return true
}
