package codec

import jsoniter "github.com/json-iterator/go"

var _json = jsoniter.ConfigCompatibleWithStandardLibrary

type JsoniterCodec struct{}

func (j *JsoniterCodec) Marshal(v any) ([]byte, error) {
	return _json.Marshal(v)
}

func (j *JsoniterCodec) Unmarshal(data []byte, v any) error {
	return _json.Unmarshal(data, v)
}
