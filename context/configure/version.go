package configure

import (
	"runtime"
	"strings"
)

// undefined variables they will be
// set by the linker
// example: go build -ldflags "-X github.com/swaros/contxt/context/configure.minversion=1-alpha -X github.com/swaros/contxt/context/configure.midversion=0 -X github.com/swaros/contxt/context/configure.mainversion=0"
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
