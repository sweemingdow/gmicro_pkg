package myerr

import (
	"github.com/pkg/errors"
)

type CodeMsgError interface {
	error
	ErrCode() string
	ErrMsg() string
}

type SubCodeMsgError interface {
	CodeMsgError
	SubCode() string
}

func AsCodeMsgError(err error) (CodeMsgError, bool) {
	var e CodeMsgError
	if errors.As(err, &e) {
		return e, true
	}

	return e, false
}

func AsSubCodeError(err error) (SubCodeMsgError, bool) {
	var e SubCodeMsgError
	if errors.As(err, &e) {
		return e, true
	}

	return e, false
}

func withStack(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := err.(interface{ StackTrace() errors.StackTrace }); ok {
		return err
	}

	return errors.WithStack(err)
}
