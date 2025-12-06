package cfgnacos

import (
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// Nacos配置中心
type NacosConfigCenter struct {
	cCli config_client.IConfigClient
}

type AcquireParam struct {
	CfgId     string
	GroupName string
}

func NewNacosConfigCenter(cCli config_client.IConfigClient) *NacosConfigCenter {
	return &NacosConfigCenter{
		cCli: cCli,
	}
}

// 获取配置
func (ncc *NacosConfigCenter) Acquire(ap AcquireParam) (string, error) {
	return ncc.cCli.GetConfig(vo.ConfigParam{
		DataId: ap.CfgId,
		Group:  ap.GroupName,
	})
}

type AcquireListenParam struct {
	CfgId     string
	GroupName string
	OnChanged func(namespace, group, dataId, data string)
}

// 获取配置并监听
func (ncc *NacosConfigCenter) AcquireAndListen(alp AcquireListenParam) (string, bool, error) {
	// 先获取一次
	content, err := ncc.cCli.GetConfig(vo.ConfigParam{
		DataId: alp.CfgId,
		Group:  alp.GroupName,
	})

	if err != nil {
		return "", false, err
	}

	// 监听配置变化
	err = ncc.cCli.ListenConfig(vo.ConfigParam{
		DataId:   alp.CfgId,
		Group:    alp.GroupName,
		OnChange: alp.OnChanged,
	})

	if err != nil {
		return content, true, err
	}

	return content, true, nil
}

func (ncc *NacosConfigCenter) UnListen(ap AcquireParam) error {
	return ncc.cCli.CancelListenConfig(vo.ConfigParam{
		DataId: ap.CfgId,
		Group:  ap.GroupName,
	})
}
