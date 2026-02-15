package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(p []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	written, err := r.ResponseWriter.Write(p)
	r.bytes += written
	return written, err
}

func RequestLogger(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(recorder, r)

			if recorder.status == 0 {
				recorder.status = http.StatusOK
			}

			attrs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"remote_ip", remoteIP(r.RemoteAddr),
				"user_agent", r.UserAgent(),
				"request_id", GetRequestID(r.Context()),
				"bytes", recorder.bytes,
			}
			if r.URL.RawQuery != "" {
				attrs = append(attrs, "query", r.URL.RawQuery)
			}

			if recorder.status >= http.StatusInternalServerError {
				logger.Error("http request", attrs...)
				return
			}
			logger.Info("http request", attrs...)
		})
	}
}

func remoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
