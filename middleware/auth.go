package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"

	"github.com/luannt2002/go-modular-template/framework/base/errs"
	"github.com/luannt2002/go-modular-template/framework/service/httpx"
)

// BearerToken bảo vệ một route cụ thể: gắn khi đăng ký route, không dựa vào thứ tự đăng ký.
// shortcut: một token tĩnh cho cả hệ thống; có user thật thì thay bằng JWT (kiểm alg, exp, iss).
func BearerToken(token string, log *slog.Logger) Middleware {
	want := []byte(token)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || subtle.ConstantTimeCompare([]byte(got), want) != 1 {
				httpx.Error(w, r, log, errs.Unauthorized())
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
