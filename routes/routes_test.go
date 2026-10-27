package routes_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luannt2002/go-modular-template/framework/config"
	"github.com/luannt2002/go-modular-template/server"
)

const token = "test-token-0123456789"

func newApp() http.Handler {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	return server.New(config.Config{AuthToken: token}, log).Handler()
}

func do(h http.Handler, method, path, body, auth string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// Mọi route GHI phải từ chối khi không có / sai token.
func TestWriteRoutesRequireToken(t *testing.T) {
	h := newApp()
	for _, auth := range []string{"", "Bearer wrong-token", token} { // token thiếu chữ "Bearer "
		rec := do(h, http.MethodPost, "/hello", `{"name":"An"}`, auth)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("auth=%q: got %d, want 401", auth, rec.Code)
		}
	}
}

func TestHelloFlow(t *testing.T) {
	h := newApp()

	if rec := do(h, http.MethodGet, "/healthz", "", ""); rec.Code != http.StatusOK {
		t.Fatalf("healthz: got %d", rec.Code)
	}

	rec := do(h, http.MethodPost, "/hello", `{"name":"An"}`, "Bearer "+token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: got %d body=%s", rec.Code, rec.Body)
	}
	var created struct{ ID, Message string }
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.ID == "" {
		t.Fatalf("create body: %s (err %v)", rec.Body, err)
	}
	if created.Message != "Hello, An!" {
		t.Fatalf("message: %q", created.Message)
	}

	if rec := do(h, http.MethodGet, "/hello/"+created.ID, "", ""); rec.Code != http.StatusOK {
		t.Fatalf("get: got %d", rec.Code)
	}
	if rec := do(h, http.MethodGet, "/hello/khong-co", "", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("get missing: got %d", rec.Code)
	}
	// route không tồn tại vẫn trả lỗi JSON thống nhất
	rec = do(h, http.MethodGet, "/khong-ton-tai", "", "")
	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), `"code":"route.not_found"`) {
		t.Fatalf("unknown route: %d %s", rec.Code, rec.Body)
	}
}

func TestBadInput(t *testing.T) {
	h := newApp()
	cases := map[string]string{
		"json hỏng":   `{"name":`,
		"field lạ":    `{"name":"An","admin":true}`,
		"tên rỗng":    `{"name":"   "}`,
		"tên quá dài": `{"name":"` + strings.Repeat("a", 101) + `"}`,
	}
	for name, body := range cases {
		rec := do(h, http.MethodPost, "/hello", body, "Bearer "+token)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: got %d, want 400 (body=%s)", name, rec.Code, rec.Body)
		}
	}
}
