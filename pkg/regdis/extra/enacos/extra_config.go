package enacos

import (
	"fmt"
	"github.com/sweemingdow/gmicro_pkg/pkg/regdis/extra"
)

const (
	NacosDefaultWeight      = 10.0
	nacosDefaultClusterName = "DEFAULT"
	nacosDefaultGroupName   = "DEFAULT_GROUP"
)

type NacosExtraConfig struct {
	ClusterName string
	GroupName   string
}

func PkgRegisterExtraParam(clusterName, groupName string) map[string]any {
	return pkgExtraParam(clusterName, groupName, extra.RegisterConfigKeyInExtra)
}

func PkgDeregisterExtraParam(clusterName, groupName string) map[string]any {
	return pkgExtraParam(clusterName, groupName, extra.DeregisterConfigKeyInExtra)
}

func PkgDiscoveryExtraParam(clusterName, groupName string) map[string]any {
	return pkgExtraParam(clusterName, groupName, extra.DiscoveryConfigKeyInExtra)
}

func pkgExtraParam(clusterName, groupName, extraKey string) map[string]any {
	if clusterName == "" {
		clusterName = nacosDefaultClusterName
	}

	if groupName == "" {
		groupName = nacosDefaultGroupName
	}
	return map[string]any{
		extraKey: NacosExtraConfig{
			ClusterName: clusterName,
			GroupName:   groupName,
		},
	}
}

func UnPkgRegisterExtraParam(extraMap map[string]any) (*NacosExtraConfig, error) {
	return unPkgExtraParam(extraMap, extra.RegisterConfigKeyInExtra)
}

func UnPkgDiscoveryExtraParam(extraMap map[string]any) (*NacosExtraConfig, error) {
	return unPkgExtraParam(extraMap, extra.DiscoveryConfigKeyInExtra)
}

func UnPkgDeregisterExtraParam(extraMap map[string]any) (*NacosExtraConfig, error) {
	return unPkgExtraParam(extraMap, extra.DeregisterConfigKeyInExtra)
}

func unPkgExtraParam(extraMap map[string]any, extraKey string) (*NacosExtraConfig, error) {
	nacosCfg, ok := extraMap[extraKey].(NacosExtraConfig)

	if !ok {
		return nil, fmt.Errorf("can not get NacosExtraConfig in Params Extra with:%s", extraKey)
	}

	return &nacosCfg, nil
}
