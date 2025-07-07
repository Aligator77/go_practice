// Package helpers contain functions for simple work
package helpers

import (
	"net/url"
)

// ValidateURL function for validation url in string
func ValidateURL(link string) (result bool, err error) {
	_, err = url.ParseRequestURI(link)
	if err != nil {
		return false, err
	}

	return true, nil
}
