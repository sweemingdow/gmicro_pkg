package apicall

import (
	"github.com/sweemingdow/gmicro_pkg/pkg/myerr"
	"github.com/sweemingdow/gmicro_pkg/pkg/response"
)

func NewParamInvalidErr(msg string) error {
	return myerr.NewApiErr(response.ParamInvalidErr, msg)
}

func NewNoPermissionErr(msg string) error {
	if msg == "" {
		msg = response.ApiCode2text(response.NoPermission)
	}

	return myerr.NewApiErr(response.NoPermission, msg)
}
