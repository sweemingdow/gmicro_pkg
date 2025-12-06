package disnacos

import (
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/sweemingdow/gmicro_pkg/pkg/regdis"
	"github.com/sweemingdow/gmicro_pkg/pkg/regdis/extra/enacos"
	"github.com/sweemingdow/gmicro_pkg/pkg/utils/usli"
	"strconv"
)

type nacosDiscovery struct {
	nCli naming_client.INamingClient
}

func NewNacosDiscovery(nCli naming_client.INamingClient) regdis.Discovery {
	return &nacosDiscovery{
		nCli: nCli,
	}
}

func (nd *nacosDiscovery) Discover(dp regdis.DiscoverParam) ([]*regdis.Instance, error) {
	ep, err := enacos.UnPkgDiscoveryExtraParam(dp.Extra)
	if err != nil {
		return nil, err
	}

	instances, err := nd.nCli.SelectInstances(vo.SelectInstancesParam{
		Clusters:    []string{ep.ClusterName},
		ServiceName: dp.ServiceName,
		GroupName:   ep.GroupName,
		HealthyOnly: true,
	})

	if err != nil {
		return nil, err
	}

	return convertIns(instances, dp.DisType), nil
}

func (nd *nacosDiscovery) Watch(dp regdis.DiscoverParam, wf regdis.WatchFunc) error {
	sp, err := dp2sp(dp)
	if err != nil {
		return err
	}

	sp.SubscribeCallback = func(instances []model.Instance, err error) {
		wf(convertIns(instances, dp.DisType), err)
	}

	err = nd.nCli.Subscribe(sp)

	if err != nil {
		return err
	}

	return nil
}

func (nd *nacosDiscovery) Unwatch(dp regdis.DiscoverParam, uwf regdis.UnwatchFunc) error {
	sp, err := dp2sp(dp)
	if err != nil {
		return err
	}

	sp.SubscribeCallback = func(instances []model.Instance, err error) {
		uwf(err)
	}

	err = nd.nCli.Unsubscribe(sp)

	if err != nil {
		return err
	}

	return nil
}

func dp2sp(dp regdis.DiscoverParam) (*vo.SubscribeParam, error) {
	ep, err := enacos.UnPkgDiscoveryExtraParam(dp.Extra)
	if err != nil {
		return nil, err
	}

	return &vo.SubscribeParam{
		ServiceName: dp.ServiceName,
		Clusters:    []string{ep.ClusterName},
		GroupName:   ep.GroupName,
	}, nil
}

func convertIns(instances []model.Instance, disType regdis.DiscoverType) []*regdis.Instance {
	return usli.Conv(instances, func(srcIns model.Instance) *regdis.Instance {
		md := srcIns.Metadata

		var port = int(srcIns.Port)
		if disType == regdis.ForHttp {
			p, err := strconv.Atoi(md["http_port"])
			if err == nil {
				port = p
			}
		} else if disType == regdis.ForRpc {
			p, err := strconv.Atoi(md["rpc_port"])
			if err == nil {
				port = p
			}
		}

		return &regdis.Instance{
			InstanceId:  srcIns.InstanceId,
			ServiceName: srcIns.ServiceName,
			Ip:          srcIns.Ip,
			Port:        port,
			Weight:      srcIns.Weight,
			Metadata:    md,
		}
	})
}
