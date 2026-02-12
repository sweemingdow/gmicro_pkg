package myerr

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sweemingdow/gmicro_pkg/pkg/response"
)

type I18nError interface {
	CodeMsgError
	I18nTag() string
	I18nKey() string
}

type i18nError struct {
	i18nTag string // 区域, eg: en, zh, zh-CN
	i18nKey string // key, eg: _user_not_exists
	err     error  // 堆栈
}

func NewI18nError(i18nTag, i18nKey string) error {
	return i18nError{
		i18nTag: i18nTag,
		i18nKey: i18nKey,
		err:     errors.New(i18nKey),
	}
}

func (e i18nError) Error() string {
	return fmt.Sprintf("i18nError[i18nTag=%s, i18nKey=%s]", e.i18nTag, e.i18nKey)
}

func (e i18nError) ErrCode() string {
	return response.ApiGenErr
}

func (e i18nError) ErrMsg() string {
	return "M-Not Found"
}

func (e i18nError) I18nTag() string {
	return e.i18nTag
}

func (e i18nError) I18nKey() string {
	return e.i18nKey
}

func (e i18nError) Unwrap() error {
	return e.err
}
