package cone

import (
	"gmicro_pkg/pkg/server/srpc/rpccall"
)

type OneReq struct {
	Name string `json:"token"`
}

type OneResp struct {
	Name string `json:"uid"`
}

type OneOneReq struct {
	Name string `json:"uid"`
}

type OneOneOneReq struct {
	Name string `json:"token"`
}

// @rpc_server one_service
//
//go:generate rpcgen -type OneRpcProvider
type OneRpcProvider interface {
	// @path /v1/one
	// @timeout 3s
	One(req rpccall.RpcReqWrapper[OneReq]) (rpccall.RpcRespWrapper[OneResp], error)

	// @path /v1/one_one
	OneOne(req rpccall.RpcReqWrapper[OneOneReq]) (rpccall.RpcRespSimple, error)

	// @path /v1/one_one_one
	OneOneOne(req rpccall.RpcReqWrapper[OneOneOneReq]) (rpccall.RpcRespSimple, error)
}

type OneStrongReq struct {
	Name string `json:"token"`
}

type OneStrongResp struct {
	Name string `json:"uid"`
}

// @rpc_server one_service
//
//go:generate rpcgen -type OneRpcStrongProvider
type OneRpcStrongProvider interface {
	// @path /v1/one_strong
	// @timeout 3s
	OneStrong(req rpccall.RpcReqWrapper[OneStrongReq]) (rpccall.RpcRespWrapper[OneStrongResp], error)
}
