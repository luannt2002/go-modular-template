// Package server là nơi lắp ráp: dựng hạ tầng, dựng từng module, nối route. Mỗi module chỉ nhận đúng thứ nó cần.
package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/luannt2002/go-modular-template/define"
	"github.com/luannt2002/go-modular-template/framework/config"
	"github.com/luannt2002/go-modular-template/framework/dbmem/memory"
	membus "github.com/luannt2002/go-modular-template/framework/pubsub/memory"
	"github.com/luannt2002/go-modular-template/framework/service/httpx"
	hellohandler "github.com/luannt2002/go-modular-template/modules/hello/handler"
	"github.com/luannt2002/go-modular-template/modules/hello/model"
	hellorepo "github.com/luannt2002/go-modular-template/modules/hello/repo"
	helloservice "github.com/luannt2002/go-modular-template/modules/hello/service"
	"github.com/luannt2002/go-modular-template/routes"
)

const busBuffer = 1024

type App struct {
	cfg     config.Config
	log     *slog.Logger
	bus     *membus.Bus
	handler http.Handler
}

func New(cfg config.Config, log *slog.Logger) *App {
	bus := membus.New(log, busBuffer)

	// module hello
	helloSvc := helloservice.New(hellorepo.NewMemory(memory.New[model.Greeting]()), bus)
	hello := hellohandler.New(helloSvc, log)
	bus.Subscribe(define.TopicHello, hello.OnGreetingCreated)

	h := routes.New(routes.Deps{Log: log, AuthToken: cfg.AuthToken, Hello: hello})
	return &App{cfg: cfg, log: log, bus: bus, handler: h}
}

// Handler cho test gọi thẳng, không cần mở cổng.
func (a *App) Handler() http.Handler { return a.handler }

// Run chạy consumer + HTTP server tới khi ctx bị huỷ.
func (a *App) Run(ctx context.Context) error {
	go a.bus.Run(ctx)
	return httpx.Run(ctx, a.cfg.HTTPAddr, a.handler, a.log)
}
