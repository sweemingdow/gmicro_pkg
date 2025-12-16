package boot

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/sweemingdow/gmicro_pkg/pkg/app"
	"github.com/sweemingdow/gmicro_pkg/pkg/cfgcenter/cfgnacos"
	"github.com/sweemingdow/gmicro_pkg/pkg/component/cnacos"
	"github.com/sweemingdow/gmicro_pkg/pkg/decorate/dlog"
	"github.com/sweemingdow/gmicro_pkg/pkg/decorate/dnacos"
	"github.com/sweemingdow/gmicro_pkg/pkg/graceful"
	"github.com/sweemingdow/gmicro_pkg/pkg/lifetime"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"github.com/sweemingdow/gmicro_pkg/pkg/parser/cmd"
	"github.com/sweemingdow/gmicro_pkg/pkg/regdis/disnacos"
	"github.com/sweemingdow/gmicro_pkg/pkg/regdis/regnacos"
	"github.com/sweemingdow/gmicro_pkg/pkg/routebinder"
	"github.com/sweemingdow/gmicro_pkg/pkg/server/shttp"
	"github.com/sweemingdow/gmicro_pkg/pkg/server/srpc"
	"github.com/sweemingdow/gmicro_pkg/pkg/server/srpc/rclient"
	"github.com/sweemingdow/gmicro_pkg/pkg/server/srpc/rclient/rcfactory"
	"log"
	"sync"
	"time"
)

type AppContext struct {
	finalizer *lifetime.AppFinalizer

	exitChan chan error

	nacosClient *cnacos.NacosClient

	arpcClientFactory rcfactory.ArpcClientFactory

	cmdParser *cmd.CmdParser

	routeBinder routebinder.AppRouterBinder

	configureReceiver dnacos.ConfigurationReceiver

	preHooks []ShutdownHook

	postHooks []ShutdownHook

	mu sync.Mutex

	extraStore map[string]any
}

type ShutdownHook func(ac *AppContext, ctx context.Context) error

//func (ac *AppContext) GetFinalizer() *lifetime.AppFinalizer {
//	return ac.finalizer
//}

func (ac *AppContext) GetEc() chan<- error {
	return ac.exitChan
}

func (ac *AppContext) GetNacosClient() *cnacos.NacosClient {
	return ac.nacosClient
}

func (ac *AppContext) GetArpcClientFactory() rcfactory.ArpcClientFactory {
	return ac.arpcClientFactory
}

func (ac *AppContext) GetConfigureReceiver() dnacos.ConfigurationReceiver {
	return ac.configureReceiver
}

func (ac *AppContext) StoreExtra(tag string, val any) {
	ac.extraStore[tag] = val
}

func (ac *AppContext) GetStored(tag string) any {
	return ac.extraStore[tag]
}

func (ac *AppContext) CollectLifecycle(tag string, lc lifetime.LifeCycle) {
	ac.finalizer.Collect(tag, lc)
}

// 可选功能的初始化参数签名
type AppOption func(ac *AppContext) error

// boot的3个阶段, 有强制的依赖顺序
type Booter struct {
	configStageOptions []AppOption // 配置阶段: 日志初始化等

	componentStageOptions []AppOption // 组件阶段：Nacos Client, Config, Registry, DB Client

	serverStageOptions []AppOption // 服务阶段：HTTP/RPC Servers,
}

func NewBooter() *Booter {
	return &Booter{}
}

func (b *Booter) AddConfigStageOption(opt AppOption) {
	b.configStageOptions = append(b.configStageOptions, opt)
}

func (b *Booter) AddComponentStageOption(opt AppOption) {
	b.componentStageOptions = append(b.componentStageOptions, opt)
}

func (b *Booter) AddServerOption(opt AppOption) {
	b.serverStageOptions = append(b.serverStageOptions, opt)
}

type ReadyForRouterMount func(ac *AppContext) (routebinder.AppRouterBinder, error)

