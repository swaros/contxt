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

package configure

import (
	"runtime"
	"strings"
)

// undefined variables they will be
// set by the linker
// example: go build -ldflags "-X github.com/swaros/contxt/module/context/configure.minversion=1-alpha -X github.com/swaros/contxt/module/context/configure.midversion=0 -X github.com/swaros/contxt/module/context/configure.mainversion=0"
// for all the variables you can use
//     go tool nm bin/contxt | grep version
// to figure out how the variable can be set

var build string
var mainversion string
var midversion string
var minversion string
var operatingSystem string

// GetVersion deleivers the current build version
func GetVersion() string {
	var outVersion = ""
	if mainversion != "" {
		outVersion = mainversion
	}
	if midversion != "" {
		if outVersion != "" {
			outVersion = outVersion + "." + midversion
		} else {
			outVersion = midversion
		}
	}

	if minversion != "" {
		if outVersion != "" {
			outVersion = outVersion + "." + minversion
		} else {
			outVersion = minversion
		}
	}

	return outVersion
}

func CheckCurrentVersion(askWithVersion string) bool {
	curentVersion := GetVersion()

	if curentVersion == "" || askWithVersion == "" {
		return true
	}
	return CheckVersion(askWithVersion, curentVersion)
}

func CheckVersion(askWithVersion string, versionStr string) bool {
	return askWithVersion <= versionStr // til now this seems simple enough
}

func GetOs() string {
	if operatingSystem == "" {
		return strings.ToLower(runtime.GOOS)
	}
	return strings.ToLower(operatingSystem)
}

// GetBuild returns Build time as build NO
func GetBuild() string {
	return build
}
