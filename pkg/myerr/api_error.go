package myerr

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sweemingdow/gmicro_pkg/pkg/response"
)

type apiError struct {
	code string
	msg  string
	err  error
}

func NewApiErr(code string, msg string) CodeMsgError {
	return apiError{
		code: code,
		msg:  msg,
		err:  errors.New(msg), // 携带堆栈
	}
}

func NewMsgApiErr(msg string) CodeMsgError {
	return apiError{
		code: response.ApiGenErr,
		msg:  msg,
		err:  errors.New(msg),
	}
}

func NewCodeApiErr(code string) CodeMsgError {
	return apiError{
		code: code,
		err:  errors.New("biz code err"),
	}
}

func NewCodeErrApiErr(code string, err error) CodeMsgError {
	return apiError{
		code: code,
		err:  withStack(err),
	}
}

func NewCodeMsgErrApiErr(code, msg string, err error) CodeMsgError {
	return apiError{
		code: code,
		msg:  msg,
		err:  withStack(err),
	}
}

func (e apiError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("apiError[code=%s, err=%v]", e.code, e.err)
	}

	return fmt.Sprintf("apiError[code=%s, msg=%s]", e.code, e.msg)
}

func (e apiError) ErrCode() string {
	return e.code
}

func (e apiError) ErrMsg() string {
	return e.msg
}

func (e apiError) Unwrap() error {
	return e.err
}
