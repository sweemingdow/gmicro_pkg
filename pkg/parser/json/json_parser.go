package json

import (
	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func Parse(data []byte, val any) error {
	return json.Unmarshal(data, val)
}

func Fmt(val any) ([]byte, error) {
	return json.Marshal(val)
}

func FmtPretty(val any) ([]byte, error) {
	if contents, err := json.MarshalIndent(val, "", "\t"); err != nil {
		return nil, err
	} else {
		return contents, nil
	}
}
