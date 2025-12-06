package rclient

import (
	"context"
	"errors"
	"fmt"
	"github.com/lesismal/arpc"
	"gmicro_pkg/pkg/mylog"
	"gmicro_pkg/pkg/regdis"
	"gmicro_pkg/pkg/utils/usli"
	"sync"
	"sync/atomic"
	"time"
)

type ArpcClientProxy interface {
	IsReady() bool

	Call(method string, req any, rsp any, timeout time.Duration, args ...any) error

	CallContext(ctx context.Context, method string, req any, rsp any, timeout time.Duration, args ...any) error

	CallAsync(method string, req any, handler arpc.AsyncHandlerFunc, timeout time.Duration, args ...any) error

	PushMsg(msg *arpc.Message, timeout time.Duration) error

	Notify(method string, data any, timeout time.Duration, args ...any) error

	NotifyContext(ctx context.Context, method string, data any, args ...any) error

	Stop(ctx context.Context) error
}

type arpcClientProxy struct {
	// belong which service
	serviceName string

	// extra data for discovery
	extraMap map[string]any

	discovery regdis.Discovery

	// clients pool
	clients []*arpcClientWrap

	lb LoadBalancer

	rwMu sync.RWMutex

	hadWatched atomic.Bool

	initErr atomic.Pointer[error]

	dialTimeout time.Duration
}

func NewArpcClientProxy(
	serviceName string,
	discovery regdis.Discovery,
	extraMap map[string]any,
	lb LoadBalancer,
	dialTimeout time.Duration,
) (ArpcClientProxy, error) {
	cp := &arpcClientProxy{
		serviceName: serviceName,
		extraMap:    extraMap,
		clients:     make([]*arpcClientWrap, 0),
		lb:          lb,
		discovery:   discovery,
		dialTimeout: dialTimeout,
	}

	// attempts to obtain available instances in init
	instances, err := discovery.Discover(
		regdis.DiscoverParam{
			ServiceName: serviceName,
			DisType:     regdis.ForRpc,
			Extra:       cp.extraMap,
		},
	)

	// start watch
	defer cp.doWatch()

	lg := mylog.AppLoggerWithInit()
	if err != nil {
		lg.Error().
			Stack().
			Err(err).
			Str("service_name", serviceName).
			Msgf("discover server instance failed in arpc client proxy, extraMap:%+v", cp.extraMap)

		cp.initErr.Store(&err)
	} else {
		lg.Debug().
			Str("service_name", serviceName).
			Msgf("discover server instance success in arpc client proxy, extraMap:%+v, instances:%s", cp.extraMap, regdis.PrettyOutput(instances))

		cp.replaceAllClients(instances)
	}

	// always return the proxy client
	return cp, err
}

func (acp *arpcClientProxy) IsReady() bool {
	if acp.initErr.Load() != nil {
		return false
	}

	acp.rwMu.RLock()
	ready := len(acp.clients) > 0
	acp.rwMu.RUnlock()

	return ready
}

func (acp *arpcClientProxy) Call(method string, req any, resp any, timeout time.Duration, args ...any) error {
	err := acp.readyOrErr()
	if err != nil {
		return err
	}

	cp, err := acp.chooseClient()
	if err != nil {
		return err
	}

	return cp.cli.Call(method, req, resp, timeout, args...)
}

func (acp *arpcClientProxy) CallContext(ctx context.Context, method string, req any, resp any, timeout time.Duration, args ...any) error {
	err := acp.readyOrErr()
	if err != nil {
		return err
	}

	cp, err := acp.chooseClient()
	if err != nil {
		return err
	}

	return cp.cli.CallContext(ctx, method, req, resp, args...)
}

func (acp *arpcClientProxy) CallAsync(method string, req any, handler arpc.AsyncHandlerFunc, timeout time.Duration, args ...any) error {
	err := acp.readyOrErr()
	if err != nil {
		return err
	}

	cp, err := acp.chooseClient()
	if err != nil {
		return err
	}

	return cp.cli.CallAsync(method, req, handler, timeout, args...)
}

func (acp *arpcClientProxy) PushMsg(msg *arpc.Message, timeout time.Duration) error {
	err := acp.readyOrErr()
	if err != nil {
		return err
	}

	cp, err := acp.chooseClient()
	if err != nil {
		return err
	}

	return cp.cli.PushMsg(msg, timeout)
}

func (acp *arpcClientProxy) Notify(method string, data any, timeout time.Duration, args ...any) error {
	err := acp.readyOrErr()
	if err != nil {
		return err
	}

	cp, err := acp.chooseClient()
	if err != nil {
		return err
	}

	return cp.cli.Notify(method, data, timeout, args...)
}

func (acp *arpcClientProxy) NotifyContext(ctx context.Context, method string, data any, args ...any) error {
	err := acp.readyOrErr()
	if err != nil {
		return err
	}

	cp, err := acp.chooseClient()
	if err != nil {
		return err
	}

	return cp.cli.NotifyContext(ctx, method, data, args...)
}

func (acp *arpcClientProxy) Stop(ctx context.Context) error {
	acp.rwMu.Lock()
	defer acp.rwMu.Unlock()

	var errs []error
	for _, cli := range acp.clients {
		cli.close()

		select {
		case <-ctx.Done():
			errs = append(errs, ctx.Err())
			return errors.Join(errs...)
		default:

		}
	}

	if err := acp.discovery.Unwatch(
		regdis.DiscoverParam{
			ServiceName: acp.serviceName,
			Extra:       acp.extraMap,
		}, func(err error) {

		},
	); err != nil {
		errs = append(errs, err)
	}

	acp.clients = nil
	acp.hadWatched.Store(false)

	return errors.Join(errs...)
}

