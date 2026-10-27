package config

import (
	"log/slog"
	"strings"
	"testing"
)

type mapSource map[string]string

func (m mapSource) Lookup(k string) (string, bool) { v, ok := m[k]; return v, ok }

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(mapSource{"AUTH_TOKEN": "0123456789abcdef"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":8080" || cfg.Env != "local" || cfg.LogLevel != slog.LevelInfo {
		t.Fatalf("defaults: %+v", cfg)
	}
}

// Báo HẾT các khoá sai trong một lỗi, không dừng ở khoá đầu tiên.
func TestLoadReportsAllProblems(t *testing.T) {
	_, err := Load(mapSource{"LOG_LEVEL": "loud"})
	if err == nil {
		t.Fatal("want error")
	}
	for _, want := range []string{"AUTH_TOKEN: missing", "LOG_LEVEL"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q missing %q", err, want)
		}
	}
}

func TestLoadRejectsShortToken(t *testing.T) {
	if _, err := Load(mapSource{"AUTH_TOKEN": "short"}); err == nil {
		t.Fatal("want error for short token")
	}
}
