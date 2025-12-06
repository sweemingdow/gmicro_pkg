package hrpc

import (
	"fmt"
	"github.com/lesismal/arpc"
	"gmicro_pkg/external/call/crpc/cauth"
	"gmicro_pkg/pkg/app"
	"gmicro_pkg/pkg/mylog"
	"gmicro_pkg/pkg/server/srpc"
	"gmicro_pkg/pkg/server/srpc/rpccall"
)

const (
	TestModuleLogger = "testLogger"
)

type TestHandler struct {
}

func NewTestHandler() *TestHandler {
	mylog.AddModuleLogger(TestModuleLogger)

	return &TestHandler{}
}

func (th *TestHandler) HandleTest(c *arpc.Context) {
	//var req map[string]any
	//err := c.Bind(&req)
	//if err != nil {
	//	c.Write(map[string]any{
	//		"err": err.Error(),
	//	})
	//}
	//
	//_ = c.Write(map[string]any{
	//	"code":    1,
	//	"content": fmt.Sprintf("A reply from:%s:%d with:%v", app.GetTheApp().GetAppName(), app.GetTheApp().GetRpcPort(), req),
	//})

	str := ""
	if err := c.Bind(&str); err == nil {
		c.Write(fmt.Sprintf("A reply from:%s:%d with:%v", app.GetTheApp().GetAppName(), app.GetTheApp().GetRpcPort(), str))
	}
}

func (th *TestHandler) HandleAuth(c *arpc.Context) {
	var req rpccall.RpcReqWrapper[cauth.AuthReq]

	ok := srpc.BindAndWriteLoggedIfError(c, &req)
	if !ok {
		return
	}

	var resp = rpccall.Ok(cauth.AuthResp{
		Uid: "9527",
	})

	srpc.WriteLoggedIfError(c, &resp)
}
