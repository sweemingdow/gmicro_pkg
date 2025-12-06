package yaml

import (
	"bytes"
	"gopkg.in/yaml.v3"
)

func Parse(data []byte, val any) error {
	return yaml.Unmarshal(data, val)
}

func Fmt(val any) ([]byte, error) {
	return yaml.Marshal(val)
}

func FmtPretty(val any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(val); err != nil {
		return nil, err
	} else {
		return buf.Bytes(), nil
	}
}
