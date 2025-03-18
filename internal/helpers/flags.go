// Package helpers contain functions for simple work
package helpers

import (
	"regexp"
	"strings"
)

// CheckFlag function for check flags contain server with port in string
//
// Example:
//
// dbDsn := flag.String("d", "", "input db dsn address")
//
// flag.Parse()
//
// if len(*serverAddrFlag) > 0 && helpers.CheckFlag(serverAddrFlag) {
func CheckFlag(str *string) bool {
	return strings.Contains(*str, ":")
}

// CheckFlagHTTP function for check http in string for flags
//
// Example:
//
// baseURLFlag := flag.String("b", "", "input server address")
//
// flag.Parse()
//
// if len(*baseURLFlag) > 0 && helpers.CheckFlagHTTP(baseURLFlag) {
func CheckFlagHTTP(str *string) bool {
	matched, _ := regexp.MatchString("^http", *str)

	return strings.Contains(*str, ":") && matched
}
