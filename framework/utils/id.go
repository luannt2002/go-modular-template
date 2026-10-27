// Package utils: tiện ích nhỏ không thuộc năng lực hạ tầng nào.
package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID trả 32 ký tự hex ngẫu nhiên (crypto/rand).
func NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("utils: crypto/rand failed: " + err.Error()) // không có nguồn ngẫu nhiên thì không nên chạy tiếp
	}
	return hex.EncodeToString(b[:])
}
