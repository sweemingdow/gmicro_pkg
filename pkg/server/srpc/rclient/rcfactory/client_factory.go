package rcfactory

import (
	"context"
	"github.com/sweemingdow/gmicro_pkg/pkg/lifetime"
	"github.com/sweemingdow/gmicro_pkg/pkg/server/srpc/rclient"
)

// arpc client abstract factory
type ArpcClientFactory interface {
	AcquireClient(serviceName string) rclient.ArpcClientProxy

	Stop(ctx context.Context) error
}

type ArpcClientFactoryLifecycle interface {
	lifetime.LifeCycle

	ArpcClientFactory
}
