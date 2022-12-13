package yamc

import (
	"encoding/json"
	"os"
)

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

func (j *JsonReader) FileDecode(path string, decodeInterface any) (err error) {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	return decoder.Decode(&decodeInterface)
}

func (j *JsonReader) SupportsExt() []string {
	return []string{"json"}
}
