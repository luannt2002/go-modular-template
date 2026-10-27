// Package middleware: các lớp bọc HTTP dùng chung.
// Thứ tự toàn cục: RequestID → Recover → AccessLog. Auth gắn THEO TỪNG ROUTE (xem routes/).
package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/luannt2002/go-modular-template/framework/base/errs"
	"github.com/luannt2002/go-modular-template/framework/service/httpx"
	"github.com/luannt2002/go-modular-template/framework/utils"
)

type Middleware func(http.Handler) http.Handler

// Chain bọc h theo thứ tự truyền vào: ms[0] là lớp ngoài cùng.
func Chain(h http.Handler, ms ...Middleware) http.Handler {
	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i](h)
	}
	return h
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" || len(id) > 64 {
			id = utils.NewID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(httpx.WithRequestID(r.Context(), id)))
	})
}

// Recover biến panic trong handler thành 500 thay vì làm sập process.
func Recover(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					if v == http.ErrAbortHandler {
						panic(v)
					}
					httpx.Error(w, r, log, errs.Internal("panic", fmt.Errorf("%v", v)))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// AccessLog ghi mỗi request một dòng. Không log header/body (có thể chứa token, dữ liệu cá nhân).
func AccessLog(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			log.InfoContext(r.Context(), "http request",
				"method", r.Method, "path", r.URL.Path, "status", sw.status,
				"duration_ms", time.Since(start).Milliseconds(), "request_id", httpx.RequestID(r.Context()))
		})
	}
}
