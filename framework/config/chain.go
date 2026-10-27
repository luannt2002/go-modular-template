package config

// Chain hỏi lần lượt từng Source; nguồn đứng trước thắng. Giá trị rỗng coi như không có, hỏi tiếp nguồn sau.
type Chain []Source

func (c Chain) Lookup(key string) (string, bool) {
	for _, s := range c {
		if v, ok := s.Lookup(key); ok && v != "" {
			return v, true
		}
	}
	return "", false
}
