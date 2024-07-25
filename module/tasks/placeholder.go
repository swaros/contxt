// Copyright (c) 2023 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
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
package tasks

import (
	"errors"
	"os"
	"strings"
)

type DefaultPhHandler struct {
	ph map[string]string
}

// this is a simple implementation of the PlaceHolder interface
// it is used to store and retrieve placeholder
// this implementation is not thread safe and should only be used for basic testing

func NewDefaultPhHandler() *DefaultPhHandler {
	return &DefaultPhHandler{
		ph: make(map[string]string),
	}
}

func (d *DefaultPhHandler) SetPH(key, value string) {
	d.ph[key] = value
}

func (d *DefaultPhHandler) AppendToPH(key, value string) bool {
	if _, ok := d.ph[key]; ok {
		d.ph[key] += value
		return true
	}
	return false
}

func (d *DefaultPhHandler) SetIfNotExists(key, value string) {
	if _, ok := d.ph[key]; !ok {
		d.ph[key] = value
	}
}

func (d *DefaultPhHandler) GetPHExists(key string) (string, bool) {
	if value, ok := d.ph[key]; ok {
		return value, true
	}
	return "", false
}

func (d *DefaultPhHandler) GetPH(key string) string {
	return d.ph[key]
}

func (d *DefaultPhHandler) GetPlaceHoldersFnc(inspectFunc func(phKey string, phValue string)) {
	for key, value := range d.ph {
		inspectFunc(key, value)
	}
}

func (d *DefaultPhHandler) HandlePlaceHolder(line string) string {
	return d.HandlePlaceHolderWithScope(line, d.ph)
}

func (d *DefaultPhHandler) HandlePlaceHolderWithScope(line string, scopeVars map[string]string) string {
	for key, value := range scopeVars {
		line = strings.ReplaceAll(line, key, value)
	}
	return line
}

func (d *DefaultPhHandler) ClearAll() {
	d.ph = make(map[string]string)
}

func (d *DefaultPhHandler) ExportVarToFile(variable string, filename string) error {
	strData := d.GetPH(variable)
	if strData == "" {
		return errors.New("variable " + variable + " can not be used for export to file. not exists or empty")
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err2 := f.WriteString(d.HandlePlaceHolder(strData)); err2 != nil {
		return err2
	}

	return nil
}
