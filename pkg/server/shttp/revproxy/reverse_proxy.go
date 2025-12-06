package revproxy

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/valyala/fasthttp"
	"gmicro_pkg/pkg/app"
	"gmicro_pkg/pkg/mylog"
	"gmicro_pkg/pkg/regdis"
	"gmicro_pkg/pkg/utils/usli"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultTimeoutMills = 15_000
)

type HostClientConfig struct {
	MaxConns            int
	MaxIdleConnDuration time.Duration
	MaxConnDuration     time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	MaxResponseBodySize int
	MaxConnWaitTimeout  time.Duration
}

type HttpServerReverseProxy struct {
	serviceName string
	lbCli       *fasthttp.LBClient
	discovery   regdis.Discovery
	disExtraMap map[string]any
	stopped     atomic.Bool
	hadWatched  atomic.Bool
	mu          sync.Mutex
	discovered  []*regdis.Instance
	unavailable atomic.Bool
	hcCfg       HostClientConfig
}

func NewHttpServerReverseProxy(serviceName string, discovery regdis.Discovery, disExtraMap map[string]any, cfg HostClientConfig) *HttpServerReverseProxy {
	timeout := int(math.Max(float64(cfg.ReadTimeout), float64(cfg.WriteTimeout))) + 1500
	if timeout == 0 {
		timeout = defaultTimeoutMills
	}

	revProxy := &HttpServerReverseProxy{
		serviceName: serviceName,
		disExtraMap: disExtraMap,
		discovery:   discovery,
		hcCfg:       cfg,
		lbCli: &fasthttp.LBClient{
			Timeout: time.Duration(timeout) * time.Millisecond,
			Clients: make([]fasthttp.BalancingClient, 0),
		},
	}

	// 拉取一次配置
	instances, err := discovery.Discover(
		regdis.DiscoverParam{
			ServiceName: serviceName,
			DisType:     regdis.ForHttp,
			Extra:       disExtraMap,
		},
	)

	lg := mylog.AppLoggerWithInit()
	if err != nil {
		lg.Error().
			Stack().
			Err(err).
			Str("service_name", serviceName).
			Msgf("discover server instance failed in reverse proxy, extraMap:%+v", revProxy.disExtraMap)
		revProxy.unavailable.Store(true)
	} else {
		lg.Debug().
			Str("service_name", serviceName).
			Msgf("discover server instance success in reverse proxy, extraMap:%+v, instances:%s", revProxy.disExtraMap, regdis.PrettyOutput(instances))

		revProxy.modifyClients(instances)
	}

	revProxy.watch()

	return revProxy
}

func (srp *HttpServerReverseProxy) ReverseProxy(modifyReqs, modifyResps []fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if srp.unavailable.Load() {
			isDev := app.GetTheApp().IsDevProfile()

			c.Status(http.StatusServiceUnavailable)
			if isDev {
				return c.SendString(fmt.Sprintf("Upstream server:[%s] unavailable", srp.serviceName))
			} else {
				return c.SendString("Upstream server unavailable")
			}
		}

		// Set request and response
		req := c.Request()
		res := c.Response()

		// Don't proxy "Connection" header
		req.Header.Del(fiber.HeaderConnection)

		for _, handler := range modifyReqs {
			if handler == nil {
				continue
			}

			if err := handler(c); err != nil {
				return err
			}
		}

		req.SetRequestURI(utils.UnsafeString(req.RequestURI()))

		// Forward request
		if err := srp.lbCli.Do(req, res); err != nil {
			return err
		}

		// Don't proxy "Connection" header
		res.Header.Del(fiber.HeaderConnection)

		for _, handler := range modifyResps {
			if handler == nil {
				continue
			}

			if err := handler(c); err != nil {
				return err
			}
		}

		// Return nil to end proxying if no error
		return nil
	}
}

func (srp *HttpServerReverseProxy) Shutdown() error {
	if !srp.stopped.CompareAndSwap(false, true) {
		return nil
	}

	srp.unavailable.Store(true)

	err := srp.unwatch()

	if err != nil {
		return err
	}

	srp.lbCli.RemoveClients(func(bc fasthttp.BalancingClient) bool {
		return true
	})

	return nil
}

