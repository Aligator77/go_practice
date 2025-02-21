package helpers

import (
	"regexp"
	"strings"
)

func CheckFlag(str *string) bool {
	return strings.Contains(*str, ":")
}

func CheckFlagHttp(str *string) bool {
	matched, _ := regexp.MatchString("^http", *str)

	return strings.Contains(*str, ":") && matched
}