func (acp *arpcClientProxy) readyOrErr() error {
	if errVal := acp.initErr.Load(); errVal != nil {
		return fmt.Errorf("arpc proxy client was not ready:%w for service:%s", *errVal, acp.serviceName)
	}

	return nil
}

func (acp *arpcClientProxy) chooseClient() (*arpcClientWrap, error) {
	acp.rwMu.RLock()
	cp := acp.lb.Select(acp.clients)
	acp.rwMu.RUnlock()

	if cp == nil {
		return nil, fmt.Errorf("can not found server instance with serviceName:%s", acp.serviceName)
	}

	return cp, nil
}

// 监听服务变化, 动态更新连接池
func (acp *arpcClientProxy) doWatch() {
	if !acp.hadWatched.CompareAndSwap(false, true) {
		return
	}

	err := acp.discovery.Watch(
		regdis.DiscoverParam{
			ServiceName: acp.serviceName,
			DisType:     regdis.ForRpc,
			Extra:       acp.extraMap,
		},
		func(instances []*regdis.Instance, err error) {
			lg := mylog.AppLoggerWithListen()

			if err != nil {
				lg.Error().
					Stack().
					Err(err).
					Str("service_name", acp.serviceName).
					Msgf("receive server instances changed event but has a error, extraMap:%+v", acp.extraMap)
			} else {
				lg.Info().
					Str("service_name", acp.serviceName).
					Msgf("receive server instances changed event, extraMap:%v, instances:%s", acp.extraMap, regdis.PrettyOutput(instances))

				acp.modifyClients(instances)
			}
		},
	)

	if err != nil {
		lg := mylog.AppLoggerWithListen()

		lg.Error().Stack().Err(err).Str("service_name", acp.serviceName).Msgf("watch server instance failed")
	}
}

func (acp *arpcClientProxy) replaceAllClients(instances []*regdis.Instance) {
	if len(instances) == 0 {
		return
	}

	lg := mylog.AppLoggerWithStop()

	// create firstly, outside the lock
	newInsSli := make([]*arpcClientWrap, 0, len(instances))
	for _, ins := range instances {
		cli, err := newArpcClientWrap(ins, acp.dialTimeout)
		if err != nil {
			lg.Error().Stack().Err(err).Str("service_name", acp.serviceName).Msg("created clientWrap failed in replace clients")
			continue
		}
		newInsSli = append(newInsSli, cli)
	}

	if len(newInsSli) == 0 {
		return
	}

	acp.rwMu.Lock()
	defer acp.rwMu.Unlock()

	// release old
	if len(acp.clients) > 0 {
		for _, cli := range acp.clients {
			cli.close()
		}
	}

	// replace new
	acp.clients = newInsSli

	acp.resetInitErr()

	if e := lg.Debug(); e.Enabled() {
		remainAddr := usli.Conv(newInsSli, func(t *arpcClientWrap) string {
			return t.ins.InsIdentity()
		})

		e.Str("service_name", acp.serviceName).Msgf("replace clients completed, final clients:%+v", remainAddr)
	}
}

func (acp *arpcClientProxy) resetInitErr() {
	acp.initErr.Store(nil)
}

func (acp *arpcClientProxy) modifyClients(instances []*regdis.Instance) {
	acp.rwMu.RLock()
	if len(acp.clients) == 0 {
		acp.rwMu.RUnlock()
		acp.replaceAllClients(instances)
		return
	}
	acp.rwMu.RUnlock()

	acp.rwMu.Lock()
	defer acp.rwMu.Unlock()

	if len(instances) == 0 {
		// just keep exists client
		return
	}

	// clientWrap: keep old, insert new, delete not exists

	// create new instances map for quick query
	newInsMap := make(map[string]struct{}, len(instances))
	for _, ins := range instances {
		newInsMap[ins.InsIdentity()] = struct{}{}
	}

	// found existed client, but not in new ins list
	beClosedClients := make([]*arpcClientWrap, 0)
	keepClients := make([]*arpcClientWrap, 0)

	clients := acp.clients

	for _, cli := range clients {
		if _, ok := newInsMap[cli.ins.InsIdentity()]; ok {
			// keep it
			keepClients = append(keepClients, cli)
		} else {
			// not exists, should be close
			beClosedClients = append(beClosedClients, cli)
		}
	}

	// found need to create client, but not in exists list
	beCreatedIns := make([]*regdis.Instance, 0)
	for _, ins := range instances {
		newKey := ins.InsIdentity()
		found := false

		for _, cli := range keepClients {
			oldKey := cli.ins.InsIdentity()
			if newKey == oldKey {
				found = true
				break
			}
		}

		if !found {
			beCreatedIns = append(beCreatedIns, ins)
		}
	}

	// close dropped client
	for _, cli := range beClosedClients {
		cli.close()
	}

	lg := mylog.AppLoggerWithListen()

	// create new client
	for _, ins := range beCreatedIns {
		cli, err := newArpcClientWrap(ins, acp.dialTimeout)
		if err != nil {
			lg.Error().Stack().Err(err).Str("service_name", acp.serviceName).Msg("created clientWrap failed in modify clients")
			continue
		}

		keepClients = append(keepClients, cli)
	}

	acp.clients = keepClients

	acp.resetInitErr()

	if e := lg.Debug(); e.Enabled() {
		remainAddr := usli.Conv(keepClients, func(t *arpcClientWrap) string {
			return t.ins.InsIdentity()
		})

		e.Str("service_name", acp.serviceName).Msgf("modify clients completed, final clients:%+v", remainAddr)
	}
}
