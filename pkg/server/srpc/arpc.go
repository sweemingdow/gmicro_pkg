package srpc

import "gmicro_pkg/pkg/parser/json"

type JsonIterCodec struct {
}

func (jic JsonIterCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Fmt(v)
}

func (jic JsonIterCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Parse(data, v)
}
