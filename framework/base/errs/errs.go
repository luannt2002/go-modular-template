// Package errs là kiểu lỗi chuẩn của toàn hệ thống: mọi module trả lỗi cùng một hình dạng.
package errs

import "errors"

type Kind int

const (
	KindInternal Kind = iota
	KindInvalid
	KindNotFound
	KindUnauthorized
)

type Error struct {
	Kind      Kind
	Key       string // mã máy đọc được, ví dụ "hello.name_required"
	Msg       string // câu cho người đọc, được trả ra API
	Retriable bool   // worker dựa vào đây để quyết định có thử lại không
	Err       error  // lỗi gốc, chỉ để log, không trả ra API
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Key + ": " + e.Err.Error()
	}
	return e.Key + ": " + e.Msg
}

func (e *Error) Unwrap() error { return e.Err }

func Invalid(key, msg string) *Error  { return &Error{Kind: KindInvalid, Key: key, Msg: msg} }
func NotFound(key, msg string) *Error { return &Error{Kind: KindNotFound, Key: key, Msg: msg} }

func Unauthorized() *Error {
	return &Error{Kind: KindUnauthorized, Key: "auth.unauthorized", Msg: "unauthorized"}
}

func Internal(key string, err error) *Error {
	return &Error{Kind: KindInternal, Key: key, Msg: "internal error", Retriable: true, Err: err}
}

// From lấy *Error trong chuỗi lỗi; lỗi lạ coi là Internal.
func From(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return Internal("internal", err)
}
