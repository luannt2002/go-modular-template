// Package vault là Source đọc cấu hình từ HashiCorp Vault (KV v2) qua HTTP API, chỉ dùng thư viện chuẩn.
//
// Đọc MỘT lần lúc khởi động: mỗi path là một folder (1 lần GET), path sau đè path trước
// (ví dụ "shared" rồi "go-modular-template"). Token không được giữ lại sau khi đọc xong.
package vault

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/luannt2002/go-modular-template/framework/config"
)

var (
	ErrNotFound  = errors.New("vault: path not found")
	ErrForbidden = errors.New("vault: permission denied")
)

type Options struct {
	Addr     string   // https://vault.example.com:8200
	Mount    string   // mount của KV v2, mặc định "secret"
	Paths    []string // đọc theo thứ tự, path sau đè path trước
	Token    string   // dùng token có sẵn, HOẶC
	RoleID   string   // AppRole (cách đăng nhập dành cho service)
	SecretID string
	CACert   string // file PEM của CA nội bộ; rỗng = dùng CA của hệ thống
	Timeout  time.Duration
}

// OptionsFrom đọc thông số kết nối Vault (VAULT_*) từ một Source khác, thường là biến môi trường.
func OptionsFrom(src config.Source) (Options, error) {
	get := func(k string) string { v, _ := src.Lookup(k); return strings.TrimSpace(v) }
	o := Options{
		Addr:     get("VAULT_ADDR"),
		Mount:    get("VAULT_KV_MOUNT"),
		Token:    get("VAULT_TOKEN"),
		RoleID:   get("VAULT_ROLE_ID"),
		SecretID: get("VAULT_SECRET_ID"),
		CACert:   get("VAULT_CACERT"),
		Timeout:  10 * time.Second,
	}
	if o.Mount == "" {
		o.Mount = "secret"
	}
	for _, p := range strings.Split(get("VAULT_PATHS"), ",") {
		if p = strings.Trim(strings.TrimSpace(p), "/"); p != "" {
			o.Paths = append(o.Paths, p)
		}
	}

	var problems []string
	if o.Addr == "" {
		problems = append(problems, "VAULT_ADDR: missing")
	}
	if len(o.Paths) == 0 {
		problems = append(problems, "VAULT_PATHS: missing")
	}
	if o.Token == "" && (o.RoleID == "" || o.SecretID == "") {
		problems = append(problems, "VAULT_TOKEN or VAULT_ROLE_ID+VAULT_SECRET_ID: missing")
	}
	if len(problems) > 0 {
		return Options{}, fmt.Errorf("invalid vault config: %s", strings.Join(problems, "; "))
	}
	return o, nil
}

// Source là ảnh chụp cấu hình đọc từ Vault lúc khởi động.
type Source struct {
	values map[string]string
}

func (s *Source) Lookup(key string) (string, bool) {
	v, ok := s.values[key]
	return v, ok
}

// New đăng nhập (nếu dùng AppRole) rồi đọc lần lượt các path.
// shortcut: chỉ đọc lúc khởi động, không gia hạn token, không theo dõi thay đổi; đổi secret thì restart.
func New(ctx context.Context, o Options) (*Source, error) {
	c, err := newClient(o)
	if err != nil {
		return nil, err
	}
	token := o.Token
	if token == "" {
		if token, err = c.loginAppRole(ctx, o.RoleID, o.SecretID); err != nil {
			return nil, err
		}
	}
	values := map[string]string{}
	for _, p := range o.Paths {
		data, err := c.readKV2(ctx, token, o.Mount, p)
		if err != nil {
			return nil, err
		}
		for k, v := range data {
			values[k] = v
		}
	}
	return &Source{values: values}, nil
}

type client struct {
	base string
	http *http.Client
}

func newClient(o Options) (*client, error) {
	u, err := url.Parse(o.Addr)
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("vault: invalid VAULT_ADDR %q", o.Addr)
	}
	// http thuần chỉ cho Vault dev chạy trên máy mình; ra mạng thì bắt buộc https.
	if u.Scheme != "https" && !(u.Scheme == "http" && isLoopback(u.Hostname())) {
		return nil, fmt.Errorf("vault: %s must use https (http is allowed only for localhost)", o.Addr)
	}

	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12} // KHÔNG bao giờ InsecureSkipVerify
	if o.CACert != "" {
		pem, err := os.ReadFile(o.CACert)
		if err != nil {
			return nil, fmt.Errorf("vault: read CA cert: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("vault: no certificate found in %s", o.CACert)
		}
		tlsCfg.RootCAs = pool
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = tlsCfg
	return &client{base: strings.TrimRight(o.Addr, "/"), http: &http.Client{Transport: tr, Timeout: o.Timeout}}, nil
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (c *client) loginAppRole(ctx context.Context, roleID, secretID string) (string, error) {
	body, _ := json.Marshal(map[string]string{"role_id": roleID, "secret_id": secretID})
	var out struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
	}
	if err := c.do(ctx, http.MethodPost, "/v1/auth/approle/login", "", body, &out); err != nil {
		return "", fmt.Errorf("vault approle login: %w", err)
	}
	if out.Auth.ClientToken == "" {
		return "", errors.New("vault approle login: empty token in response")
	}
	return out.Auth.ClientToken, nil
}

func (c *client) readKV2(ctx context.Context, token, mount, path string) (map[string]string, error) {
	var out struct {
		Data struct {
			Data map[string]json.RawMessage `json:"data"`
		} `json:"data"`
	}
	if err := c.do(ctx, http.MethodGet, "/v1/"+mount+"/data/"+path, token, nil, &out); err != nil {
		return nil, fmt.Errorf("vault read %s/%s: %w", mount, path, err)
	}
	values := make(map[string]string, len(out.Data.Data))
	for k, raw := range out.Data.Data {
		var s string
		if json.Unmarshal(raw, &s) == nil {
			values[k] = s
		} else {
			values[k] = string(raw) // số/bool/object giữ nguyên dạng JSON: 8080, true, {...}
		}
	}
	return values, nil
}

const maxResponseBytes = 1 << 20

func (c *client) do(ctx context.Context, method, path, token string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-Vault-Token", token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err // lỗi mạng / TLS: KHÁC với "không tìm thấy"
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case resp.StatusCode == http.StatusForbidden:
		return ErrForbidden
	case resp.StatusCode/100 != 2:
		return fmt.Errorf("unexpected status %d", resp.StatusCode) // không chép body: có thể chứa dữ liệu nhạy cảm
	}
	return json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(out)
}
