package yamc

import (
	"os"

	"gopkg.in/yaml.v3"
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

func (y *YamlReader) FileDecode(path string, decodeInterface any) (err error) {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := yaml.NewDecoder(file)
	return decoder.Decode(&decodeInterface)
}

func (y *YamlReader) SupportsExt() []string {
	return []string{"yml", "yaml"}
}
