package gwncfg

import (
	"github.com/sweemingdow/gmicro_pkg/gateway/pkg/gserver"
	"github.com/sweemingdow/gmicro_pkg/pkg/decorate/dnacos"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"github.com/sweemingdow/gmicro_pkg/pkg/parser/json"
	"github.com/sweemingdow/gmicro_pkg/pkg/parser/yaml"
)

const (
	GatewayRouterTableConfigName = gserver.GatewayRouterTableConfigName
)

type gatewayConfigurationReceiver struct {
	// 配置留底
	cs *dnacos.ConfigureStorage
}

func NewGatewayConfigurationReceiver() dnacos.ConfigurationReceiver {
	return &gatewayConfigurationReceiver{
		cs: dnacos.NewConfigureStorage(),
	}
}

func (gcr *gatewayConfigurationReceiver) OnReceiveStatic(dataId, groupName, data string) {
	lg := dnacos.LogWhenReceived(dataId, groupName, data, true, false)

	if dnacos.IsDefaultStaticConfig(dataId) {
		sc, err := parseGatewayStaticConfig(data)
		if err != nil {
			lg.Error().Stack().Str("data", data).Err(err).Msg("parse gateway static config failed")
			return
		}

		gcr.cs.Store(dataId, sc)
	}
}

func (gcr *gatewayConfigurationReceiver) OnReceiveDynamic(dataId, groupName, data string, firstLoad bool) {
	lg := dnacos.LogWhenReceived(dataId, groupName, data, false, firstLoad)

	if dnacos.IsDefaultDynamicConfig(dataId) {
		dc, err := parseGatewayDynamicConfig(data)
		if err != nil {
			lg.Error().Stack().Str("data", data).Err(err).Msg("parse gateway dynamic config failed")
			return
		}

		gcr.cs.Store(dataId, dc)
		mylog.SetLoggersLevel(dc.LogLevel)

	} else if dataId == GatewayRouterTableConfigName {
		cfg, err := parseRouterTable(data)
		if err != nil {
			lg.Error().Stack().Err(err).Msg("parse gateway router table failed")
			return
		}

		gcr.cs.Store(dataId, cfg)

		dnacos.Notify(dataId, cfg)
	}
}

func (gcr *gatewayConfigurationReceiver) RecentlyConfigure(dataId string) (any, bool) {
	return gcr.cs.Get(dataId)
}

func parseGatewayStaticConfig(data string) (GatewayStaticConfig, error) {
	var sc GatewayStaticConfig
	if err := yaml.Parse([]byte(data), &sc); err != nil {
		return sc, err
	}

	return sc, nil
}

func parseGatewayDynamicConfig(data string) (GatewayDynamicConfig, error) {
	var dc GatewayDynamicConfig
	if err := yaml.Parse([]byte(data), &dc); err != nil {
		return dc, err
	}

	return dc, nil
}

func parseRouterTable(data string) (gserver.RouterTableConfig, error) {
	var cfg gserver.RouterTableConfig
	if err := json.Parse([]byte(data), &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
