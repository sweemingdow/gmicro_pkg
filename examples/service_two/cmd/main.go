package main

import (
	"gmicro_pkg/examples/service_two/internal/handlers/hhttp"
	"gmicro_pkg/examples/service_two/internal/handlers/hrpc"
	"gmicro_pkg/examples/service_two/internal/routers"
	"gmicro_pkg/pkg/boot"
	"gmicro_pkg/pkg/routebinder"
)

func main() {
	booter := boot.NewBooter()

	booter.AddConfigStageOption(boot.WithLogger())

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
