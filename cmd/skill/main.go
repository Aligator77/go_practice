// пакеты исполняемых приложений должны называться main
package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/joho/godotenv"

	"github.com/Aligator77/go_practice/internal/config"
	"github.com/Aligator77/go_practice/internal/handlers"
	_ "github.com/Aligator77/go_practice/internal/handlers"
	"github.com/Aligator77/go_practice/internal/helpers"
	"github.com/Aligator77/go_practice/internal/stores"
)

const (
	exitCodeSuccess = 0
	exitCodeFailure = 1
	httpPort        = "8080"
)

/*func TimerTrace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// перед началом выполнения функции сохраняем текущее время
		start := time.Now()
		// вызываем следующий обработчик
		next.ServeHTTP(w, r)
		// после завершения замеряем время выполнения запроса
		duration := time.Since(start)
		// сохраняем или сразу обрабатываем полученный результат
		level.Warn(logger).Log("Time Duration", duration, "Method", r.Method, "URL.Path", r.URL.Path)
	})
}*/

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
	doneCh := make(chan struct{})
	sigc := make(chan os.Signal, 1)

	if err := godotenv.Load(); err != nil {
		level.Warn(logger).Log("msg", "error loading .env file")
	}

	cfg, err := config.New()
	if err != nil {
		level.Error(logger).Log("msg", "failed to load config", "err", err)
		os.Exit(exitCodeFailure)
	}
	serverAddrFlag := flag.String("a", "", "input server address")
	siteHostFlag := flag.String("b", "", "input server address")
	flag.Parse()

	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	if len(*serverAddrFlag) > 0 && helpers.CheckFlag(serverAddrFlag) {
		serverAddr = *serverAddrFlag
	}

	siteHost := cfg.SiteHost
	if len(*siteHostFlag) > 0 && helpers.CheckFlagHttp(siteHostFlag) {
		siteHost = *siteHostFlag
	}

	db, err := helpers.CreateDbConn(&cfg)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create db connection", "err", err)
		os.Exit(exitCodeFailure)
	}

	urlServices := stores.CreateUrlService(db, logger, siteHost)
	server := &http.Server{
		Addr:                         serverAddr,
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
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.NoCache)
	//r.Use(TimerTrace)

	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", handlers.GetHandler)
		r.Post("/", urlServices.CreatePostHandler)
	})
	r.Get("/health", handlers.HealthCheck)

	err = http.ListenAndServe(":8080", r)

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
			close(doneCh)
		case <-errc:
			level.Info(logger).Log("msg", "now exiting with error", "error code", exitCodeFailure)
			os.Exit(exitCodeFailure)
		}
	}()

	<-doneCh
	level.Info(logger).Log("msg", "goodbye")
}
