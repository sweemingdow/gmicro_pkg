package dnacos

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/sweemingdow/gmicro_pkg/pkg/cfgcenter/cfgnacos"
	"github.com/sweemingdow/gmicro_pkg/pkg/config"
	"github.com/sweemingdow/gmicro_pkg/pkg/lifetime"
	"github.com/sweemingdow/gmicro_pkg/pkg/mylog"
	"sync"
)

const (
	StaticConfigName  = "static-config.yaml"
	DynamicConfigName = "dynamic-config.yaml"
)

type ConfigurationReceiver interface {
	OnReceiveStatic(dataId, groupName, data string)

	OnReceiveDynamic(dataId, groupName, data string, firstLoad bool)

	RecentlyConfigure(dataId string) (any, bool)
}

type nacosAutoConfiguration struct {
	cfgCenter *cfgnacos.NacosConfigCenter
	cfgConfig config.NacosConfigConfig
	receiver  ConfigurationReceiver
	mu        sync.Mutex
	listens   []cfgnacos.AcquireParam
}

func NewNacosAutoConfiguration(
	cfgCenter *cfgnacos.NacosConfigCenter,
	cfgConfig config.NacosConfigConfig,
	receiver ConfigurationReceiver,
) lifetime.LifeCycle {
	return &nacosAutoConfiguration{
		cfgCenter: cfgCenter,
		cfgConfig: cfgConfig,
		receiver:  receiver,
		listens:   make([]cfgnacos.AcquireParam, 0),
	}
}

func (nac *nacosAutoConfiguration) OnCreated(ec chan<- error) {
	cfg := nac.cfgConfig
	defGroupName := cfg.GroupName

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		for _, item := range cfg.Static {
			grpName := item.GroupName
			if grpName == "" {
				grpName = defGroupName
			}

			data, err := nac.cfgCenter.Acquire(cfgnacos.AcquireParam{
				CfgId:     item.Name,
				GroupName: grpName,
			})
			if err != nil {
				ec <- err
				return
			}

			nac.receiver.OnReceiveStatic(item.Name, grpName, data)
		}
	}()

	go func() {
		defer wg.Done()

		for _, item := range cfg.Dynamic {
			grpName := item.GroupName
			if grpName == "" {
				grpName = defGroupName
			}

			acqData, _, err := nac.cfgCenter.AcquireAndListen(cfgnacos.AcquireListenParam{
				CfgId:     item.Name,
				GroupName: grpName,
				OnChanged: func(namespace, group, dataId, data string) {
					nac.receiver.OnReceiveDynamic(item.Name, grpName, data, false)
				},
			})

			if err != nil {
				ec <- err
				return
			}

			nac.mu.Lock()
			nac.listens = append(
				nac.listens,
				cfgnacos.AcquireParam{
					CfgId:     item.Name,
					GroupName: grpName,
				},
			)
			nac.mu.Unlock()

			nac.receiver.OnReceiveDynamic(item.Name, grpName, acqData, true)
		}

	}()

	wg.Wait()
}

func (nac *nacosAutoConfiguration) OnDispose(ctx context.Context) error {
	UnregisterAll()

	nac.mu.Lock()
	var errs []error
	for _, ap := range nac.listens {
		err := nac.cfgCenter.UnListen(ap)
		if err != nil {
			errs = append(errs, fmt.Errorf("unListen config listener failed:%w, dataId:%s, group:%s", err, ap.CfgId, ap.GroupName))
		}
	}
	nac.mu.Unlock()

	select {
	case <-ctx.Done():
		errs = append(errs, ctx.Err())
	default:

	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	lg := mylog.AppLoggerWithStop()
	lg.Info().Msg("nacos configuration stopped successfully")

	return nil
}

func IsDefaultStaticConfig(dataId string) bool {
	return dataId == StaticConfigName
}

func IsDefaultDynamicConfig(dataId string) bool {
	return dataId == DynamicConfigName
}

func LogWhenReceived(dataId, groupName, data string, isStatic, firstLoad bool) zerolog.Logger {
	lg := mylog.AppLoggerWithListen()

	if isStatic {
		lg.Info().Str("data_id", dataId).Str("group_name", groupName).Str("data", data).Msg("receive static config data")
	} else {
		lg.Info().Str("data_id", dataId).Str("group_name", groupName).Str("data", data).Msgf("receive dynamic config data, firstLoad:%t", firstLoad)
	}

	return lg
}

// 存储最新的配置, 配置留底
type ConfigureStorage struct {
	m    map[string]any
	rwMu sync.RWMutex
}

func NewConfigureStorage() *ConfigureStorage {
	return &ConfigureStorage{
		m: make(map[string]any),
	}
}

func (cs *ConfigureStorage) Store(key string, val any) {
	cs.rwMu.Lock()
	cs.m[key] = val
	cs.rwMu.Unlock()
}

func (cs *ConfigureStorage) Get(key string) (any, bool) {
	cs.rwMu.RLock()
	val, ok := cs.m[key]
	cs.rwMu.RUnlock()

	return val, ok
}
