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
package yacl

// SupportMigrate apply a callback they will be executed if the config is loaded somehow.
// this means for any config that is loaded by the defined rules.
// that is different to the config that is used at the end,
// because yacl can loads a couple of config files and just use the last one.

func (c *ConfigModel) SupportMigrate(fileHandelFn func(path string, cfg interface{})) *ConfigModel {
	c.fileLoadCallback = fileHandelFn
	c.supportMigrate = true
	return c
}
