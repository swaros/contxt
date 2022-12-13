package yamc

import "encoding/json"

type JsonReader struct{}

func NewJsonReader() *JsonReader {
	return &JsonReader{}
}

func (j *JsonReader) Unmarshal(in []byte, out interface{}) (err error) {
	return json.Unmarshal(in, out)
}

func (j *JsonReader) Marshal(in interface{}) (out []byte, err error) {
	return json.Marshal(in)
}
