package cnacos

import (
	"context"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gmicro_pkg/pkg/mylog"
	"gmicro_pkg/pkg/utils"
	"gmicro_pkg/pkg/utils/usli"
	"sync/atomic"
)

type NacosCfg struct {
	NamespaceId string
	ClusterName string
	Addresses   string // 192.168.1.101:8848,192.168.1.102:8848,192.168.1.103:8848,
	Username    string
	Password    string
	LogLevel    string
	LogDir      string
	CacheDir    string
}

type NacosClient struct {
	nCli   naming_client.INamingClient
	cCli   config_client.IConfigClient
	closed atomic.Bool
}

func NewNacosClient(cfg NacosCfg) (*NacosClient, error) {
	hpSli := utils.ExtractHp(cfg.Addresses)
	if len(hpSli) == 0 {
		panic("nacos addresses is required")
	}

	scSli := usli.Conv(hpSli, func(t utils.Hp) constant.ServerConfig {
		return constant.ServerConfig{
			IpAddr: t.Host,
			Port:   uint64(t.Port),
		}
	})

	cc := constant.ClientConfig{
		NamespaceId:         cfg.NamespaceId,
		NotLoadCacheAtStart: true,
		Username:            cfg.Username,
		Password:            cfg.Password,
		LogLevel:            cfg.LogLevel,
		LogDir:              cfg.LogDir,
		CacheDir:            cfg.CacheDir,
		ClusterName:         cfg.ClusterName,
	}

	cp := vo.NacosClientParam{
		ServerConfigs: scSli,
		ClientConfig:  &cc,
	}

	cCli, err := clients.NewConfigClient(cp)
	if err != nil {
		return nil, err
	}

	nCli, err := clients.NewNamingClient(cp)
	if err != nil {
		return nil, err
	}

	nc := &NacosClient{
		cCli: cCli,
		nCli: nCli,
	}
	return nc, nil
}

func (nc *NacosClient) GetNamingClient() naming_client.INamingClient {
	return nc.nCli
}

func (nc *NacosClient) GetConfigClient() config_client.IConfigClient {
	return nc.cCli
}

func (nc *NacosClient) OnCreated(_ chan<- error) {
}

func (nc *NacosClient) OnDispose(ctx context.Context) error {
	if nc.nCli != nil {
		nc.nCli.CloseClient()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:

	}

	if nc.cCli != nil {
		nc.cCli.CloseClient()
	}

	lg := mylog.AppLoggerWithStop()
	lg.Info().Msgf("nacos client stopped successfully")

	return nil
}
