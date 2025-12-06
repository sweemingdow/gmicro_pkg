package gserver

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/sweemingdow/gmicro_pkg/pkg/decorate/dnacos"
	"github.com/sweemingdow/gmicro_pkg/pkg/lifetime"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"github.com/sweemingdow/gmicro_pkg/pkg/regdis"
	"github.com/sweemingdow/gmicro_pkg/pkg/server/shttp/revproxy"
	"time"
)

const (
	GatewayRouterTableConfigName = "router-tables.json"
)

// 用于网关的HostClient
const (
	defaultMaxConn                  = 256
	defaultMaxIdleConnDurationMills = 120_000
	defaultMaxConnDurationMills     = 150_000
	defaultReadTimeoutMills         = 10_500
	defaultWriteTimeoutMills        = 10_500
	defaultMaxConnWaitTimeoutMills  = 5_000
	defaultMaxResponseBodySize      = 2 * 1024 * 1024
)

type RouterTableConfig struct {
	CommonUpstreamClientCfg UpstreamClientConfig `json:"commonUpstreamClientCfg,omitempty"`
	Tables                  []TableItem          `json:"tables,omitempty"`
}

type TableItem struct {
	Id                string               `json:"id,omitempty"`
	MatchRule         RouteMatchRule       `json:"matchRule,omitempty"`
	UpstreamClientCfg UpstreamClientConfig `json:"upstreamClientConfig,omitempty"`
}

type RouteMatchRule struct {
	Type string         `json:"type,omitempty"`
	Path string         `json:"path,omitempty"`
	Args map[string]any `json:"args,omitempty"`
}

type UpstreamClientConfig struct {
	MaxConns                 int `json:"maxConns,omitempty"`
	MaxIdleConnDurationMills int `json:"maxIdleConnDurationMills,omitempty"`
	MaxConnDurationMills     int `json:"maxConnDurationMills,omitempty"`
	ReadTimeoutMills         int `json:"readTimeoutMills,omitempty"`
	WriteTimeoutMills        int `json:"writeTimeoutMills,omitempty"`
	MaxResponseBodySize      int `json:"maxResponseBodySize,omitempty"`
	MaxConnWaitTimeoutMills  int `json:"maxConnWaitTimeoutMills,omitempty"`
}

type ConfigurableGatewayServer struct {
	gwSrv *GatewayServer
}

func NewConfigurableGatewayServer(
	tables []RouterTableItem,
	hsCfg HotLoadServerConfig,
	ec chan<- error,
	discovery regdis.Discovery,
	disExtraMap map[string]any,
	modifyReqs, modifyResps []fiber.Handler,
	errHandler fiber.ErrorHandler,
) *ConfigurableGatewayServer {
	cgs := &ConfigurableGatewayServer{
		gwSrv: NewGatewayServer(
			tables,
			hsCfg,
			ec,
			discovery,
			disExtraMap,
			modifyReqs,
			modifyResps,
			errHandler,
		),
	}

	lifetime.GetAppFinalizer().Collect("configurable_gw_server", cgs)

	dnacos.RegisterObserver(
		GatewayRouterTableConfigName,
		func(dataId string, val any) {
			cfg := val.(RouterTableConfig)
			tabItems := Cfg2routerItems(cfg)

			if len(tabItems) > 0 {
				cgs.gwSrv.OnRouterTableRefresh(tabItems)
			}
		},
	)

	return cgs
}

func (cgs *ConfigurableGatewayServer) OnCreated(_ chan<- error) {
}

func (cgs *ConfigurableGatewayServer) OnDispose(ctx context.Context) error {
	lg := mylog.AppLoggerWithStop()

	if err := cgs.gwSrv.Shutdown(ctx); err != nil {
		lg.Error().Stack().Err(err).Msg("configurable gw server shutdown failed")
		return err
	}

	lg.Info().Msg("configurable gw server stopped successfully")

	return nil
}

