package systools

import (
	"errors"
	"regexp"
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
	s = strings.ReplaceAll(s, ".", "-") // remove dots at least vor using configs versions
	if isMatch := regexp.MustCompile(`^[A-Za-z0-9_-]*$`).MatchString(s); isMatch {
		return s, nil
	}
	return "", errors.New("string contains not accepted chars ")
}
