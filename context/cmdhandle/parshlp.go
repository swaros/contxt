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
