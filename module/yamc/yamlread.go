// Copyright (c) 2022 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// # Licensed under the MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package yamc

import (
	"errors"
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

func (y *YamlReader) FileDecode(path string, decodeInterface interface{}) (err error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if !IsPointer(decodeInterface) {
		return errors.New("decode will work on pointers only")

	}
	err2 := yaml.Unmarshal(file, decodeInterface)

	return err2
}

func (y *YamlReader) SupportsExt() []string {
	return []string{"yml", "yaml"}
}
