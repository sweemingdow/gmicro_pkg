package response

// api提示
const (
	ApiOk = "1"

	ApiGenErr = "0"
)

const (
	UnpredictableErr = "9999"

	AccessReject = "9998"

	Forbidden = "9997"

	UnAuthErr = "9996"

	RpcCallErr = "9995"

	RpcRespErr = "9994"

	// 请求验证不通过
	VerifyErr = "2000"

	// 编解码错误(json)
	CodecError = "2001"

	// 参数验证错误
	ParamInvalidErr = "2002"

	// 调用第三方错误
	ThirdCallErr = "2003"

	// 第三方响应错误
	ThirdRespErr = "2004"

	// 请求频繁
	FrequentReqErr = "2005"
)

var apiCode2desc = map[string]string{
	UnpredictableErr: "unpredictable error",
	AccessReject:     "access rejected",
	Forbidden:        "forbidden",
	UnAuthErr:        "unauthorized",
	RpcCallErr:       "inner call error",
	RpcRespErr:       "inner resp error",
	VerifyErr:        "verify error",
	CodecError:       "codec error",
	ParamInvalidErr:  "param invalid",
	ThirdCallErr:     "third call error",
	ThirdRespErr:     "third resp error",
	FrequentReqErr:   "frequent req err",
}

func ApiCode2text(code string) string {
	return apiCode2desc[code]
}
