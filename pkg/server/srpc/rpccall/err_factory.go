package rpccall

import (
	"github.com/sweemingdow/gmicro_pkg/pkg/myerr"
	"github.com/sweemingdow/gmicro_pkg/pkg/response"
)

func NewRpcUnpredictableErr(err error) error {
	return myerr.NewRpcRespError(response.RpcUnpredictableErr, err.Error(), response.RpcCode2text(response.RpcUnpredictableErr), nil)
}

func NewRpcParamInvalidErr(desc string) error {
	return myerr.NewRpcRespError(response.RpcParamError, desc, response.RpcCode2text(response.RpcParamError), nil)

}

func NewRpcCodecErr(err error) error {
	return myerr.NewRpcRespError(response.RpcCodecError, err.Error(), response.RpcCode2text(response.RpcCodecError), nil)
}
