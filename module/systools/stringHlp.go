package systools

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
