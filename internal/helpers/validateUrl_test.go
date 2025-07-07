package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLGeneration(t *testing.T) {

	// описываем набор данных: метод запроса, ожидаемый код ответа, ожидаемое тело
	testCases := []struct {
		url           string
		testName      string
		expectedValue bool
	}{
		{url: "http://localhost.ru", expectedValue: true, testName: "Good http 1"},
		{url: "http://localhost.com", expectedValue: true, testName: "Good http 2"},
		{url: "https://abc.com", expectedValue: true, testName: "Good https 3"},
		{url: "https://abc.ru", expectedValue: true, testName: "Good https 4"},
		{url: "https://abc.by", expectedValue: true, testName: "Good https 5"},
		{url: "http://abc", expectedValue: true, testName: "Good http 6"},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			res, err := ValidateURL(tc.url)
			if err != nil {
				assert.Equal(t, tc.expectedValue, res, "Ошибка в работе регулярки")
			}

			assert.Equal(t, tc.expectedValue, res, "Ошибка в тесте")
		})
	}
}
