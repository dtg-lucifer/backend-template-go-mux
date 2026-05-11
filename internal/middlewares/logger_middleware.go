package middlewares

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/your-username/go-mux-backend-template/internal/utils"
)

// LoggerMiddleware logs every HTTP request with method, path, status code, duration, and
// client IP. It writes to both stdout and logs/events.log.
// In development, it also logs query params, path params, and request body.
func LoggerMiddleware(environment string) func(http.Handler) http.Handler {
	// Ensure the log directory and file exist before the first request arrives
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		if err := os.Mkdir("logs", os.ModePerm); err != nil {
			slog.Error("Failed to create logs directory", "error", err)
		}
	}
	if _, err := os.Stat("logs/events.log"); os.IsNotExist(err) {
		if _, err := os.Create("logs/events.log"); err != nil {
			slog.Error("Failed to create events log file", "error", err)
		}
	}

	stdLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	isDev := strings.EqualFold(environment, "development")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			var body string
			var query map[string][]string
			var params map[string]string

			if isDev {
				query = r.URL.Query()
				params = mux.Vars(r)
				if r.Body != nil {
					if bodyBytes, err := io.ReadAll(r.Body); err == nil {
						body = string(bodyBytes)
						r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
					}
				}
			}

			// Wrap the writer so we can capture the status code after the handler runs
			rw := &utils.ResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.StatusCode),
				slog.Duration("duration", duration),
				slog.String("ip", utils.GetIP(r)),
				slog.String("request_id", w.Header().Get("X-Request-ID")),
			}

			if isDev {
				attrs = append(attrs,
					slog.Any("query", query),
					slog.Any("params", params),
				)
				if body != "" {
					attrs = append(attrs, slog.String("body", body))
				}
			}

			args := make([]any, 0, len(attrs))
			for _, a := range attrs {
				args = append(args, a)
			}

			// Open the events log for this request (append mode)
			f, err := os.OpenFile("logs/events.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
			if err != nil {
				slog.Error("Failed to open events log", "error", err)
			} else {
				defer f.Close()
				fileLogger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelInfo}))
				fileLogger.Info("HTTP request", args...)
			}

			stdLogger.Info("HTTP request", args...)
		})
	}
}
