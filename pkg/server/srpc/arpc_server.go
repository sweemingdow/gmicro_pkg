package srpc

import (
	"context"
	"fmt"
	"github.com/lesismal/arpc"
	"github.com/pkg/errors"
	"github.com/sweemingdow/gmicro_pkg/pkg/app"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"github.com/sweemingdow/gmicro_pkg/pkg/server/srpc/rpccall"
	"net"
	"strings"
	"time"
)

var didaReply = []byte{'\n'}

func DiDa(c *arpc.Context) error {
	return c.Write(didaReply)
}

const (
	arpcProcessLoggerName = "arpcProcessLogger"
)

var (
	dl *mylog.DecoLogger
)

type ArpcServer struct {
	srv  *arpc.Server
	port int
}

func NewArpcServer(port int) *ArpcServer {
	rpcSrv := arpc.NewServer()

	rpcSrv.Codec = JsonIterCodec{}
	rpcSrv.Handler.SetLogTag(fmt.Sprintf("[%s]", app.GetTheApp().GetAppName()))

	dl = mylog.NewDecoLogger(arpcProcessLoggerName)

	return &ArpcServer{
		srv:  rpcSrv,
		port: port,
	}
}

func (ars *ArpcServer) GetArpcSrv() *arpc.Server {
	return ars.srv
}

func (ars *ArpcServer) OnCreated(ec chan<- error) {
	lg := mylog.GetInitMarkerLogger()
	lg.Debug().Msgf("arpc rpc server start now, port:%d", ars.port)

	go func() {
		if err := ars.srv.Run(fmt.Sprintf(":%d", ars.port)); err != nil {
			ilg := mylog.GetInitMarkerLogger()

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
	lg := mylog.GetStopMarkLogger()
	lg.Debug().Msg("arpc rpc server stop now")

	start := time.Now()
	err := ars.srv.Shutdown(ctx)
	if err != nil {
		return err
	}

	lg.Info().Msgf("arpc rpc server stopped successfully, took:%v", time.Since(start))

	return nil
}

func ParseReq(c *arpc.Context, val any) bool {
	if err := c.Bind(val); err != nil {
		err = errors.Wrap(err, "parse rpc req failed")
		dl.Error().Stack().Err(err)

		return WriteResp(c, rpccall.NewCodecErrResp(err))
	}

	return true
}

func WriteResp(c *arpc.Context, val any) bool {
	if err := c.Write(val); err != nil {
		err = errors.Wrap(err, "write rpc resp failed")

		dl.Error().Stack().Err(err).Any("data", val).Msg("write rpc resp failed")

		return false
	}

	return true
}
