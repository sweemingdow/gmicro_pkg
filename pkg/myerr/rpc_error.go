package myerr

import (
	"errors"
	"fmt"
	"github.com/lesismal/arpc"
)

type RpcCallError struct {
	err     error
	timeout bool
}

func NewRpcCallError(err error) RpcCallError {
	timeout := err == arpc.ErrClientTimeout
	return RpcCallError{err: err, timeout: timeout}
}

func (rce RpcCallError) Error() string {
	return fmt.Sprintf("rpc_call_error[err:%v]", rce.err)
}

func (rce RpcCallError) Unwrap() error {
	return rce.err
}

func (rce RpcCallError) Timeout() bool {
	return rce.timeout
}

func IsRpcCallErr(err error) bool {
	var rce RpcCallError
	return errors.As(err, &rce)
}

func DecodeRpcCallErr(err error) (RpcCallError, bool) {
	var rce RpcCallError
	if errors.As(err, &rce) {
		return rce, true
	}

	return rce, false
}

type RpcBindError struct {
	err error
}

func NewRpcBindError(err error) RpcBindError {
	return RpcBindError{err: err}
}

func (rbe RpcBindError) Error() string {
	return fmt.Sprintf("rpc_bind_error[err:%v]", rbe.err)
}

func (rbe RpcBindError) Unwrap() error {
	return rbe.err
}

func IsRpcBindError(err error) bool {
	var rbe RpcBindError
	return errors.As(err, &rbe)
}

type RpcRespError struct {
	code    string
	ErrDesc string
	Msg     string
}

func NewRpcRespError(code, desc, msg string) RpcRespError {
	return RpcRespError{
		code:    code,
		ErrDesc: desc,
		Msg:     msg,
	}
}

func (rre RpcRespError) Code() string {
	return rre.code
}

func (rre RpcRespError) Error() string {
	return fmt.Sprintf("rpc_resp_err[code:%s, desc:%s, msg:%s]", rre.code, rre.ErrDesc, rre.Msg)
}

func DecodeRpcRespError(err error) (RpcRespError, bool) {
	var rre RpcRespError
	if errors.As(err, &rre) {
		return rre, true
	}

	return rre, false
}
