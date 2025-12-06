package regdis

import (
	"fmt"
	"gmicro_pkg/pkg/parser/json"
)

type DiscoverType uint8

const (
	ForHttp DiscoverType = 1
	ForRpc  DiscoverType = 2
)

type DiscoverParam struct {
	ServiceName string
	DisType     DiscoverType
	Extra       map[string]any
}

type Instance struct {
	InstanceId  string            `json:"instanceId,omitempty"`
	ServiceName string            `json:"serviceName,omitempty"`
	Ip          string            `json:"ip,omitempty"`
	Port        int               `json:"port,omitempty"`
	Weight      float64           `json:"weight,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

func (ins *Instance) InsIdentity() string {
	return fmt.Sprintf("%s:%d", ins.Ip, ins.Port)
}

type WatchFunc func(instances []*Instance, err error)

type UnwatchFunc func(err error)

type Discovery interface {
	Discover(dp DiscoverParam) ([]*Instance, error)

	Watch(dp DiscoverParam, wf WatchFunc) error

	Unwatch(dp DiscoverParam, uwf UnwatchFunc) error
}

func PrettyOutput(instances []*Instance) string {
	var output string
	data, err := json.Fmt(&instances)
	if err != nil {
		output = fmt.Sprintf("%+v", instances)
	}
	output = string(data)

	return output
}
