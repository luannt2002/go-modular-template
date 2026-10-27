package memory

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/luannt2002/go-modular-template/framework/base/event"
)

// Một event làm handler panic hoặc lỗi KHÔNG được làm chết vòng tiêu thụ.
func TestBadEventDoesNotKillConsumer(t *testing.T) {
	bus := New(slog.New(slog.NewJSONHandler(io.Discard, nil)), 8)
	got := make(chan string, 1)
	bus.Subscribe("t", func(_ context.Context, ev event.Event) error {
		switch ev.ID {
		case "panic":
			panic("boom")
		case "fail":
			return errors.New("fail")
		}
		got <- ev.ID
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go bus.Run(ctx)

	for _, id := range []string{"panic", "fail", "ok"} {
		if err := bus.Publish(ctx, "t", event.Event{ID: id}); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case id := <-got:
		if id != "ok" {
			t.Fatalf("got %q", id)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("consumer died: event 'ok' never processed")
	}
}

func TestPublishFullQueueReturnsError(t *testing.T) {
	bus := New(slog.New(slog.NewJSONHandler(io.Discard, nil)), 1)
	_ = bus.Publish(context.Background(), "t", event.Event{ID: "1"})
	if err := bus.Publish(context.Background(), "t", event.Event{ID: "2"}); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("want ErrQueueFull, got %v", err)
	}
}
