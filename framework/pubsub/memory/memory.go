// Package memory là bus pub/sub trong process, dùng cho dev/test. Chạy nhiều instance thì thay bằng pubsub/kafka.
package memory

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/luannt2002/go-modular-template/framework/base/event"
	"github.com/luannt2002/go-modular-template/framework/pubsub"
)

var ErrQueueFull = errors.New("pubsub/memory: queue full")

type delivery struct {
	topic string
	ev    event.Event
}

type Bus struct {
	log      *slog.Logger
	queue    chan delivery
	mu       sync.RWMutex
	handlers map[string][]pubsub.Handler
}

func New(log *slog.Logger, buffer int) *Bus {
	return &Bus{log: log, queue: make(chan delivery, buffer), handlers: map[string][]pubsub.Handler{}}
}

func (b *Bus) Subscribe(topic string, h pubsub.Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[topic] = append(b.handlers[topic], h)
}

// Publish không chặn request: hàng đầy thì trả lỗi cho caller biết, không âm thầm bỏ.
func (b *Bus) Publish(_ context.Context, topic string, ev event.Event) error {
	select {
	case b.queue <- delivery{topic: topic, ev: ev}:
		return nil
	default:
		return ErrQueueFull
	}
}

// Run tiêu thụ tới khi ctx bị huỷ. Lỗi hoặc panic của một event KHÔNG làm chết vòng lặp.
// shortcut: event còn trong hàng bị bỏ khi tắt; bản Kafka giữ offset nên không mất.
func (b *Bus) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case d := <-b.queue:
			b.mu.RLock()
			hs := b.handlers[d.topic]
			b.mu.RUnlock()
			for _, h := range hs {
				b.handle(ctx, h, d)
			}
		}
	}
}

func (b *Bus) handle(ctx context.Context, h pubsub.Handler, d delivery) {
	defer func() {
		if r := recover(); r != nil {
			b.log.Error("event handler panic", "topic", d.topic, "event_id", d.ev.ID, "panic", fmt.Sprint(r))
		}
	}()
	if err := h(ctx, d.ev); err != nil {
		// shortcut: không retry/dead-letter trong RAM; bản Kafka cần topic retry + DLQ.
		b.log.Error("event handler failed", "topic", d.topic, "event_id", d.ev.ID, "err", err)
	}
}
