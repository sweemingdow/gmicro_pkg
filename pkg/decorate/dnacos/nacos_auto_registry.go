package dnacos

import (
	"context"
	"fmt"
	"gmicro_pkg/pkg/app"
	"gmicro_pkg/pkg/component/cnacos"
	"gmicro_pkg/pkg/config"
	"gmicro_pkg/pkg/lifetime"
	"gmicro_pkg/pkg/mylog"
	"gmicro_pkg/pkg/regdis"
	"gmicro_pkg/pkg/regdis/extra/enacos"
	"gmicro_pkg/pkg/utils"
)

type nacosAutoRegistry struct {
	registry    regdis.Registry
	registryCfg config.NacosRegistryDiscoverConfig
}

func NewNacosAutoRegistry(registry regdis.Registry, registryCfg config.NacosRegistryDiscoverConfig) lifetime.LifeCycle {
	return &nacosAutoRegistry{
		registry:    registry,
		registryCfg: registryCfg,
	}
}

func (nar *nacosAutoRegistry) OnCreated(ec chan<- error) {
	regCfg := nar.registryCfg

	ta := app.GetTheApp()
	var (
		appName  = ta.GetAppName()
		appId    = ta.GetAppId()
		ip       = ta.GetLocalIp()
		metadata = make(map[string]string)
		httpPort = ta.GetHttpPort()
		rpcPort  = ta.GetRpcPort()
	)

	metadata["app_id"] = appId

	var regPort int
	if httpPort != 0 {
		metadata["http_port"] = utils.I2a(httpPort)
		regPort = httpPort
	}

	if rpcPort != 0 {
		if regPort == 0 {
			regPort = rpcPort
		}

		metadata["rpc_port"] = utils.I2a(rpcPort)
	}

	rp := regdis.RegisterParam{
		ServiceName: appName,
		Addr:        fmt.Sprintf("%s:%d", ip, regPort),
		Weight:      enacos.NacosDefaultWeight,
		Metadata:    metadata,
		Extra:       enacos.PkgRegisterExtraParam(regCfg.ClusterName, regCfg.GroupName),
	}

	err := nar.registry.Register(rp)
	if err != nil {
		ec <- err
	}
}

func (nar *nacosAutoRegistry) OnDispose(_ context.Context) error {
	regCfg := nar.registryCfg

	ta := app.GetTheApp()
	var (
		appName = ta.GetAppName()
		ip      = ta.GetLocalIp()
	)

	dp := regdis.DeregisterParam{
		ServiceName: appName,
		Addr:        fmt.Sprintf("%s:%d", ip, ta.GetHttpPort()),
		Extra:       enacos.PkgDeregisterExtraParam(regCfg.ClusterName, regCfg.GroupName),
	}

	err := nar.registry.Deregister(dp)
	if err != nil {
		return err
	}

	lg := mylog.AppLoggerWithStop()
	lg.Info().Msg("instance had be deregister in nacos")

	return nil
}

func ToNacosCfg(nc config.NacosConfig) cnacos.NacosCfg {
	return cnacos.NacosCfg{
		NamespaceId: nc.NamespaceId,
		Addresses:   nc.Addresses,
		Username:    nc.Username,
		Password:    nc.Password,
		LogLevel:    nc.LogLevel,
		LogDir:      nc.LogDir,
		CacheDir:    nc.CacheDir,
	}
}