func Cfg2routerItems(cfg RouterTableConfig) []RouterTableItem {
	if len(cfg.Tables) == 0 {
		return nil
	}

	commCliCfg := &cfg.CommonUpstreamClientCfg
	correctCommonClientConfig(commCliCfg)

	tabItems := make([]RouterTableItem, len(cfg.Tables))
	for idx, tab := range cfg.Tables {
		tabItems[idx] = RouterTableItem{
			ServiceName: tab.Id,
			MatchRule: MatchRule{
				Type: tab.MatchRule.Type,
				Path: tab.MatchRule.Path,
				Args: tab.MatchRule.Args,
			},
			HostClientCfg: convertHostClientConfig(commCliCfg, tab.UpstreamClientCfg),
		}
	}

	return tabItems
}

func correctCommonClientConfig(commCliCfg *UpstreamClientConfig) {
	if commCliCfg.MaxConns == 0 {
		commCliCfg.MaxConns = defaultMaxConn
	}

	if commCliCfg.MaxIdleConnDurationMills == 0 {
		commCliCfg.MaxIdleConnDurationMills = defaultMaxIdleConnDurationMills
	}

	if commCliCfg.MaxConnDurationMills == 0 {
		commCliCfg.MaxConnDurationMills = defaultMaxConnDurationMills
	}

	if commCliCfg.ReadTimeoutMills == 0 {
		commCliCfg.ReadTimeoutMills = defaultReadTimeoutMills
	}

	if commCliCfg.WriteTimeoutMills == 0 {
		commCliCfg.WriteTimeoutMills = defaultWriteTimeoutMills
	}

	if commCliCfg.MaxResponseBodySize == 0 {
		commCliCfg.MaxResponseBodySize = defaultMaxResponseBodySize
	}

	if commCliCfg.MaxConnWaitTimeoutMills == 0 {
		commCliCfg.MaxConnWaitTimeoutMills = defaultMaxConnWaitTimeoutMills
	}
}

func convertHostClientConfig(commCliCfg *UpstreamClientConfig, cliCfg UpstreamClientConfig) revproxy.HostClientConfig {
	var maxConns = cliCfg.MaxConns
	if maxConns == 0 {
		maxConns = commCliCfg.MaxConns
	}

	var maxIdleConnDurationMills = cliCfg.MaxIdleConnDurationMills
	if maxIdleConnDurationMills == 0 {
		maxIdleConnDurationMills = commCliCfg.MaxIdleConnDurationMills
	}

	var maxConnDurationMills = cliCfg.MaxConnDurationMills
	if maxConnDurationMills == 0 {
		maxConnDurationMills = commCliCfg.MaxConnDurationMills
	}
	var readTimeoutMills = cliCfg.ReadTimeoutMills
	if readTimeoutMills == 0 {
		readTimeoutMills = commCliCfg.ReadTimeoutMills
	}

	var writeTimeoutMills = cliCfg.WriteTimeoutMills
	if writeTimeoutMills == 0 {
		writeTimeoutMills = commCliCfg.WriteTimeoutMills
	}

	var maxResponseBodySize = cliCfg.MaxResponseBodySize
	if maxResponseBodySize == 0 {
		maxResponseBodySize = commCliCfg.MaxResponseBodySize
	}

	var maxConnWaitTimeoutMills = cliCfg.MaxConnWaitTimeoutMills
	if maxConnWaitTimeoutMills == 0 {
		maxConnWaitTimeoutMills = commCliCfg.MaxConnWaitTimeoutMills
	}

	return revproxy.HostClientConfig{
		MaxConns:            maxConns,
		MaxIdleConnDuration: time.Duration(maxIdleConnDurationMills) * time.Millisecond,
		MaxConnDuration:     time.Duration(maxConnDurationMills) * time.Millisecond,
		ReadTimeout:         time.Duration(readTimeoutMills) * time.Millisecond,
		WriteTimeout:        time.Duration(writeTimeoutMills) * time.Millisecond,
		MaxResponseBodySize: maxResponseBodySize,
		MaxConnWaitTimeout:  time.Duration(maxConnWaitTimeoutMills) * time.Millisecond,
	}
}