func (b *Booter) StartAndServe(ready ReadyForRouterMount) {
	ec := make(chan error, 2)
	finalizer := lifetime.NewFinalizer(ec)

	// 解析命令行
	cp := cmd.NewCmdParser()
	cp.Parse(cmd.DefaultParseEntry)

	// 初始化app
	ta := app.NewApp(cp)

	ac := &AppContext{
		finalizer:  finalizer,
		exitChan:   ec,
		extraStore: make(map[string]any),
	}

	ac.cmdParser = cp

	// 先执行: ConfigStage
	if err := b.stageRun("Config", ac, b.configStageOptions); err != nil {
		log.Fatal(err)
	}

	lg := mylog.AppLoggerWithInit()

	lg.Debug().Msgf("application is starting, app:%v", ta)

	lg.Debug().Msg("config stage completed")

	// 在执行: ComponentStage
	if err := b.stageRun("Component", ac, b.componentStageOptions); err != nil {
		ec <- err
		return
	}

	lg.Debug().Msg("component stage completed")

	if ready != nil {
		router, err := ready(ac)
		if err != nil {
			ec <- err
			return
		}

		ac.routeBinder = router
	}

	// 最后执行: ServerStage
	if err := b.stageRun("Server", ac, b.serverStageOptions); err != nil {
		ec <- err
		return
	}

	lg.Debug().Msg("server stage completed")

	graceful.ListenExitSignal(ec)

	// blocking until receive exit error signal
	exitErr := <-ec

	b.shutdown(ac, exitErr)
}

// stageRun 统一执行某个阶段的所有 Option
func (b *Booter) stageRun(name string, ctx *AppContext, options []AppOption) error {
	for _, opt := range options {
		if err := opt(ctx); err != nil {
			return fmt.Errorf("stage %s execution failed: %w", name, err)
		}
	}

	return nil
}

func (b *Booter) shutdown(ac *AppContext, exitErr error) {
	lg := mylog.AppLoggerWithStop()
	lg.Error().Stack().Err(exitErr).Msg("received signal, exit now")

	ta := app.GetTheApp()

	timeout := time.Duration(ta.GetConfig().AppCfg.GracefulExitTimeoutMills) * time.Millisecond
	ctxTimeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var allErrs []error
	err := b.shutdownWithHook(ac, ctxTimeout, ac.preHooks)
	if err != nil {
		allErrs = append(allErrs, err)
	}

	select {
	case <-ctxTimeout.Done():
		allErrs = append(allErrs, ctxTimeout.Err())
		exit(allErrs, true)
		return
	default:

	}

	// release all collected resources
	errs, aborted := ac.finalizer.Release(ctxTimeout)
	if len(errs) > 0 {
		allErrs = append(allErrs, ctxTimeout.Err())
	}

	if aborted {
		exit(allErrs, true)
		return
	}

	err = b.shutdownWithHook(ac, ctxTimeout, ac.postHooks)
	if err != nil {
		allErrs = append(allErrs, err)
	}

	exit(allErrs, false)
}

func exit(errs []error, aborted bool) {
	log.Printf("app finalizer release completed, errs:%+v, aborted:%t\n", errs, aborted)

	time.Sleep(16 * time.Millisecond)
}

func (b *Booter) shutdownWithHook(ac *AppContext, ctx context.Context, hooks []ShutdownHook) error {
	if len(hooks) == 0 {
		return nil
	}

	var errs []error
	for _, hook := range hooks {
		err := hook(ac, ctx)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)

	}

	return nil
}

// 初始化日志
func WithLogger(nameGenFunc mylog.LogFileNameGenerator) AppOption {
	return func(ac *AppContext) error {
		ta := app.GetTheApp()

		remoteWriter := mylog.InitLogger(ta.GetConfig().LogCfg, ta.IsDevProfile(), ta.GetAppName(), nameGenFunc)

		ac.finalizer.Collect("log_writer", dlog.NewLogRemoteWriter(remoteWriter))
		return nil
	}
}

