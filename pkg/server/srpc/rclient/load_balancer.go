package rclient

import "sync/atomic"

type LoadBalancer interface {
	Select(clients []*arpcClientWrap) *arpcClientWrap
}

type roundRobinLb struct {
	counter atomic.Uint64
}

// warning: Under concurrency protection, operate Select
func (rb *roundRobinLb) Select(clients []*arpcClientWrap) *arpcClientWrap {
	if len(clients) == 0 {
		return nil
	}

	return clients[rb.counter.Add(1)%uint64(len(clients))]
}

func NewRoundRobinLoadBalancer() LoadBalancer {
	return &roundRobinLb{}
}
