// Package pubsub là hợp đồng publish/subscribe. Bản cài nằm ở thư mục con: memory/ (thêm kafka/ khi cần).
package pubsub

import (
	"context"

	"github.com/luannt2002/go-modular-template/framework/base/event"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, ev event.Event) error
}

// Handler xử lý một event; trả lỗi để bản cài quyết định retry / dead-letter.
type Handler func(ctx context.Context, ev event.Event) error

type Subscriber interface {
	Subscribe(topic string, h Handler)
}
