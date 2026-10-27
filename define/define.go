// Package define giữ hằng số dùng chung toàn app: tên service, tên topic, tên event.
package define

const ServiceName = "go-modular-template"

// Topic gom event theo domain; tên event: <domain>.<việc đã xảy ra>.
const (
	TopicHello        = "hello"
	EventHelloCreated = "hello.created"
)
