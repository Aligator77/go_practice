package main

import (
	"github.com/Aligator77/go_practice/internal/config"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/stretchr/testify/assert"

	"github.com/Aligator77/go_practice/internal/helpers"
	"github.com/Aligator77/go_practice/internal/stores"
)

const localhost = "http://localhost"

func TestUrlGeneration(t *testing.T) {
	logger := log.NewJSONLogger(os.Stdout)
	logger = log.NewSyncLogger(logger)
	logger = level.NewFilter(logger, level.AllowDebug())
	logger = log.With(logger, "caller", log.DefaultCaller, "ts", log.DefaultTimestampUTC)

	cfg, err := config.New()
	if err != nil {
		level.Error(logger).Log("msg", "failed to load config", "err", err)
		os.Exit(exitCodeFailure)
	}

	db, err := helpers.CreateDbConn(&cfg)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create db connection", "err", err)
		os.Exit(exitCodeFailure)
	}

	urlServices := stores.CreateUrlService(db, logger, cfg.BaseUrl, cfg.DisableDbStore)
	generatedUrl := ""

	link := helpers.GenerateRandomUrl(10)
	path := helpers.GenerateRandomUrl(15)
	parsedLink, _ := url.Parse(localhost)
	parsedLink.Host = link
	parsedLink.Path = path

	t.Run("POST", func(t *testing.T) {
		body := strings.NewReader(parsedLink.String())
		r := httptest.NewRequest(http.MethodPost, "/", body)
		w := httptest.NewRecorder()

		// вызовем хендлер как обычную функцию, без запуска самого сервера
		urlServices.CreatePostHandler(w, r)

		assert.Equal(t, http.StatusCreated, w.Code, "Код ответа не совпадает с ожидаемым")
		generatedUrl = w.Body.String()

	})

	t.Run("GET", func(t *testing.T) {
		needPath, _ := url.Parse(generatedUrl)
		r := httptest.NewRequest(http.MethodGet, needPath.Path, nil)
		w := httptest.NewRecorder()

		// вызовем хендлер как обычную функцию, без запуска самого сервера
		urlServices.GetHandler(w, r)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code, "Код ответа не совпадает с ожидаемым")

		// проверим корректность полученного тела ответа
		assert.Equal(t, parsedLink.String(), w.Header().Get("Location"), "Тело ответа не совпадает с ожидаемым")
	})

	t.Run("Wrong GET", func(t *testing.T) {
		wrongLink := urlServices.MakeFullUrl("abcde")
		r := httptest.NewRequest(http.MethodGet, wrongLink, nil)
		w := httptest.NewRecorder()

		// вызовем хендлер как обычную функцию, без запуска самого сервера
		urlServices.GetHandler(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Код ответа не совпадает с ожидаемым")
	})
}
