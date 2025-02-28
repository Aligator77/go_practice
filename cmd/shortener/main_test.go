package main

import (
	"context"
	"github.com/rs/zerolog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/Aligator77/go_practice/internal/config"
	"github.com/Aligator77/go_practice/internal/helpers"
	"github.com/Aligator77/go_practice/internal/stores"
)

const localhost = "http://localhost"

func TestURLGeneration(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.New()
	if err != nil {
		logger.Error().Err(err).Msg("failed to load config")
		os.Exit(exitCodeFailure)
	}

	db, err := config.CreateDBConn(&cfg)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create db connection")
		os.Exit(exitCodeFailure)
	}

	urlServices := stores.CreateURLService(db, logger, cfg.BaseURL, cfg.LocalStore, cfg.DisableDBStore)
	generatedURL := ""

	link := helpers.GenerateRandomURL(10)
	path := helpers.GenerateRandomURL(15)
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
		generatedURL = w.Body.String()

	})

	t.Run("GET", func(t *testing.T) {
		needFullPath, _ := url.Parse(generatedURL)
		needPath := strings.Replace(needFullPath.Path, "/", "", -1)

		r := httptest.NewRequest(http.MethodGet, needFullPath.Path, nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", needPath)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

		// вызовем хендлер как обычную функцию, без запуска самого сервера
		urlServices.GetHandler(w, r)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code, "Код ответа не совпадает с ожидаемым")

		// проверим корректность полученного заголовка ответа
		assert.Equal(t, parsedLink.String(), w.Header().Get("Location"), "Заголовок ответа не совпадает с ожидаемым")
	})

	t.Run("Wrong GET", func(t *testing.T) {
		wrongLink := urlServices.MakeFullURL("abcde")
		r := httptest.NewRequest(http.MethodGet, wrongLink, nil)
		w := httptest.NewRecorder()

		// вызовем хендлер как обычную функцию, без запуска самого сервера
		urlServices.GetHandler(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Код ответа не совпадает с ожидаемым")
	})
}
