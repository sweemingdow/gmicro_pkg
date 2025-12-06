package cauth

import (
	"gmicro_pkg/pkg/server/srpc/rclient"
	"gmicro_pkg/pkg/server/srpc/rclient/rcfactory"
	"gmicro_pkg/pkg/server/srpc/rpccall"
	"time"
)

const (
	rpcServerName = "auth_service"
	defaultTimout = 2 * time.Second
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

type AuthRpcProvider interface {
	Auth(req rpccall.RpcReqWrapper[AuthReq]) (rpccall.RpcRespWrapper[AuthResp], error)

	Clear(req rpccall.RpcReqWrapper[ClearReq]) (rpccall.RpcRespSimple, error)

	UpdateToken(req rpccall.RpcReqWrapper[UpdateReq]) (rpccall.RpcRespSimple, error)
}

type authRpcProvider struct {
	clientFactory rcfactory.ArpcClientFactory
}

func NewAuthRpcProvider(clientFactory rcfactory.ArpcClientFactory) AuthRpcProvider {
	return &authRpcProvider{
		clientFactory: clientFactory,
	}
}

func (arp *authRpcProvider) Auth(req rpccall.RpcReqWrapper[AuthReq]) (rpccall.RpcRespWrapper[AuthResp], error) {
	cp := arp.acquireClientProxy()

	var resp rpccall.RpcRespWrapper[AuthResp]
	if err := cp.Call("/auth", &req, &resp, defaultTimout); err != nil {
		return resp, err
	}

	return resp, nil
}

func (arp *authRpcProvider) Clear(req rpccall.RpcReqWrapper[ClearReq]) (rpccall.RpcRespSimple, error) {
	cp := arp.acquireClientProxy()

	var resp rpccall.RpcRespSimple
	if err := cp.Call("/clear", &req, &resp, defaultTimout); err != nil {
		return resp, err
	}

	return resp, nil
}

func (arp *authRpcProvider) UpdateToken(req rpccall.RpcReqWrapper[UpdateReq]) (rpccall.RpcRespSimple, error) {
	cp := arp.acquireClientProxy()

	var resp rpccall.RpcRespSimple

	if err := cp.Call("/update_token", &req, &resp, defaultTimout); err != nil {
		return resp, err
	}

	return resp, nil
}

func (arp *authRpcProvider) acquireClientProxy() rclient.ArpcClientProxy {
	return arp.clientFactory.AcquireClient(rpcServerName)
}
