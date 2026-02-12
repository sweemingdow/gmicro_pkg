package response

const (
	RpcOk = "1"

	RpcGenErr = "0"
)

const (
	RpcUnpredictableErr = "999999" // 无法预料, 未知错误

	// 编解码错误
	RpcCodecError = "999998"

	// 参数错误
	RpcParamError = "999997"
)

var rpcCode2desc = map[string]string{
	RpcUnpredictableErr: "unpredictable error",
	RpcCodecError:       "codec error",
	RpcParamError:       "param invalid",
}

func RpcCode2text(code string) string {
	return rpcCode2desc[code]
}
