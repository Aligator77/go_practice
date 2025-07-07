// Package main пакеты исполняемых приложений должны называться main
package main

import (
	"compress/gzip"
	"context"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	"github.com/Aligator77/go_practice/internal/config"
	"github.com/Aligator77/go_practice/internal/controllers"
	"github.com/Aligator77/go_practice/internal/handlers"
	"github.com/Aligator77/go_practice/internal/middlewares"
	"github.com/Aligator77/go_practice/internal/stores"
)

const (
	exitCodeSuccess = 0
	exitCodeFailure = 1
)

// Example:
// POST / HTTP/1.1
// Host: localhost:8080
// Content-Type: text/plain
// https://practicum.yandex.ru/

// HTTP/1.1 201 Created
// Content-Type: text/plain
// Content-Length: 30
// http://localhost:8080/EwHXdJfB

// GET /EwHXdJfB HTTP/1.1
// Host: localhost:8080
// Content-Type: text/plain

// HTTP/1.1 307 Temporary Redirect
// Location: https://practicum.yandex.ru/

// TODO add POST / path and return 201
// input and output text/plain
// TODO add GET /{id} if exist url return 307 and info, else 400 err
// функция main вызывается автоматически при запуске приложения
func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	errc := make(chan error, 1)
	doneCh := make(chan struct{})
	sigc := make(chan os.Signal, 1)

	if err := godotenv.Load(); err != nil {
		logger.Warn().Msg("error loading .env file")
	}

	cfg, err := config.New()
	if err != nil {
		logger.Error().Err(err).Msg("failed to load config")
		os.Exit(exitCodeFailure)
	}

	db, err := config.NewDBConn(&cfg)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create db connection")
	}
	dbController := controllers.NewDBController(ctx, db)
	_, err = dbController.Migrate(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to migrate db")
	}
	urlServices := stores.NewURLService(db, logger, cfg.BaseURL, cfg.LocalStore, cfg.DisableDBStore)
	urlController := controllers.NewURLController(urlServices)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.NoCache)
	r.Use(middleware.Compress(gzip.DefaultCompression, "text/html", "application/json"))
	r.Use(middlewares.GzipAndLogger)

	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", urlController.GetHandler)
		r.Post("/", urlController.CreatePostHandler)
		r.Post("/api/shorten", urlController.CreateRestHandler)
		r.Post("/api/shorten/batch", urlController.CreateBatchHandler)
		r.Get("/api/user/urls", urlController.CreateFullRestHandler)
		r.Delete("/api/user/urls", urlController.CreateFullRestHandler)
		r.Post("/api/user/urls", urlController.CreateFullRestHandler)
		r.Get("/ping", dbController.CheckConnectHandler)

		// Регистрация pprof-обработчиков
		r.HandleFunc("/debug/pprof/", pprof.Index)
		r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		r.HandleFunc("/debug/pprof/profile", pprof.Profile)
		r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		r.HandleFunc("/debug/pprof/trace", pprof.Trace)
	})
	r.Get("/health", handlers.HealthCheck)
	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: r,
	}
	go func() {
		err = server.ListenAndServe()
	}()
	logger.Info().Msg("go service Started")

	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	defer func() {
		signal.Stop(sigc)
		cancel()
	}()

	go func() {
		select {
		case sig := <-sigc:
			logger.Info().Str("signal", sig.String()).Msg("received signal, exiting")

			err := server.Shutdown(ctx)
			if err != nil {
				return
			} // Close http connection
			_ = urlServices.Shutdown()
			signal.Stop(sigc)
			close(doneCh)
		case <-errc:
			logger.Info().Str("error code", strconv.Itoa(exitCodeFailure)).Msg("now exiting with error")
			os.Exit(exitCodeFailure)
		}
	}()

	<-doneCh
	logger.Info().Msg("goodbye")
}
