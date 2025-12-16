package rpccall

import (
	"github.com/rs/zerolog"
	"github.com/sweemingdow/gmicro_pkg/pkg/myerr"
)

const (
	CallOk                 = "1"
	GeneralErr             = "0"
	ServerUnpredictableErr = "1000" // 无法预料, 未知错误
	ParamValidateErr       = "1100" // 参数验证失败
)

type RpcRespSimple struct {
	Code    string `json:"code,omitempty"`
	ErrDesc string `json:"errDesc,omitempty"`
	Msg     string `json:"msg,omitempty"`
}

func (resp RpcRespSimple) IsOk() bool {
	return resp.Code == CallOk
}

func (resp RpcRespSimple) IsNotOk() bool {
	return !resp.IsOk()
}

func SimpleOk() RpcRespSimple {
	return RpcRespSimple{
		Code: CallOk,
	}
}

func SimpleErrAll(code, desc, msg string) RpcRespSimple {
	return RpcRespSimple{
		Code:    code,
		ErrDesc: desc,
		Msg:     msg,
	}

}

func SimpleErrDesc(desc string) RpcRespSimple {
	return RpcRespSimple{
		Code:    GeneralErr,
		ErrDesc: desc,
	}
}

func SimpleErrCodeDesc(code, desc string) RpcRespSimple {
	return RpcRespSimple{
		Code:    code,
		ErrDesc: desc,
	}
}

func SimpleErrDescMsg(desc, msg string) RpcRespSimple {
	return SimpleErrAll(GeneralErr, desc, msg)
}

func SimpleUnpredictableErr(err error) RpcRespSimple {
	return RpcRespSimple{
		Code:    ServerUnpredictableErr,
		ErrDesc: err.Error(),
	}
}

func SimpleParamValidateErr(desc, msg string) RpcRespSimple {
	return RpcRespSimple{
		Code:    ParamValidateErr,
		ErrDesc: desc,
		Msg:     msg,
	}
}

type RpcRespWrapper[T any] struct {
	Code    string `json:"code,omitempty"`
	ErrDesc string `json:"errDesc,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Resp    T      `json:"resp,omitempty"`
}

func (resp RpcRespWrapper[T]) IsOk() bool {
	return resp.Code == CallOk
}

func (resp RpcRespWrapper[T]) IsNotOk() bool {
	return !resp.IsOk()
}

func (resp RpcRespWrapper[T]) OkOrErr() (T, error) {
	if resp.IsOk() {
		return resp.Resp, nil
	}

	var zero T
	return zero, myerr.NewRpcRespError(resp.Code, resp.ErrDesc, resp.Msg)
}

func Ok[T any](resp T) RpcRespWrapper[T] {
	return RpcRespWrapper[T]{
		Code: CallOk,
		Resp: resp,
	}
}

func ErrAll[T any](code, desc, msg string, resp T) RpcRespWrapper[T] {
	return RpcRespWrapper[T]{
		Code:    code,
		ErrDesc: desc,
		Msg:     msg,
		Resp:    resp,
	}
}

func ErrGeneralAll[T any](desc, msg string, resp T) RpcRespWrapper[T] {
	return RpcRespWrapper[T]{
		Code:    GeneralErr,
		ErrDesc: desc,
		Msg:     msg,
		Resp:    resp,
	}
}

func LoggerWrapWithResp[T any](reqId string, resp RpcRespWrapper[T], lg zerolog.Logger) zerolog.Logger {
	return lg.With().Str("req_id", reqId).Any("rpc_resp", resp).Logger()
}
