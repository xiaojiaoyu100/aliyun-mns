package alimns

import (
	"github.com/json-iterator/go"
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
	var j = jsoniter.ConfigCompatibleWithStandardLibrary
	return j.Marshal(&value)
}

// Decode 解码
func (codec JSONCodec) Decode(data []byte, value interface{}) error {
	var j = jsoniter.ConfigCompatibleWithStandardLibrary
	return j.Unmarshal(data, &value)
}
