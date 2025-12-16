package cauth

import (
	"github.com/sweemingdow/gmicro_pkg/pkg/server/srpc/rpccall"
)

type AuthReq struct {
	Token string `json:"token"`
}

type AuthResp struct {
	Uid string `json:"uid"`
}

type ClearReq struct {
	Uid string `json:"uid"`
}

type UpdateReq struct {
	Token      string `json:"token"`
	RealAuthed int    `json:"realAuthed"`
}

type PingReq struct {
	PingId string `json:"pingId"`
}

type PingResp struct {
	PongId string `json:"pongId"`
}

// @rpc_server two_service
//
// #go:generate go run ../../../../tools/rpcgen/main.go -type AuthRpcProvider -pkg_prefix github.com/sweemingdow/gmicro_pkg
//
//go:generate rpcgen -type AuthRpcProvider
type AuthRpcProvider interface {
	// @path /auth
	// @timeout 1s
	Auth(req rpccall.RpcReqWrapper[AuthReq]) (rpccall.RpcRespWrapper[AuthResp], error)

	// @timeout 2s
	Clear(req rpccall.RpcReqWrapper[ClearReq]) (rpccall.RpcRespSimple, error)

	// @path /update_token
	UpdateToken(req rpccall.RpcReqWrapper[UpdateReq]) (rpccall.RpcRespSimple, error)

	Ping(req PingReq) (PingResp, error)

	//NoArg() error
}
