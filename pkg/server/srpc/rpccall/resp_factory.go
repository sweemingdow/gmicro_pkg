package rpccall

import "github.com/sweemingdow/gmicro_pkg/pkg/response"

func NewCodecErrResp(err error) RpcRespWrapper[any] {
	return ErrAll[any](response.RpcCodecError, err.Error(), response.RpcCode2text(response.RpcCodecError), nil)
}

func NewUnpredictableErrResp(err error) RpcRespWrapper[any] {
	return ErrAll[any](response.RpcUnpredictableErr, err.Error(), response.RpcCode2text(response.RpcUnpredictableErr), nil)
}

func NewParamErrResp(err error) RpcRespWrapper[any] {
	return ErrAll[any](response.RpcParamError, err.Error(), response.RpcCode2text(response.RpcUnpredictableErr), nil)
}
