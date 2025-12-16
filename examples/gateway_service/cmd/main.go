package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sweemingdow/gmicro_pkg/examples/gateway_service/internal/config/gwncfg"
	"github.com/sweemingdow/gmicro_pkg/examples/gateway_service/internal/gmiddleware"
	"github.com/sweemingdow/gmicro_pkg/external/call/crpc/cauth"
	"github.com/sweemingdow/gmicro_pkg/gateway/pkg/gserver"
	"github.com/sweemingdow/gmicro_pkg/pkg/app"
	"github.com/sweemingdow/gmicro_pkg/pkg/boot"
	"github.com/sweemingdow/gmicro_pkg/pkg/regdis/disnacos"
	"github.com/sweemingdow/gmicro_pkg/pkg/regdis/extra/enacos"
	"github.com/sweemingdow/gmicro_pkg/pkg/routebinder"
)

func main() {
	booter := boot.NewBooter()

	booter.AddConfigStageOption(boot.WithLogger(nil))

	booter.AddComponentStageOption(boot.WithNacosClient())

	booter.AddComponentStageOption(boot.WithNacosConfig(gwncfg.NewGatewayConfigurationReceiver()))

	booter.AddComponentStageOption(boot.WithNacosRegistry())

	booter.AddComponentStageOption(boot.WithRpcClientFactory(nil))

	booter.StartAndServe(func(ac *boot.AppContext) routebinder.AppRouterBinder {
		receiver := ac.GetConfigureReceiver()
		//val, _ := receiver.RecentlyConfigure(dnacos.StaticConfigName)
		//sc := val.(gwncfg.GatewayStaticConfig)

		var tableCfg gserver.RouterTableConfig
		tableCfgVal, ok := receiver.RecentlyConfigure(gwncfg.GatewayRouterTableConfigName)
		if ok {
			tableCfg = tableCfgVal.(gserver.RouterTableConfig)
		}

		tables := gserver.Cfg2routerItems(tableCfg)

		ta := app.GetTheApp()
		cfg := ta.GetConfig()

		gserver.NewConfigurableGatewayServer(
			tables,
			gserver.HotLoadServerConfig{
				Port:                       ta.GetHttpPort(),
				ReloadShutdownTimeoutMills: 15_000,
				IdleTimeoutMills:           60_000,
				ReadTimeoutMills:           45_000,
				WriteTimeoutMills:          45_000,
			},
			ac.GetEc(),
			disnacos.NewNacosDiscovery(ac.GetNacosClient().GetNamingClient()),
			enacos.PkgDiscoveryExtraParam(cfg.NacosCenterCfg.RegistryDiscoverCfg.ClusterName, cfg.NacosCenterCfg.RegistryDiscoverCfg.GroupName),
			[]fiber.Handler{gmiddleware.AuthWithRpc(cauth.NewAuthRpcProvider(ac.GetArpcClientFactory()))},
			[]fiber.Handler{gmiddleware.RespAttach()},
			gmiddleware.RespInterceptWhenError(),
		)

		return nil
	})
}
