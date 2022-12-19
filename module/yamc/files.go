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

import "os"

// NewYmacByFile loads file content and returns a new Ymac
func NewByFile(filename string, rdr DataReader) (*Yamc, error) {
	if data, err := os.ReadFile(filename); err == nil {
		yetAnohterMapConverter := New()
		err := yetAnohterMapConverter.Parse(rdr, data)
		return yetAnohterMapConverter, err
	} else {
		return &Yamc{}, err
	}
}

// NewYmacByYaml shortcut for reading Yaml File by using NewYmacByFile
func NewByYaml(filename string) (*Yamc, error) {
	return NewByFile(filename, NewYamlReader())
}

// NewYmacByJson json file loading shortcut
func NewByJson(filename string) (*Yamc, error) {
	return NewByFile(filename, NewJsonReader())
}
