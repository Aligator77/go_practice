package helpers

import (
	"net/url"
)

func ValidateURL(link string) (result bool, err error) {
	_, err = url.ParseRequestURI(link)
	if err != nil {
		return false, err
	}

	return true, nil
}
