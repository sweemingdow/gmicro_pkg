package myerr

import (
	"fmt"
	"github.com/lesismal/arpc"
	"github.com/pkg/errors"
	"github.com/sweemingdow/gmicro_pkg/pkg/response"
)

type RpcCallError interface {
	CodeMsgError
	IsTimeout() bool
}

type rpcCallError struct {
	err     error
	timeout bool
}

func NewRpcCallError(err error) RpcCallError {
	timeout := err == arpc.ErrClientTimeout

	return rpcCallError{
		err:     withStack(err),
		timeout: timeout,
	}
}

func AsRpcCallError(err error) (RpcCallError, bool) {
	var e RpcCallError
	if errors.As(err, &e) {
		return e, true
	}

	return e, false
}

func (e rpcCallError) Error() string {
	return fmt.Sprintf("rpcCallError[timeout=%t, err=%v]", e.timeout, e.err)
}

func (e rpcCallError) Unwrap() error {
	return e.err
}

func (e rpcCallError) ErrCode() string {
	return response.RpcCallErr
}

func (e rpcCallError) ErrMsg() string {
	return e.Error()
}

func (e rpcCallError) IsTimeout() bool {
	return e.timeout
}

func AsRpcCallErr(err error) (RpcCallError, bool) {
	var rce RpcCallError
	if errors.As(err, &rce) {
		return rce, true
	}

	return rce, false
}

type RpcRespError interface {
	CodeMsgError

	ErrDesc() string

	Data() any
}

type rpcRespError struct {
	code string
	desc string // 内部描述
	msg  string // 用户可见消息
	err  error  // origin error
	data any
}

func NewRpcRespErrorAll(code, desc, msg string, err error, data any) RpcRespError {
	return rpcRespError{
		code: code,
		desc: desc,
		msg:  msg,
		err:  withStack(err),
		data: data,
	}
}

func NewRpcRespError(code, desc, msg string, data any) RpcRespError {
	errMsg := desc
	if errMsg == "" {
		errMsg = msg
	}

	return rpcRespError{
		code: code,
		desc: desc,
		msg:  msg,
		err:  errors.New(errMsg),
		data: data,
	}
}

func AsRpcRespError(err error) (RpcRespError, bool) {
	var rre RpcRespError
	if errors.As(err, &rre) {
		return rre, true
	}

	return rre, false
}

func (e rpcRespError) Unwrap() error {
	return e.err
}

func (e rpcRespError) Error() string {
	return fmt.Sprintf("rpcRespError[code=%s, desc=%s, msg=%s, err=%v]", e.code, e.desc, e.msg, e.err)
}

func (e rpcRespError) ErrCode() string {
	return e.code
}

func (e rpcRespError) ErrDesc() string {
	return e.desc
}

func (e rpcRespError) ErrMsg() string {
	return e.msg
}

func (e rpcRespError) Data() any {
	return e.data
}
