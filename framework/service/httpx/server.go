// Package httpx là runtime cho loại server HTTP: timeout chuẩn, tắt êm, trả JSON và lỗi thống nhất.
package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

const shutdownTimeout = 20 * time.Second

// Run chạy server tới khi ctx bị huỷ rồi tắt êm: ngừng nhận request mới, chờ request đang chạy.
func Run(ctx context.Context, addr string, h http.Handler, log *slog.Logger) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	log.Info("http server started", "addr", addr)

	select {
	case err := <-errCh:
		return err // không mở được cổng
	case <-ctx.Done():
	}

	log.Info("http server shutting down")
	sctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(sctx); err != nil {
		return err
	}
	if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Info("http server stopped")
	return nil
}
