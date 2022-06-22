// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Licensed under the MIT License
//
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
package cmdhandle

import (
	"fmt"
	"strings"
)

func SplitArgs(cmdList []string, prefix string, arghandler func(string, map[string]string)) []string {
	var cleared []string
	var args map[string]string = make(map[string]string)

	for _, value := range cmdList {
		argArr := strings.Split(value, " ")
		cleared = append(cleared, argArr[0])
		if len(argArr) > 1 {
			for index, v := range argArr {
				args[fmt.Sprintf("%s%v", prefix, index)] = v
			}
			arghandler(argArr[0], args)
		}
	}
	return cleared
}

func StringSplitArgs(argLine string, prefix string) (string, map[string]string) {
	GetLogger().WithField("args", argLine).Debug("parsing argumented string")
	var args map[string]string = make(map[string]string)
	argArr := strings.Split(argLine, " ")
	for index, v := range argArr {
		args[fmt.Sprintf("%s%v", prefix, index)] = v
	}
	return argArr[0], args
}
