package helpers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShorter(t *testing.T) {

	urlLength := []int{
		1,
		10,
		100,
	}
	for _, u := range urlLength {
		ul := GenerateRandomURL(u)
		assert.Equal(t, u, len(ul), "GenerateRandomURL Длина не совпала")
	}

}
