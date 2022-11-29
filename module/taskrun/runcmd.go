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

// collection of comands they can be used for any comand interpreter like cobra and ishell.
package taskrun

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/configure"
	"github.com/swaros/contxt/dirhandle"
	"github.com/swaros/manout"
)

func DirFind(args []string) string {
	useIndex := -1
	usePath := "."
	if len(args) == 0 {
		dirhandle.PrintDir(configure.UsedConfig.LastIndex)
	} else {
		configure.PathWorker(func(index int, path string) {
			for _, search := range args {
				found := strings.Contains(path, search)
				if found {
					useIndex = index
					usePath = path
					GetLogger().WithFields(logrus.Fields{"index": useIndex, "path": usePath}).Debug("Found match by comparing strings")
				} else {
					// this part is not found. but maybe it is a index number?
					sIndex, err := strconv.Atoi(search)
					if err == nil && index == sIndex {
						useIndex = index
						usePath = path
						GetLogger().WithFields(logrus.Fields{"index": useIndex, "path": usePath}).Debug("Found match by using param as index")
					}
				}
			}
		})

		if useIndex >= 0 && useIndex != configure.UsedConfig.LastIndex {
			configure.UsedConfig.LastIndex = useIndex
			configure.SaveDefaultConfiguration(true)
		}

		fmt.Println(usePath)

	}
	return usePath
}

func PrintCnPaths(hints bool) {
	fmt.Println(manout.MessageCln("\t", "paths stored in ", manout.ForeCyan, configure.UsedConfig.CurrentSet))
	dir, err := dirhandle.Current()
	if err == nil {
		count := ShowPaths(dir)
		if count > 0 && hints {
			fmt.Println()
			fmt.Println(manout.MessageCln("\t", "if you have installed the shell functions ", manout.ForeDarkGrey, "(contxt install bash|zsh|fish)", manout.CleanTag, " change the directory by ", manout.BoldTag, "cn ", count-1))
			fmt.Println(manout.MessageCln("\t", "this will be the same as ", manout.BoldTag, "cd ", dirhandle.GetDir(count-1)))
		}
	}
}
