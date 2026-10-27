package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/luannt2002/go-modular-template/middleware"
	hellohandler "github.com/luannt2002/go-modular-template/modules/hello/handler"
)

// helloRoutes gắn dưới /hello. Auth gắn theo TỪNG route bằng r.With(auth).
func helloRoutes(r chi.Router, h *hellohandler.Handler, auth middleware.Middleware) {
	r.Get("/{id}", h.Get)            // public: chỉ đọc
	r.With(auth).Post("/", h.Create) // ghi: bắt buộc token
}
