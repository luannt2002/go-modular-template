// Package dto: dữ liệu vào/ra API của module hello. Tách khỏi model để không lộ field nội bộ ra ngoài.
package dto

import "time"

type CreateRequest struct {
	Name string `json:"name"`
}

type Greeting struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
