package regnacos

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/sweemingdow/gmicro_pkg/pkg/regdis"
	"github.com/sweemingdow/gmicro_pkg/pkg/regdis/extra/enacos"
	"github.com/sweemingdow/gmicro_pkg/pkg/utils"
)

type nacosRegistry struct {
	nCli naming_client.INamingClient
}

func NewNacosRegistry(nCli naming_client.INamingClient) regdis.Registry {
	return &nacosRegistry{
		nCli: nCli,
	}
}

func (nr *nacosRegistry) Register(rp regdis.RegisterParam) error {
	nc, err := enacos.UnPkgRegisterExtraParam(rp.Extra)

	if err != nil {
		return err
	}

	hp := utils.ExtractOneHp(rp.Addr)

	reg, err := nr.nCli.RegisterInstance(vo.RegisterInstanceParam{
		ServiceName: rp.ServiceName,
		Ip:          hp.Host,
		Port:        uint64(hp.Port),
		Weight:      float64(rp.Weight),
		Enable:      true,
		Healthy:     true,
		ClusterName: nc.ClusterName,
		GroupName:   nc.GroupName,
		Metadata:    rp.Metadata,
		Ephemeral:   true,
	})

	if err != nil {
		return err
	}

	if !reg {
		return fmt.Errorf("register to nacos failed with param:%v", rp)
	}

	return nil
}

func (nr *nacosRegistry) Deregister(dp regdis.DeregisterParam) error {
	nc, err := enacos.UnPkgDeregisterExtraParam(dp.Extra)

	if err != nil {
		return err
	}

	hp := utils.ExtractOneHp(dp.Addr)

	deReg, err := nr.nCli.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          hp.Host,
		Port:        uint64(hp.Port),
		ServiceName: dp.ServiceName,
		GroupName:   nc.GroupName,
		Cluster:     nc.ClusterName,
		Ephemeral:   true,
	})

	if err != nil {
		return err
	}

	if !deReg {
		return fmt.Errorf("deregister from nacos failed with param:%v", dp)
	}

	return nil
}
