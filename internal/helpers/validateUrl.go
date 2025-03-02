package helpers

import (
	"errors"
	"regexp"
	"strings"
)

func ValidateURL(url string) (result bool, err error) {
	matched, err := regexp.MatchString("^http", url)
	if err != nil {
		return false, err
	}
	matched2, err := regexp.MatchString("(.ru|.com|.net)", url)
	if err != nil {
		return false, err
	}
	result = strings.Contains(url, ":") && matched && matched2
	if !result {
		err = errors.New("Error Validate Url")
		return false, err
	}

	return result, err
}
