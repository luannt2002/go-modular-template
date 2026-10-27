// Package model: dữ liệu module hello lưu trong kho. Không trả thẳng ra API (API dùng dto).
package model

import "time"

type Greeting struct {
	ID        string
	Name      string
	CreatedAt time.Time
}
