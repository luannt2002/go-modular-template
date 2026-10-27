// Package env là Source đọc cấu hình từ biến môi trường.
package env

import "os"

type Source struct{}

func (Source) Lookup(key string) (string, bool) { return os.LookupEnv(key) }
