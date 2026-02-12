package rpccall

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/sweemingdow/gmicro_pkg/pkg/myerr"
	"github.com/sweemingdow/gmicro_pkg/pkg/response"
)

type RpcRespWrapper[T any] struct {
	Code    string `json:"code,omitempty"`
	ErrDesc string `json:"errDesc,omitempty"` // 调试描述
	Msg     string `json:"msg,omitempty"`     // 展示信息
	Data    T      `json:"data,omitempty"`
}

func (r RpcRespWrapper[T]) IsOk() bool {
	return r.Code == response.RpcOk
}

func (r RpcRespWrapper[T]) IsNotOk() bool {
	return !r.IsOk()
}

func (r RpcRespWrapper[T]) String() string {
	return fmt.Sprintf("RpcRespWrapper[code=%s, desc=%s, msg=%s, data=%+v]", r.Code, r.ErrDesc, r.Msg, r.Data)
}

func (r RpcRespWrapper[T]) OkOrErr() (T, myerr.RpcRespError) {
	if r.IsOk() {
		return r.Data, nil
	}

	return r.Data, myerr.NewRpcRespError(r.Code, r.ErrDesc, r.Msg, r.Data)
}

func (r RpcRespWrapper[T]) OkOrTake() (T, bool) {
	if r.IsOk() {
		return r.Data, true
	}

	return r.Data, false
}

func Ok[T any](data T) RpcRespWrapper[T] {
	return RpcRespWrapper[T]{
		Code: response.RpcOk,
		Data: data,
	}
}

func JustOk() RpcRespWrapper[any] {
	return RpcRespWrapper[any]{
		Code: response.RpcOk,
	}
}

func JustErr() RpcRespWrapper[any] {
	return RpcRespWrapper[any]{
		Code: response.RpcGenErr,
	}
}

func ErrAll[T any](code, desc, msg string, data T) RpcRespWrapper[T] {
	return RpcRespWrapper[T]{
		Code:    code,
		ErrDesc: desc,
		Msg:     msg,
		Data:    data,
	}
}

func GenErrAll[T any](desc, msg string, data T) RpcRespWrapper[T] {
	return RpcRespWrapper[T]{
		Code:    response.RpcOk,
		ErrDesc: desc,
		Msg:     msg,
		Data:    data,
	}
}

func GenErr[T any](desc string, data T) RpcRespWrapper[T] {
	return RpcRespWrapper[T]{
		Code:    response.RpcGenErr,
		ErrDesc: desc,
		Data:    data,
	}
}

func LoggerWrapWithResp[T any](reqId string, resp RpcRespWrapper[T], lg zerolog.Logger) zerolog.Logger {
	return lg.With().Str("req_id", reqId).Any("rpc_resp", resp).Logger()
}
