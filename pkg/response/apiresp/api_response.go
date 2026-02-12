package apiresp

import (
	"github.com/sweemingdow/gmicro_pkg/pkg/response"
	"time"
)

type ApiRWrapper[T any] struct {
	Code    string `json:"code,omitempty"`
	SubCode string `json:"subCode,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Data    T      `json:"data,omitempty"`
	Ts      int64  `json:"ts,omitempty"`
}

func JustOk() ApiRWrapper[any] {
	return ApiRWrapper[any]{
		Code: response.ApiOk,
		Ts:   time.Now().UnixMilli(),
	}
}

func Ok[T any](data T) ApiRWrapper[T] {
	return ApiRWrapper[T]{
		Code: response.ApiOk,
		Data: data,
		Ts:   time.Now().UnixMilli(),
	}
}

func GenResp(msg string) ApiRWrapper[any] {
	return ApiRWrapper[any]{
		Code: response.ApiGenErr,
		Msg:  msg,
		Ts:   time.Now().UnixMilli(),
	}
}

func GenSubResp(subCode, msg string) ApiRWrapper[any] {
	return ApiRWrapper[any]{
		Code:    response.ErrForSub,
		SubCode: subCode,
		Msg:     msg,
		Ts:      time.Now().UnixMilli(),
	}
}

func CodeResp(code string) ApiRWrapper[any] {
	return ApiRWrapper[any]{
		Code: code,
		Ts:   time.Now().UnixMilli(),
	}
}

func CodeMsgResp(code, msg string) ApiRWrapper[any] {
	return ApiRWrapper[any]{
		Code: code,
		Msg:  msg,
		Ts:   time.Now().UnixMilli(),
	}
}

func AllResp[T any](code, subCode, msg string, data T) ApiRWrapper[T] {
	return ApiRWrapper[T]{
		Code:    code,
		SubCode: subCode,
		Msg:     msg,
		Data:    data,
		Ts:      time.Now().UnixMilli(),
	}
}