// nacos客户端
func WithNacosClient() AppOption {
	return func(ac *AppContext) error {
		ta := app.GetTheApp()

		nc, err := cnacos.NewNacosClient(dnacos.ToNacosCfg(ta.GetConfig().NacosCfg))
		if err != nil {
			return fmt.Errorf("failed to create Nacos client: %w", err)
		}

		ac.nacosClient = nc
		ac.finalizer.Collect("cnacos", nc)
		return nil
	}
}

// nacos配置中心
func WithNacosConfig(receiver dnacos.ConfigurationReceiver) AppOption {
	return func(ac *AppContext) error {
		if ac.nacosClient == nil {
			return errors.New("nacos Client is required for Nacos Config but not found in AppContext")
		}

		ac.configureReceiver = receiver

		ta := app.GetTheApp()

		autoConfig := dnacos.NewNacosAutoConfiguration(
			cfgnacos.NewNacosConfigCenter(ac.nacosClient.GetConfigClient()),
			ta.GetConfig().NacosCenterCfg.ConfigCfg,
			receiver,
		)

		ac.finalizer.Collect("nacos_config", autoConfig)
		return nil
	}
}

// nacos注册/发现中心
func WithNacosRegistry() AppOption {
	return func(ac *AppContext) error {
		if ac.nacosClient == nil {
			return errors.New("nacos Client is required for Nacos Registry but not found in AppContext")
		}
		ta := app.GetTheApp()

		autoRegistry := dnacos.NewNacosAutoRegistry(
			regnacos.NewNacosRegistry(ac.nacosClient.GetNamingClient()),
			ta.GetConfig().NacosCenterCfg.RegistryDiscoverCfg,
		)

		ac.finalizer.Collect("nacos_registry", autoRegistry)
		return nil
	}
}

// 启动http服务
func WithHttpServer(errHandler fiber.ErrorHandler) AppOption {
	return func(ac *AppContext) error {
		fhs := shttp.NewFiberHttpServer(shttp.DefaultFiberServerConfig(app.GetTheApp().GetHttpPort()), errHandler)

		if ac.routeBinder != nil {
			ac.routeBinder.BindFiber(fhs.GetFiber())
		}

		ac.finalizer.Collect("http_server", fhs)

		return nil
	}
}

// 启动rpc服务
func WithRpcServer() AppOption {
	return func(ac *AppContext) error {
		srpc.InitArpcLogAdapter()

		as := srpc.NewArpcServer(app.GetTheApp().GetRpcPort())

		if ac.routeBinder != nil {
			ac.routeBinder.BindArpc(as.GetArpcSrv())
		}

		ac.finalizer.Collect("arpc_server", as)

		return nil
	}
}

type ConfigureLoaded func(ac *AppContext) error

// Component Stage 配置加载完毕(静态配置, 和动态配置第一次加载)
func WithConfigureLoaded(cl ConfigureLoaded) AppOption {
	return func(ac *AppContext) error {
		if ac.GetConfigureReceiver() == nil {
			return errors.New("configure can not ready without receiver")
		}

		return cl(ac)
	}
}

// 启动rpc客户端
func WithRpcClientFactory(lb rclient.LoadBalancer) AppOption {
	if lb == nil {
		lb = rclient.NewRoundRobinLoadBalancer()
	}

	return func(ac *AppContext) error {
		srpc.InitArpcLogAdapter()

		clientFactory := rcfactory.NewNacosArpcClientFactory(
			disnacos.NewNacosDiscovery(ac.nacosClient.GetNamingClient()),
			lb,
			app.GetTheApp().GetConfig().NacosCenterCfg.RegistryDiscoverCfg,
		)

		ac.arpcClientFactory = clientFactory

		ac.finalizer.Collect("rpc_client_factory", clientFactory)

		return nil
	}
}

func WithShutdownPreHooks(hooks ...ShutdownHook) AppOption {
	return func(ac *AppContext) error {
		ac.preHooks = append(ac.preHooks, hooks...)
		return nil
	}
}

func WithShutdownPostHooks(hooks ...ShutdownHook) AppOption {
	return func(ac *AppContext) error {
		ac.postHooks = append(ac.postHooks, hooks...)
		return nil
	}
}
