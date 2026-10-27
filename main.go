// Điểm vào: chỉ wiring. Đọc config → dựng logger → lắp app → chạy tới khi nhận SIGINT/SIGTERM.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/luannt2002/go-modular-template/define"
	"github.com/luannt2002/go-modular-template/framework/config"
	"github.com/luannt2002/go-modular-template/framework/config/env"
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
	cfg, err := config.Load(env.Source{})
	if err != nil {
		return err // thiếu/sai config thì dừng ngay lúc khởi động
	}
	log := applog.New(os.Stdout, define.ServiceName, cfg.LogLevel).With("env", cfg.Env)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return server.New(cfg, log).Run(ctx)
}
