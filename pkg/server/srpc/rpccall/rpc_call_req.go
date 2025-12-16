package rpccall

import (
	"github.com/rs/zerolog"
	"github.com/sweemingdow/gmicro_pkg/pkg/utils"
	"time"
)

type RpcReqWrapper[T any] struct {
	ReqId     string `json:"reqId,omitempty"`
	I18nTag   string `json:"i18nTag,omitempty"`
	SendMills int64  `json:"sendMills,omitempty"`
	Req       T      `json:"req,omitempty"`
}

func CreateIdReq[T any](reqId string, req T) RpcReqWrapper[T] {
	if reqId == "" {
		reqId = utils.RandStr(32)
	}

	return RpcReqWrapper[T]{
		ReqId:     reqId,
		SendMills: time.Now().UnixMilli(),
		Req:       req,
	}
}

func CreateReq[T any](req T) RpcReqWrapper[T] {
	return CreateIdReq[T]("", req)
}

func CreateReqAll[T any](reqId, i18nTag string, req T) RpcReqWrapper[T] {
	if reqId == "" {
		reqId = utils.RandStr(32)
	}

	return RpcReqWrapper[T]{
		ReqId:     reqId,
		SendMills: time.Now().UnixMilli(),
		I18nTag:   i18nTag,
		Req:       req,
	}
}

func CreateI18nReq[T any](i18nTag string, req T) RpcReqWrapper[T] {
	return CreateReqAll[T]("", i18nTag, req)
}

func LoggerWrapWithReq[T any](req RpcReqWrapper[T], lg zerolog.Logger) zerolog.Logger {
	return lg.With().Str("req_id", req.ReqId).Any("rpc_req", req).Logger()
}

func LoggerWrapWithReqId(reqId string, lg zerolog.Logger) zerolog.Logger {
	return lg.With().Str("req_id", reqId).Logger()
}
