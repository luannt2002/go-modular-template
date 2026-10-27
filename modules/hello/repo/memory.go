// Package repo: truy cập dữ liệu của module hello. Đây là chỗ DUY NHẤT trong module biết dữ liệu nằm ở đâu.
package repo

import (
	"context"
	"errors"

	"github.com/luannt2002/go-modular-template/framework/dbmem/memory"
	"github.com/luannt2002/go-modular-template/modules/hello/model"
)

var ErrDuplicate = errors.New("hello/repo: duplicate id")

// Memory lưu trong RAM. Thêm postgres.go cạnh file này khi cần lưu bền; service không phải đổi.
type Memory struct {
	store *memory.Store[model.Greeting]
}

func NewMemory(s *memory.Store[model.Greeting]) *Memory { return &Memory{store: s} }

func (m *Memory) Create(_ context.Context, g model.Greeting) error {
	if !m.store.PutIfAbsent(g.ID, g) {
		return ErrDuplicate
	}
	return nil
}

// Get trả found=false khi không có; err chỉ dành cho lỗi của kho (mạng, DB…).
func (m *Memory) Get(_ context.Context, id string) (model.Greeting, bool, error) {
	g, ok := m.store.Get(id)
	return g, ok, nil
}
