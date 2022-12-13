package yamc

import "gopkg.in/yaml.v3"

type YamlReader struct{}

func NewYamlReader() *YamlReader {
	return &YamlReader{}
}

func (y *YamlReader) Unmarshal(in []byte, out interface{}) (err error) {
	return yaml.Unmarshal(in, out)
}

func (y *YamlReader) Marshal(in interface{}) (out []byte, err error) {
	return yaml.Marshal(in)
}
