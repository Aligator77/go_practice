package helpers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFlags(t *testing.T) {

	checkFlagsVariant := []struct {
		s string
		b bool
	}{
		{"localhost", false},
		{"localhost:8080", true},
		{"localhost:90", true},
	}
	for _, c := range checkFlagsVariant {
		cf := CheckFlag(&c.s)
		assert.Equal(t, c.b, cf, "CheckFlag Проверка не пройдена")

	}

	checkFlagsHTTPVariant := []struct {
		s string
		b bool
	}{
		{"localhost", false},
		{"http://localhost:8080", true},
		{"http://ya.ru:8080", true},
		{"http://localhost", true},
	}
	for _, c := range checkFlagsHTTPVariant {
		cfh := CheckFlagHTTP(&c.s)
		assert.Equal(t, c.b, cfh, "CheckFlagHTTP Проверка не пройдена")

	}

}
