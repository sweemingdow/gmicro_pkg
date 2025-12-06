package main

import (
	"github.com/gofiber/fiber/v2"
	"gmicro_pkg/examples/service_one/internal/handlers/hhttp"
	"gmicro_pkg/examples/service_one/internal/routers"
	"gmicro_pkg/pkg/boot"
	"gmicro_pkg/pkg/mylog"
	"gmicro_pkg/pkg/routebinder"
	"gmicro_pkg/pkg/server/srpc/rclient"
)

func main() {
	booter := boot.NewBooter()

	booter.AddConfigStageOption(boot.WithLogger())

	booter.AddComponentStageOption(boot.WithNacosClient())

	booter.AddComponentStageOption(boot.WithNacosRegistry())

	booter.AddComponentStageOption(boot.WithRpcClientFactory(rclient.NewRoundRobinLoadBalancer()))

	booter.AddServerOption(boot.WithHttpServer(func(c *fiber.Ctx, err error) error {
		lg := mylog.AppLogger()
		lg.Info().Msg(err.Error())
		return err
	}))

	booter.StartAndServe(func(ac *boot.AppContext) routebinder.AppRouterBinder {
		// maybe wire init

		binder := routers.NewOneServiceRouteBinder(hhttp.NewOneHandler(), hhttp.NewOneTwoThreeHandler(ac.GetArpcClientFactory()))

		return binder
	})
}

/*package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"gmicro_pkg/pkg/app"
	"gmicro_pkg/pkg/component/cnacos"
	"gmicro_pkg/pkg/decorate/logdeco"
	"gmicro_pkg/pkg/decorate/nacosdeco"
	"gmicro_pkg/pkg/graceful"
	"gmicro_pkg/pkg/lifetime"
	"gmicro_pkg/pkg/mylog"
	"gmicro_pkg/pkg/parser/cmdparser"
	"gmicro_pkg/pkg/regdis/nacosreg"
	"gmicro_pkg/pkg/server/httpsrv"
	"log"
	"time"
)

func main() {
	ec := make(chan error, 2)
	finalizer := lifetime.NewFinalizer(ec)

	cp := cmdparser.NewCmdParser()
	cp.Parse(cmdparser.DefaultParseEntry)

	ta := app.NewApp(cp)

	remoteWriter := mylog.InitLogger(ta.GetConfig().LogCfg, ta.IsDevProfile(), "one_service")
	finalizer.Collect("log_writer", logdeco.NewLogRemoteWriter(remoteWriter))

	lg := app.GetAppLogger()

	lg.Info().Msgf("application is starting, app:%v", ta)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ta.GetConfig().AppCfg.GracefulExitTimeoutMills)*time.Millisecond)
		defer cancel()

		errs, aborted := finalizer.Release(ctx)

		log.Printf("app finalizer release completed, errs:%+v, aborted:%t\n", errs, aborted)

		time.Sleep(10 * time.Millisecond)
	}()

	nc, err := cnacos.NewNacosClient(nacosdeco.ToNacosCfg(ta.GetConfig().NacosCfg))
	if err != nil {
		panic(err)
	}

	finalizer.Collect("cnacos", nc)

	autoRegistry := nacosdeco.NewNacosAutoRegistry(nacosreg.NewNacosRegistry(nc.GetNamingClient()), ta.GetConfig().NacosCenterCfg.Registry)

	finalizer.Collect("nacos_registry", autoRegistry)

	fhs := httpsrv.NewFiberHttpServer(cp.GetInt("http_port"), func(fa *fiber.App) {
		fa.Get("/ping", func(ctx *fiber.Ctx) error {
			return ctx.SendString("pong")
		})
	})

	finalizer.Collect("http_server", fhs)

	graceful.ListenExitSignal(ec)

	exitErr := <-ec

	lg.Error().Stack().Err(exitErr).Msg("exit now")
}
*/
