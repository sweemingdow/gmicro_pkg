package rclient

import (
	"fmt"
	"github.com/lesismal/arpc"
	"github.com/sweemingdow/gmicro_pkg/pkg/app"
	"github.com/sweemingdow/gmicro_pkg/pkg/regdis"
	"github.com/sweemingdow/gmicro_pkg/pkg/server/srpc"
	"net"
	"sync/atomic"
	"time"
)

type arpcClientWrap struct {
	ins    *regdis.Instance
	cli    *arpc.Client
	weight float64
	closed atomic.Bool
}

func newArpcClientWrap(ins *regdis.Instance, timeout time.Duration) (*arpcClientWrap, error) {
	cli, err := arpc.NewClient(func() (net.Conn, error) {
		return net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ins.Ip, ins.Port), timeout)
	})

	if err != nil {
		return nil, err
	}

	cli.Codec = srpc.JsonIterCodec{}
	cli.Handler.SetLogTag(fmt.Sprintf("[%s]", app.GetTheApp().GetAppName()))

	return &arpcClientWrap{
		ins:    ins,
		cli:    cli,
		weight: ins.Weight,
	}, nil
}

func (acw *arpcClientWrap) close() {
	if acw.closed.CompareAndSwap(false, true) {
		acw.cli.Stop()
	}
}
