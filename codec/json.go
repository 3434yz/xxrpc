package codec

import "encoding/json"

type JSONCodec struct{}

func (j *JSONCodec) Encode(v any) ([]byte, error) {
    return json.Marshal(v)
}

func (j *JSONCodec) Decode(data []byte, v any) error {
    return json.Unmarshal(data, v)
}
