// Package config đọc cấu hình từ một Source (env, Vault…), validate một lần và báo hết khoá thiếu/sai cùng lúc.
package config

import (
	"fmt"
	"log/slog"
	"strings"
)

// Source là nơi cấu hình nằm. Bản cài: config/env (thêm config/vault khi cần).
type Source interface {
	Lookup(key string) (string, bool)
}

type Config struct {
	Env       string
	HTTPAddr  string
	LogLevel  slog.Level
	AuthToken string // secret: không bao giờ log
}

const minTokenLen = 16

func Load(src Source) (Config, error) {
	var problems []string
	get := func(key, def string, required bool) string {
		v, ok := src.Lookup(key)
		v = strings.TrimSpace(v)
		if !ok || v == "" {
			if required {
				problems = append(problems, key+": missing")
			}
			return def
		}
		return v
	}

	cfg := Config{
		Env:       get("APP_ENV", "local", false),
		HTTPAddr:  get("HTTP_ADDR", ":8080", false),
		AuthToken: get("AUTH_TOKEN", "", true),
	}
	if err := cfg.LogLevel.UnmarshalText([]byte(get("LOG_LEVEL", "info", false))); err != nil {
		problems = append(problems, "LOG_LEVEL: "+err.Error())
	}
	if cfg.AuthToken != "" && len(cfg.AuthToken) < minTokenLen {
		problems = append(problems, fmt.Sprintf("AUTH_TOKEN: must be at least %d characters", minTokenLen))
	}
	if len(problems) > 0 {
		return Config{}, fmt.Errorf("invalid config: %s", strings.Join(problems, "; "))
	}
	return cfg, nil
}
