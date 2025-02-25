package helpers

import (
	"regexp"
	"strings"
)

func CheckFlag(str *string) bool {
	return strings.Contains(*str, ":")
}

func CheckFlagHTTP(str *string) bool {
	matched, _ := regexp.MatchString("^http", *str)

	return strings.Contains(*str, ":") && matched
}
