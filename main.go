// Điểm vào: chỉ wiring. Đọc config → dựng logger → lắp app → chạy tới khi nhận SIGINT/SIGTERM.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luannt2002/go-modular-template/define"
	"github.com/luannt2002/go-modular-template/framework/config"
	"github.com/luannt2002/go-modular-template/framework/config/env"
	"github.com/luannt2002/go-modular-template/framework/config/vault"
	applog "github.com/luannt2002/go-modular-template/framework/log"
	"github.com/luannt2002/go-modular-template/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	src, err := configSource()
	if err != nil {
		return err
	}
	cfg, err := config.Load(src)
	if err != nil {
		return err // thiếu/sai config thì dừng ngay lúc khởi động
	}
	log := applog.New(os.Stdout, define.ServiceName, cfg.LogLevel).With("env", cfg.Env)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return server.New(cfg, log).Run(ctx)
}

// configSource: có VAULT_ADDR thì đọc Vault trước, thiếu khoá nào mới hỏi biến môi trường; không có thì chỉ dùng env.
func configSource() (config.Source, error) {
	if os.Getenv("VAULT_ADDR") == "" {
		return env.Source{}, nil
	}
	opts, err := vault.OptionsFrom(env.Source{})
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	vs, err := vault.New(ctx, opts)
	if err != nil {
		return nil, err
	}
	return config.Chain{vs, env.Source{}}, nil
}
