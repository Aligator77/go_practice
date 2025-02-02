// пакеты исполняемых приложений должны называться main
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/joho/godotenv"

	"github.com/Aligator77/go_practice/internal/config"
	"github.com/Aligator77/go_practice/internal/handlers"
	"github.com/Aligator77/go_practice/internal/stores"
)

const (
	exitCodeSuccess = 0
	exitCodeFailure = 1
	httpPort        = "8080"
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
	logger := log.NewJSONLogger(os.Stdout)
	logger = log.NewSyncLogger(logger)
	logger = level.NewFilter(logger, level.AllowDebug())
	logger = log.With(logger, "caller", log.DefaultCaller, "ts", log.DefaultTimestampUTC)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	errc := make(chan error, 1)
	donec := make(chan struct{})
	sigc := make(chan os.Signal, 1)

	if err := godotenv.Load(); err != nil {
		level.Warn(logger).Log("msg", "error loading .env file")
	}

	cfg, err := config.New()
	if err != nil {
		level.Error(logger).Log("msg", "failed to load config", "err", err)
		os.Exit(exitCodeFailure)
	}

	urlServices := stores.CreateUrlService()
	server := &http.Server{
		Addr:                         cfg.Server.Host + ":" + cfg.Server.Port,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  0,
		ReadHeaderTimeout:            0,
		WriteTimeout:                 0,
		IdleTimeout:                  0,
		MaxHeaderBytes:               0,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext:                  nil,
		ConnContext:                  nil,
	}

	http.HandleFunc("/", handlers.AllInOne)
	http.HandleFunc("/health", handlers.HealthCheck)
	err = server.ListenAndServe()

	level.Info(logger).Log("msg", "searchService Started")

	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	defer func() {
		signal.Stop(sigc)
		cancel()
	}()

	go func() {
		select {
		case sig := <-sigc:
			level.Info(logger).Log("msg", "received signal, exiting", "signal", sig)
			err := server.Shutdown(ctx)
			if err != nil {
				return
			} // Close http connection
			_ = urlServices.Shutdown()
			signal.Stop(sigc)
			close(donec)
		case <-errc:
			level.Info(logger).Log("msg", "now exiting with error", "error code", exitCodeFailure)
			os.Exit(exitCodeFailure)
		}
	}()

	<-donec
	level.Info(logger).Log("msg", "goodbye")
}

// функция webhook — обработчик HTTP-запроса
func webhook(w http.ResponseWriter, r *http.Request) {

}
