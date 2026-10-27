// Package event là phong bì chuẩn cho mọi event giữa các module.
package event

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Key         string          `json:"key"` // khoá phân vùng: cùng key thì giữ thứ tự (partition key khi dùng Kafka)
	Data        json.RawMessage `json:"data"`
	Attempt     int             `json:"attempt"`
	PublishedAt time.Time       `json:"published_at"`
}

func New(id, name, key string, data any, now time.Time) (Event, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return Event{}, err
	}
	return Event{ID: id, Name: name, Key: key, Data: b, PublishedAt: now}, nil
}
