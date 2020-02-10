package alimns

import (
	"encoding/json"
)

// Codec 编解码
type Codec interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
}

// JSONCodec json编解码
type JSONCodec struct {
}

// Encode 编码
func (codec JSONCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

// Decode 解码
func (codec JSONCodec) Decode(data []byte, value interface{}) error {
	return json.Unmarshal(data, value)
}
