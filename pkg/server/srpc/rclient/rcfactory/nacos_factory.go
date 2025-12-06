package rcfactory

import (
	"context"
	"errors"
	"fmt"
	"gmicro_pkg/pkg/config"
	"gmicro_pkg/pkg/mylog"
	"gmicro_pkg/pkg/regdis"
	"gmicro_pkg/pkg/regdis/extra/enacos"
	"gmicro_pkg/pkg/server/srpc/rclient"
	"sync"
	"time"
)

type nacosArpcClientFactory struct {
	discovery regdis.Discovery

	lb rclient.LoadBalancer

	name2clientProxy map[string]rclient.ArpcClientProxy

	regDisCfg config.NacosRegistryDiscoverConfig

	rwMu sync.RWMutex
}

func NewNacosArpcClientFactory(discovery regdis.Discovery, lb rclient.LoadBalancer, regDisCfg config.NacosRegistryDiscoverConfig) ArpcClientFactoryLifecycle {
	return &nacosArpcClientFactory{
		discovery:        discovery,
		lb:               lb,
		name2clientProxy: make(map[string]rclient.ArpcClientProxy, 2),
		regDisCfg:        regDisCfg,
	}
}

func (acf *nacosArpcClientFactory) AcquireClient(serviceName string) rclient.ArpcClientProxy {
	if serviceName == "" {
		panic("the service name is required")
	}

	// lazy init
	acf.rwMu.RLock()
	if cp, ok := acf.name2clientProxy[serviceName]; ok {
		return cp
	}
	acf.rwMu.RUnlock()

	acf.rwMu.Lock()
	defer acf.rwMu.Unlock()

	// check again
	if cp, ok := acf.name2clientProxy[serviceName]; ok {
		return cp
	}

	// service's client proxy not exists, do lazy init now
	cp, err := rclient.NewArpcClientProxy(
		serviceName,
		acf.discovery,
		enacos.PkgDiscoveryExtraParam(acf.regDisCfg.ClusterName, acf.regDisCfg.GroupName),
		acf.lb,
		time.Duration(acf.regDisCfg.DiscoverDialTimeoutMills)*time.Millisecond,
	)

	if err != nil {
		lg := mylog.AppLogger()
		lg.Error().Stack().Err(err).Send()
	}

	// always save, because dynamic listen may trigger data update
	acf.name2clientProxy[serviceName] = cp

	return cp
}

func (acf *nacosArpcClientFactory) Stop(ctx context.Context) error {
	acf.rwMu.RLock()
	defer acf.rwMu.RUnlock()

	var allErrs []error
	for srvName, proxy := range acf.name2clientProxy {
		if err := proxy.Stop(ctx); err != nil {
			allErrs = append(allErrs, fmt.Errorf("%s's arpc client proxy stopped failed:%w", srvName, err))
		}

		select {
		case <-ctx.Done():
			allErrs = append(allErrs, ctx.Err())
			return errors.Join(allErrs...)
		default:

		}
	}

	return errors.Join(allErrs...)
}

func (acf *nacosArpcClientFactory) OnCreated(_ chan<- error) {
}

func (acf *nacosArpcClientFactory) OnDispose(ctx context.Context) error {
	lg := mylog.AppLoggerWithStop()
	lg.Debug().Msg("arpc client factory stop now")

	err := acf.Stop(ctx)

	if err != nil {
		return err
	}

	lg.Info().Msg("arpc client factory stopped successfully")

	return nil
}
