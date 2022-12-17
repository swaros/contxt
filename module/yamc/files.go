package yamc

import "os"

// NewYmacByFile loads file content and returns a new Ymac
func NewYmacByFile(filename string, rdr DataReader) (*Yamc, error) {
	if data, err := os.ReadFile(filename); err == nil {
		yetAnohterMapConverter := NewYmac()
		err := yetAnohterMapConverter.Parse(rdr, data)
		return yetAnohterMapConverter, err
	} else {
		return &Yamc{}, err
	}
}

// NewYmacByYaml shortcut for reading Yaml File by using NewYmacByFile
func NewYmacByYaml(filename string) (*Yamc, error) {
	return NewYmacByFile(filename, NewYamlReader())
}

// NewYmacByJson json file loading shortcut
func NewYmacByJson(filename string) (*Yamc, error) {
	return NewYmacByFile(filename, NewJsonReader())
}
