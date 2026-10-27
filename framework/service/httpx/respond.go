package httpx

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/luannt2002/go-modular-template/framework/base/errs"
)

type ctxKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

type errorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error đổi lỗi miền sang HTTP status. Lỗi Internal chỉ được log, chi tiết không trả ra ngoài.
func Error(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	e := errs.From(err)
	status := statusOf(e.Kind)
	rid := RequestID(r.Context())
	if status >= http.StatusInternalServerError {
		log.ErrorContext(r.Context(), "request failed", "key", e.Key, "err", err, "request_id", rid)
	}
	JSON(w, status, errorBody{Code: e.Key, Message: e.Msg, RequestID: rid})
}

func statusOf(k errs.Kind) int {
	switch k {
	case errs.KindInvalid:
		return http.StatusBadRequest
	case errs.KindNotFound:
		return http.StatusNotFound
	case errs.KindUnauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
