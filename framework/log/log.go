// Package log dựng logger JSON ra stdout. Log đi tiếp đâu (ELK, Kafka…) là việc của agent thu log, không phải của app.
package log

import (
	"io"
	"log/slog"
)

func New(w io.Writer, service string, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})).With("service", service)
}
