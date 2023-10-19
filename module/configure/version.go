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

	"github.com/Masterminds/semver/v3"
)

// undefined variables they will be
// set by the linker
// example: go build -ldflags "-X github.com/swaros/contxt/module/context/configure.minversion=1-alpha -X github.com/swaros/contxt/module/context/configure.midversion=0 -X github.com/swaros/contxt/module/context/configure.mainversion=0"
// for all the variables you can use
//     go tool nm bin/contxt | grep version
// to figure out how the variable can be set

var build string           // the build number
var mainversion string     // the main version (main).(mid).(min)
var midversion string      // the mid version
var minversion string      // the min version
var operatingSystem string // the operating system
var shortcut string        // the shortcut for the context functions in bash, fish, zsh and powershell

// GetVersion delivers the current build version
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

// GetShortcut delivers the current shortcut
// for the context functions in bash, fish, zsh and powershell
// if no shortcut is set, the default is "ctx"
func GetShortcut() string {
	if shortcut == "" {
		return "ctx"
	}
	return shortcut
}

// CheckCurrentVersion checks if the current version is greater or equal to the given version
func CheckCurrentVersion(askWithVersion string) bool {
	curentVersion := GetVersion()

	if curentVersion == "" || askWithVersion == "" {
		return true
	}
	return CheckVersion(askWithVersion, curentVersion)
}

// CheckVersion checks if the given version is greater or equal to the verifyAgainst version
func CheckVersion(askWithVersion string, verifyAgainst string) bool {
	current, err := semver.NewVersion(askWithVersion)
	if err != nil {
		return false
	}
	verifyVersion, verr := semver.NewVersion(verifyAgainst)
	if verr != nil {
		return false
	}

	if verifyVersion.GreaterThan(current) {
		return true
	}

	if current.Equal(verifyVersion) {
		return true
	}

	return false
}

// GetOs returns the current operating system
// if no operating system is set, the default is the runtime.GOOS
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
