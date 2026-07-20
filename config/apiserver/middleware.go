package apiserver

import (
	"log/slog"
	"net/http"
	"strings"
)

func NewLoggerMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				logger.Info("Http requrest", "path", strings.Join([]string{r.Method, r.URL.Path}, " "))
				next.ServeHTTP(w, r)
			})
	}
}
