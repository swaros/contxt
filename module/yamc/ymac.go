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
	"errors"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

const (
	TYPE_ARRAY      = 1 // source type []interface{}
	TYPE_STRING_MAP = 2 // ... map[string]interface{}
	UNSET           = 0 // initial status
)

type DataReader interface {
	Unmarshal(in []byte, out interface{}) (err error)
	Marshal(in interface{}) (out []byte, err error)
	FileDecode(path string, decodeInterface interface{}) (err error)
	SupportsExt() []string
}

type Yamc struct {
	data           map[string]interface{} // holds the data after parse
	loaded         bool                   // is true if we at least tried to get data and got no error (can still be empty)
	sourceDataType int                    // holds the information about the structure of the source
}

func New() *Yamc {
	return &Yamc{
		loaded:         false,
		sourceDataType: 0,
		data:           make(map[string]interface{}),
	}
}

// GetSourceDataType returns the flag what tells us how the sourece was
// stuctured
// yamc.TYPE_ARRAY      = []interface{}
// yamc.TYPE_STRING_MAP = map[string]interface{}
// yamc.UNSET           = current nothing is loaded. so we have no type
func (y *Yamc) GetSourceDataType() int {
	return y.sourceDataType
}

// IsLoaded returns the loaded flag. this is not the same as having data
// it just means it is read without having errors
func (y *Yamc) IsLoaded() bool {
	return y.loaded
}

// Parse is wrapping the Unmarshal for json and yaml.
// because the default format is map[string]interface{}
// it fallback to read []interface{} and convert them.
func (y *Yamc) Parse(use DataReader, in []byte) error {
	y.Reset()
	if err := use.Unmarshal([]byte(in), &y.data); err != nil {
		return y.testAndConvertJsonType(use, in)
	}
	y.sourceDataType = TYPE_STRING_MAP
	y.loaded = true
	return nil

}

func (y *Yamc) SetData(data map[string]interface{}) {
	y.Reset()
	y.data = data
}

// GetData is just the getter for the actual
// data. this is independend if they are loaded or not
func (y *Yamc) GetData() map[string]interface{} {
	return y.data
}

// ToString uses the reader to create the string output of the
// data content
func (y *Yamc) ToString(use DataReader) (str string, err error) {
	if byteData, err := use.Marshal(y.data); err != nil {
		return "", err
	} else {
		return string(byteData), nil
	}
}

// Resets the whole Ymac
func (y *Yamc) Reset() {
	y.loaded = false
	y.data = make(map[string]interface{})
	y.sourceDataType = UNSET
}

// Gson wrapps gson and rerurns the gsonResult or
// the error while using Marshall the data into json
// what can be used by gson
func (y *Yamc) Gjson(path string) (gjson.Result, error) {
	jsonData, err := json.Marshal(y.data)
	if err == nil {
		return gjson.Get(string(jsonData), path), nil
	}
	return gjson.Result{}, err
}

// GetGjsonString returns the content of the path as json string result
// or the error while processing the data
func (y *Yamc) GetGjsonString(path string) (jsonStr string, err error) {
	if result, err := y.Gjson(path); err == nil {
		return result.String(), nil
	} else {
		return "", err
	}
}

// GetGjsonString returns the content of the path as json string result
// or the error while processing the data
func (y *Yamc) FindValue(path string) (content any, err error) {

	return FindChain(y.data, strings.Split(path, ".")...)

}

// testAndConvertJsonType is the Fallback Reader for []interface{}
// it converts the data to map[string]interface{} and uses the numeric index
// as string (key)
// so [{"hallo": "a"},{"welt":"b"}] will be afterwards {"0":{"hallo": "a"},"1":{"welt","b"}}
func (y *Yamc) testAndConvertJsonType(use DataReader, data []byte) error {
	var m []interface{}

	if err := use.Unmarshal([]byte(data), &m); err == nil {
		for key, val := range m {
			keyStr := fmt.Sprintf("%d", key)
			switch val.(type) {
			case string, interface{}:
				y.data[keyStr] = val
			default:
				y.loaded = false
				return errors.New("unsupported json structure")
			}
		}
		y.sourceDataType = TYPE_ARRAY
		y.loaded = true
		return nil
	} else {
		return err
	}
}
