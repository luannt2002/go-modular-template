// Package memory là kho key-value trong RAM, an toàn khi nhiều goroutine dùng. Dữ liệu MẤT khi tắt process.
package memory

import "sync"

type Store[V any] struct {
	mu   sync.RWMutex
	data map[string]V
}

func New[V any]() *Store[V] { return &Store[V]{data: map[string]V{}} }

func (s *Store[V]) Get(key string) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

// PutIfAbsent chỉ ghi khi key chưa có; trả false nếu đã có (không ghi đè).
func (s *Store[V]) PutIfAbsent(key string, v V) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; ok {
		return false
	}
	s.data[key] = v
	return true
}
