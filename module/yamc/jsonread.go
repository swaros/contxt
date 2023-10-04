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
	"encoding/json"
	"os"
	"strings"
)

type JsonReader struct {
	fields *StructDef
}

func NewJsonReader() *JsonReader {
	return &JsonReader{
		fields: &StructDef{},
	}
}

func (j *JsonReader) HaveFields() bool {
	return j.fields != nil && j.fields.Init
}

func (j *JsonReader) GetFields() *StructDef {
	return j.fields
}

func (j *JsonReader) Unmarshal(in []byte, out interface{}) (err error) {
	j.fields = NewStructDef(out)
	if err := j.fields.ReadStruct(parseJsonTagfunc); err != nil {
		return err
	}
	return json.Unmarshal(in, out)
}

func (j *JsonReader) Marshal(in interface{}) (out []byte, err error) {
	return json.Marshal(in)
}

func (j *JsonReader) FileDecode(path string, decodeInterface interface{}) (err error) {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	j.fields = NewStructDef(decodeInterface)
	if err := j.fields.ReadStruct(parseYamlTagfunc); err != nil {
		return err
	}
	return decoder.Decode(&decodeInterface)
}

func (j *JsonReader) SupportsExt() []string {
	return []string{"json"}
}

// parser fuctions to resolve reflection tags for yaml struct tags
func parseJsonTagfunc(info StructField) ReflectTagRef {
	if info.Tag.Get("json") != "" {
		all := info.Tag.Get("json")
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
