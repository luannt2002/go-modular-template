package vault

import (
	"context"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeVault giả lập đủ phần API mà Source dùng: AppRole login + đọc KV v2.
func fakeVault(t *testing.T, secrets map[string]map[string]any) http.Handler {
	t.Helper()
	const token = "tok-123"
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/auth/approle/login", func(w http.ResponseWriter, r *http.Request) {
		var in map[string]string
		_ = json.NewDecoder(r.Body).Decode(&in)
		if in["role_id"] != "role" || in["secret_id"] != "secret" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"auth": map[string]string{"client_token": token}})
	})
	mux.HandleFunc("GET /v1/secret/data/{path...}", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != token {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		data, ok := secrets[r.PathValue("path")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": data}})
	})
	return mux
}

func opts(addr string) Options {
	return Options{Addr: addr, Mount: "secret", Token: "tok-123", Timeout: 2 * time.Second}
}

func TestLayeredPathsLaterWins(t *testing.T) {
	srv := httptest.NewServer(fakeVault(t, map[string]map[string]any{
		"shared": {"LOG_LEVEL": "warn", "AUTH_TOKEN": "shared-token-0000000"},
		"app":    {"AUTH_TOKEN": "app-token-000000000", "PORT": 8080, "DEBUG": true},
	}))
	defer srv.Close()

	o := opts(srv.URL)
	o.Paths = []string{"shared", "app"}
	s, err := New(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]string{
		"LOG_LEVEL": "warn", "AUTH_TOKEN": "app-token-000000000", "PORT": "8080", "DEBUG": "true",
	} {
		if got, _ := s.Lookup(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
}

func TestAppRoleLogin(t *testing.T) {
	srv := httptest.NewServer(fakeVault(t, map[string]map[string]any{"app": {"K": "v"}}))
	defer srv.Close()

	o := opts(srv.URL)
	o.Token, o.RoleID, o.SecretID, o.Paths = "", "role", "secret", []string{"app"}
	s, err := New(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := s.Lookup("K"); v != "v" {
		t.Fatalf("K = %q", v)
	}

	o.SecretID = "wrong"
	if _, err := New(context.Background(), o); err == nil {
		t.Fatal("want login error with wrong secret_id")
	}
}

// Thiếu path và sai quyền phải ra lỗi KHÁC NHAU (bản cũ gộp hết thành chuỗi rỗng).
func TestNotFoundAndForbiddenAreDistinct(t *testing.T) {
	srv := httptest.NewServer(fakeVault(t, map[string]map[string]any{}))
	defer srv.Close()

	o := opts(srv.URL)
	o.Paths = []string{"missing"}
	if _, err := New(context.Background(), o); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	o.Token = "bad-token"
	if _, err := New(context.Background(), o); !errors.Is(err, ErrForbidden) {
		t.Fatalf("want ErrForbidden, got %v", err)
	}
}

// TLS phải được kiểm thật: không có CA đúng thì từ chối, có CA thì qua.
func TestTLSIsVerified(t *testing.T) {
	srv := httptest.NewTLSServer(fakeVault(t, map[string]map[string]any{"app": {"K": "v"}}))
	defer srv.Close()

	o := opts(srv.URL)
	o.Paths = []string{"app"}
	if _, err := New(context.Background(), o); err == nil || !strings.Contains(err.Error(), "certificate") {
		t.Fatalf("want certificate error without CA, got %v", err)
	}

	caFile := filepath.Join(t.TempDir(), "ca.pem")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srv.Certificate().Raw})
	if err := os.WriteFile(caFile, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	o.CACert = caFile
	if _, err := New(context.Background(), o); err != nil {
		t.Fatalf("with CA: %v", err)
	}
}

func TestPlainHTTPOnlyForLoopback(t *testing.T) {
	o := opts("http://vault.example.com:8200")
	o.Paths = []string{"app"}
	if _, err := New(context.Background(), o); err == nil || !strings.Contains(err.Error(), "https") {
		t.Fatalf("want https error, got %v", err)
	}
}

type mapSource map[string]string

func (m mapSource) Lookup(k string) (string, bool) { v, ok := m[k]; return v, ok }

func TestOptionsFromReportsAllProblems(t *testing.T) {
	_, err := OptionsFrom(mapSource{})
	if err == nil {
		t.Fatal("want error")
	}
	for _, want := range []string{"VAULT_ADDR", "VAULT_PATHS", "VAULT_TOKEN"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q missing %q", err, want)
		}
	}

	o, err := OptionsFrom(mapSource{"VAULT_ADDR": "https://v:8200", "VAULT_PATHS": " shared , /app/ ", "VAULT_TOKEN": "t"})
	if err != nil {
		t.Fatal(err)
	}
	if o.Mount != "secret" || len(o.Paths) != 2 || o.Paths[1] != "app" {
		t.Fatalf("options = %+v", o)
	}
}
