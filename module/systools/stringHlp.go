package systools

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func StringSubLeft(path string, max int) string {
	if len := StrLen(path); len <= max {
		return path
	} else {
		runes := []rune(path)
		return string(runes[0:max])
	}
}

func StringSubRight(path string, max int) string {
	if len := StrLen(path); len <= max {
		return path
	} else {
		runes := []rune(path)
		left := len - max
		return string(runes[left:])
	}
}

func StrLen(str string) int {
	return len([]rune(str))
}

func CheckForCleanString(s string) (cleanString string, err error) {
	// replace some expected chars comming from version (dots) nd paths (/\:)
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ":", "--") // windows paths
	if isMatch := regexp.MustCompile(`^[A-Za-z0-9_-]*$`).MatchString(s); isMatch {
		return s, nil
	}
	return "", errors.New("string contains not accepted chars ")
}

func PrintableChars(str string) string {
	var result []rune
	for _, r := range str {
		if r > 31 && r < 127 {
			result = append(result, r)
		}
	}
	return string(result)
}

func PrintableCharsByUnquote(str string) string {
	s2, err := strconv.Unquote(`"` + str + `"`)
	if err != nil {
		panic(err)
	}
	return s2
}

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
	var args map[string]string = make(map[string]string)
	argArr := strings.Split(argLine, " ")
	for index, v := range argArr {
		args[fmt.Sprintf("%s%v", prefix, index)] = v
	}
	return argArr[0], args
}

func GetArgQuotedEntries(oristr string) ([]string, bool) {
	var result []string
	found := false
	re := regexp.MustCompile(`'[^']+'`)
	newStrs := re.FindAllString(oristr, -1)
	for _, s := range newStrs {
		found = true
		result = append(result, s)

	}
	return result, found
}

func SplitQuoted(oristr string, sep string) []string {
	var result []string
	var placeHolder map[string]string = make(map[string]string)

	found := false
	re := regexp.MustCompile(`'[^']+'`)
	newStrs := re.FindAllString(oristr, -1)
	i := 0
	for _, s := range newStrs {
		pl := "[$" + strconv.Itoa(i) + "]"
		placeHolder[pl] = strings.ReplaceAll(s, `'`, "")
		oristr = strings.Replace(oristr, s, pl, 1)
		found = true
		i++
	}
	result = strings.Split(oristr, sep)
	if !found {
		return result
	}

	for index, val := range result {
		if orgStr, fnd := placeHolder[val]; fnd {
			result[index] = orgStr
		}
	}

	return result
}

func AnyToStrNoTabs(any interface{}) string {
	return strings.ReplaceAll(fmt.Sprintf("%v", any), "\t", " ")
}
