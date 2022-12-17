package yamc

import (
	"os"

	"gopkg.in/yaml.v2"
)

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

func (y *YamlReader) FileDecode(path string, decodeInterface interface{}) (err error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(file, decodeInterface)
}

func (y *YamlReader) SupportsExt() []string {
	return []string{"yml", "yaml"}
}
