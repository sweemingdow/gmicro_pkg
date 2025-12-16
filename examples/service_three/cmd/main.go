package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sweemingdow/gmicro_pkg/examples/service_three/internal/routers"
	"github.com/sweemingdow/gmicro_pkg/pkg/boot"
	"github.com/sweemingdow/gmicro_pkg/pkg/routebinder"
	"github.com/sweemingdow/gmicro_pkg/pkg/server/srpc/rclient"
)

func main() {
	booter := boot.NewBooter()

	booter.AddConfigStageOption(boot.WithLogger(nil))

	booter.AddComponentStageOption(boot.WithNacosClient())

	booter.AddComponentStageOption(boot.WithNacosRegistry())

	booter.AddComponentStageOption(boot.WithRpcClientFactory(rclient.NewRoundRobinLoadBalancer()))

	booter.AddServerOption(boot.WithHttpServer(func(c *fiber.Ctx, err error) error {
		return err
	}))

	booter.StartAndServe(func(ac *boot.AppContext) routebinder.AppRouterBinder {
		// 加载路由配置
		return routers.NewThreeServiceRouteBinder()
	})
}
