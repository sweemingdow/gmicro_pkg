package gserver

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"gmicro_pkg/pkg/mylog"
	"gmicro_pkg/pkg/regdis"
	"gmicro_pkg/pkg/server/shttp/revproxy"
	"gmicro_pkg/pkg/utils/usli"
	"sync"
)

type RouterTableItem struct {
	ServiceName   string
	MatchRule     MatchRule
	HostClientCfg revproxy.HostClientConfig
}

type GatewayServer struct {
	mu          sync.Mutex
	name2proxy  map[string]*revproxy.HttpServerReverseProxy
	name2item   map[string]RouterTableItem
	discovery   regdis.Discovery
	disExtraMap map[string]any
	hlSrv       *HotLoadServer
	modifyReqs  []fiber.Handler // origin req handlers
	modifyResps []fiber.Handler // origin resp handlers
}

func NewGatewayServer(
	tables []RouterTableItem,
	hsCfg HotLoadServerConfig,
	ec chan<- error,
	discovery regdis.Discovery,
	disExtraMap map[string]any,
	modifyReqs, modifyResps []fiber.Handler,
	errHandler fiber.ErrorHandler,
) *GatewayServer {
	gs := &GatewayServer{
		name2proxy:  make(map[string]*revproxy.HttpServerReverseProxy),
		name2item:   make(map[string]RouterTableItem),
		discovery:   discovery,
		disExtraMap: disExtraMap,
		modifyReqs:  modifyReqs,
		modifyResps: modifyResps,
	}

	gs.OnRouterTableRefresh(tables)

	// copy
	gs.mu.Lock()
	path2handler := gs.createPath2handler()
	gs.mu.Unlock()

	gs.hlSrv = NewHotLoadServer(ec, hsCfg, path2handler, errHandler)

	return gs
}

// 提供对外接口, 动态更新路由表
func (gs *GatewayServer) OnRouterTableRefresh(tables []RouterTableItem) {
	if len(tables) == 0 {
		panic("tables is required")
	}

	lg := mylog.AppLoggerWithListen()
	lg.Info().Msgf("router tables refreshed, tables:%+v", tables)

	name2item := usli.ToItMap(tables, func(tab RouterTableItem) string {
		return tab.ServiceName
	})

	gs.mu.Lock()
	gs.name2item = name2item

	name2proxy := gs.name2proxy

	beRemoved := make(map[string]struct{})

	for name, _ := range name2proxy {
		//
		if _, ok := name2item[name]; !ok {
			beRemoved[name] = struct{}{}
		}

	}

	beAdded := make(map[string]struct{})

	for name, _ := range name2item {
		if _, ok := name2proxy[name]; !ok {
			beAdded[name] = struct{}{}
		}
	}

	for name, _ := range beRemoved {
		if proxy, ok := name2proxy[name]; ok {
			if err := proxy.Shutdown(); err != nil {
				lg.Error().Stack().Err(err).Msgf("shutdown reverse proxy for %s failed", name)
			}

			delete(name2proxy, name)
		}
	}

	for name, _ := range beAdded {
		name2proxy[name] = revproxy.NewHttpServerReverseProxy(name, gs.discovery, gs.disExtraMap, gs.name2item[name].HostClientCfg)
	}

	if gs.hlSrv != nil {
		gs.hlSrv.Reload(gs.createPath2handler())
	}

	gs.mu.Unlock()
}

func (gs *GatewayServer) Shutdown(ctx context.Context) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.hlSrv != nil {
		return gs.hlSrv.Shutdown(ctx)
	}

	clear(gs.name2proxy)

	clear(gs.name2item)

	return nil
}

func (gs *GatewayServer) createPath2handler() map[string]fiber.Handler {
	path2handler := make(map[string]fiber.Handler, len(gs.name2proxy))

	for name, proxy := range gs.name2proxy {
		item := gs.name2item[name]

		reqHandlers := make([]fiber.Handler, 0, len(gs.modifyReqs)+1)
		if pathHandler := createRuleHandler(name, item.MatchRule); pathHandler != nil {
			reqHandlers = append(reqHandlers, pathHandler)
		}
		reqHandlers = append(reqHandlers, gs.modifyReqs...)

		path2handler[item.MatchRule.Path] = proxy.ReverseProxy(reqHandlers, gs.modifyResps)
	}

	return path2handler
}
