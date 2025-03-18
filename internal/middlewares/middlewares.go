// Package middlewares contain middlewares
package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

// create own middleware func, to pass logger variable
// создали свою функцию, чтобы пробросить логгер
func GzipAndLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

		if slices.Contains(r.Header.Values("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				logger.Error().Err(err).Msg("failed to create gzip reader")
			}
			defer func(gzipReader *gzip.Reader) {
				err := gzipReader.Close()
				if err != nil {
					logger.Error().Err(err).Msg("failed to close gzip reader")
				}
			}(gzipReader)

			r.Body = io.NopCloser(gzipReader)
		}

		start := time.Now()
		// вызываем следующий обработчик
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		logger.Info().Strs("data", []string{"Time Duration", strconv.FormatInt(int64(duration), 10), "Method", r.Method, "URL.Path", r.URL.Path})
	})
}
