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
	"strings"

	"gopkg.in/yaml.v3"
)

// YamlReader is a reader for yaml files
type YamlReader struct {
	fields *StructDef
}

// NewYamlReader creates a new YamlReader
func NewYamlReader() *YamlReader {
	return &YamlReader{
		fields: &StructDef{},
	}
}

// HaveFields returns true if the reader has field information
func (y *YamlReader) HaveFields() bool {
	return y.fields != nil && y.fields.Init
}

// GetFields returns the field information
func (y *YamlReader) GetFields() *StructDef {
	return y.fields
}

// Unmarshal unmarshals like yaml.Unmarshal
func (y *YamlReader) Unmarshal(in []byte, out interface{}) (err error) {
	y.fields = NewStructDef(out)
	if err := y.fields.ReadStruct(parseYamlTagfunc); err != nil {
		return err
	}
	return yaml.Unmarshal(in, out)
}

// Marshal marshals like yaml.Marshal
func (y *YamlReader) Marshal(in interface{}) (out []byte, err error) {
	return yaml.Marshal(in)
}

// FileDecode decodes a yaml file into a struct
func (y *YamlReader) FileDecode(path string, decodeInterface interface{}) (err error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if !IsPointer(decodeInterface) {
		return errors.New("decode will work on pointers only")

	}
	y.fields = NewStructDef(decodeInterface)
	if err := y.fields.ReadStruct(parseYamlTagfunc); err != nil {
		return err
	}
	err2 := yaml.Unmarshal(file, decodeInterface)

	return err2
}

// parser fuctions to resolve reflection tags for yaml struct tags
func parseYamlTagfunc(info StructField) ReflectTagRef {
	if info.Tag.Get("yaml") != "" {
		all := info.Tag.Get("yaml")
		parts := strings.Split(all, ",")
		adds := []string{}
		if len(parts) > 1 {
			adds = parts[1:]
		}
		return ReflectTagRef{
			TagRenamed:    parts[0],
			TagAdditional: adds,
		}
	}

	return ReflectTagRef{}
}

func (y *YamlReader) SupportsExt() []string {
	return []string{"yml", "yaml"}
}
