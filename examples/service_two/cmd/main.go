package main

import (
	"github.com/sweemingdow/gmicro_pkg/examples/service_two/internal/handlers/hhttp"
	"github.com/sweemingdow/gmicro_pkg/examples/service_two/internal/handlers/hrpc"
	"github.com/sweemingdow/gmicro_pkg/examples/service_two/internal/routers"
	"github.com/sweemingdow/gmicro_pkg/pkg/boot"
	"github.com/sweemingdow/gmicro_pkg/pkg/routebinder"
)

func main() {
	booter := boot.NewBooter()

	booter.AddConfigStageOption(boot.WithLogger(nil))

	booter.AddComponentStageOption(boot.WithNacosClient())

	booter.AddServerOption(boot.WithNacosRegistry())

	// 启动rpc服务
	booter.AddServerOption(boot.WithHttpServer(nil))

	// 启动rpc服务
	booter.AddServerOption(boot.WithRpcServer())

	booter.StartAndServe(func(ac *boot.AppContext) routebinder.AppRouterBinder {
		return routers.NewTwoServiceRouteBinder(hrpc.NewTestHandler(), hhttp.NewApiTestHandler())
	})
}