func (srp *HttpServerReverseProxy) watch() {
	if !srp.hadWatched.CompareAndSwap(false, true) {
		return
	}

	err := srp.discovery.Watch(
		regdis.DiscoverParam{
			ServiceName: srp.serviceName,
			DisType:     regdis.ForHttp,
			Extra:       srp.disExtraMap,
		},
		func(instances []*regdis.Instance, err error) {
			lg := mylog.AppLoggerWithListen()

			if err != nil {
				lg.Error().
					Stack().
					Err(err).
					Str("service_name", srp.serviceName).
					Msgf("receive server instances changed event but has a error in reverse proxy, extraMap:%+v", srp.disExtraMap)
				return
			}

			if srp.stopped.Load() {
				return
			}

			lg.Info().
				Str("service_name", srp.serviceName).
				Msgf("receive server instances changed event in reverse proxy, extraMap:%v, instances:%s", srp.disExtraMap, regdis.PrettyOutput(instances))

			srp.modifyClients(instances)
		},
	)

	if err != nil {
		if err != nil {
			lg := mylog.AppLoggerWithListen()

			lg.Error().Stack().Err(err).Str("service_name", srp.serviceName).Msgf("watch server instance failed in http server reverse proxy")
		}
	}
}

func (srp *HttpServerReverseProxy) unwatch() error {
	return srp.discovery.Unwatch(
		regdis.DiscoverParam{
			ServiceName: srp.serviceName,
			DisType:     regdis.ForHttp,
			Extra:       srp.disExtraMap,
		},
		func(_ error) {

		},
	)
}

func (srp *HttpServerReverseProxy) modifyClients(instances []*regdis.Instance) {
	srp.mu.Lock()

	if len(instances) == 0 {
		srp.discovered = make([]*regdis.Instance, 0)
		srp.unavailable.Store(true)
		srp.mu.Unlock()

		// remove all host clients
		// internal concurrency safety
		srp.lbCli.RemoveClients(func(bc fasthttp.BalancingClient) bool {
			return true
		})

		return
	}

	discovered := srp.discovered

	defer srp.mu.Unlock()

	// keep old, insert new, delete not exists
	newInsMap := make(map[string]bool, len(instances))
	for _, ins := range instances {
		newInsMap[ins.InsIdentity()] = true
	}

	beRemoved := make(map[string]bool, 0)
	shouldKeep := make([]*regdis.Instance, 0)

	for _, ins := range discovered {
		idt := ins.InsIdentity()
		if exists := newInsMap[idt]; exists {
			// keep it
			shouldKeep = append(shouldKeep, ins)
		} else {
			// not exists, should be removed
			beRemoved[idt] = true
		}
	}

	// found need to create client, but not in exists list
	beCreatedIns := make([]*regdis.Instance, 0)
	for _, ins := range instances {
		newIdt := ins.InsIdentity()
		found := false

		for _, keepIns := range shouldKeep {
			if newIdt == keepIns.InsIdentity() {
				found = true
				break
			}
		}

		if !found {
			beCreatedIns = append(beCreatedIns, ins)
		}
	}

	newClients := make([]fasthttp.BalancingClient, 0)
	// create new client
	for _, ins := range beCreatedIns {
		newClients = append(newClients, srp.createHostClient(ins))

		shouldKeep = append(shouldKeep, ins)
	}

	srp.discovered = shouldKeep
	srp.unavailable.Store(false)

	// remove dropped client
	srp.lbCli.RemoveClients(func(bc fasthttp.BalancingClient) bool {
		hc := bc.(*fasthttp.HostClient)
		return beRemoved[hc.Addr]
	})

	clients := make([]fasthttp.BalancingClient, 0, len(shouldKeep))
	for _, cli := range srp.lbCli.Clients {
		clients = append(clients, cli)
	}

	// add new client
	for _, cli := range newClients {
		srp.lbCli.AddClient(cli)
		clients = append(clients, cli)
	}

	srp.lbCli.Clients = clients

	lg := mylog.AppLoggerWithListen()
	if e := lg.Debug(); e.Enabled() {
		remainAddr := usli.Conv(shouldKeep, func(ins *regdis.Instance) string {
			return ins.InsIdentity()
		})

		e.Str("service_name", srp.serviceName).Msgf("modify clients completed, final clients:%+v", remainAddr)
	}
}

func (srp *HttpServerReverseProxy) createHostClient(ins *regdis.Instance) fasthttp.BalancingClient {
	cfg := srp.hcCfg
	return &fasthttp.HostClient{
		NoDefaultUserAgentHeader: true,
		DisablePathNormalizing:   true,
		Addr:                     ins.InsIdentity(),
		ReadTimeout:              cfg.ReadTimeout,
		WriteTimeout:             cfg.WriteTimeout,
		MaxConns:                 cfg.MaxConns,
		MaxConnDuration:          cfg.MaxConnDuration,
		MaxIdleConnDuration:      cfg.MaxIdleConnDuration,
		MaxConnWaitTimeout:       cfg.MaxConnWaitTimeout,
		MaxResponseBodySize:      cfg.MaxResponseBodySize,
	}
}
