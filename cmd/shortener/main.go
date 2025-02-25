// пакеты исполняемых приложений должны называться main
package main

import (
	"context"
	"flag"
	"github.com/rs/zerolog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Aligator77/go_practice/internal/config"
	"github.com/Aligator77/go_practice/internal/handlers"
	"github.com/Aligator77/go_practice/internal/helpers"
	"github.com/Aligator77/go_practice/internal/stores"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	//"github.com/go-kit/kit/log"
	//"github.com/go-kit/kit/log/level"
	"github.com/joho/godotenv"
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
	serverAddrFlag := flag.String("a", "", "input server address")
	baseUrlFlag := flag.String("b", "", "input server address")
	localStoreFile := flag.String("f", "", "input server address")
	flag.Parse()

	serverAddr := cfg.Server.Address
	if len(*serverAddrFlag) > 0 && helpers.CheckFlag(serverAddrFlag) {
		serverAddr = *serverAddrFlag
	}

	BaseUrl := cfg.BaseUrl
	if len(*baseUrlFlag) > 0 && helpers.CheckFlagHttp(baseUrlFlag) {
		BaseUrl = *baseUrlFlag
	}

	localStore := cfg.LocalStore
	if len(*localStoreFile) > 0 {
		localStore = *localStoreFile
	}

	db, err := helpers.CreateDbConn(&cfg)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create db connection")
		os.Exit(exitCodeFailure)
	}

	urlServices := stores.CreateUrlService(db, logger, BaseUrl, localStore, cfg.DisableDbStore)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.NoCache)
	r.Use(middleware.Compress(5, "text/html", "application/json"))
	// create own middleware func, to pass logger variable
	// создали свою функцию, чтобы пробросить логгер
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			// вызываем следующий обработчик
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			logger.Info().Strs("data", []string{"Time Duration", strconv.FormatInt(int64(duration), 10), "Method", r.Method, "URL.Path", r.URL.Path})
		})
	})

	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", urlServices.GetHandler)
		r.Post("/", urlServices.CreatePostHandler)
		r.Post("/api/shorten", urlServices.CreateRestHandler)
	})
	r.Get("/health", handlers.HealthCheck)
	server := &http.Server{
		Addr:    serverAddr,
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
