// Package routes là bản đồ API. 0_init.go gắn phần chung và gọi route của từng module; mỗi module một file.
package routes

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/luannt2002/go-modular-template/framework/base/errs"
	"github.com/luannt2002/go-modular-template/framework/service/httpx"
	"github.com/luannt2002/go-modular-template/middleware"
	hellohandler "github.com/luannt2002/go-modular-template/modules/hello/handler"
)

type Deps struct {
	Log       *slog.Logger
	AuthToken string
	Hello     *hellohandler.Handler
}

func New(d Deps) http.Handler {
	r := chi.NewRouter()
	// chi bắt buộc Use() đứng TRƯỚC mọi route (sai thứ tự là panic lúc khởi động).
	r.Use(middleware.RequestID, middleware.Recover(d.Log), middleware.AccessLog(d.Log))
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		httpx.Error(w, req, d.Log, errs.NotFound("route.not_found", "route not found"))
	})

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	auth := middleware.BearerToken(d.AuthToken, d.Log)
	r.Route("/hello", func(r chi.Router) { helloRoutes(r, d.Hello, auth) })
	return r
}
