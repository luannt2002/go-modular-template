// Package handler: mọi cửa vào của module hello (HTTP, event). Đọc input, gọi service, trả output. Không chứa nghiệp vụ.
package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/luannt2002/go-modular-template/framework/base/errs"
	"github.com/luannt2002/go-modular-template/framework/base/event"
	"github.com/luannt2002/go-modular-template/framework/service/httpx"
	"github.com/luannt2002/go-modular-template/modules/hello/dto"
	"github.com/luannt2002/go-modular-template/modules/hello/model"
	"github.com/luannt2002/go-modular-template/modules/hello/service"
)

const maxBodyBytes = 1 << 20

type Handler struct {
	svc *service.Service
	log *slog.Logger
}

func New(svc *service.Service, log *slog.Logger) *Handler { return &Handler{svc: svc, log: log} }

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		httpx.Error(w, r, h.log, errs.Invalid("request.bad_json", "invalid JSON body"))
		return
	}
	g, err := h.svc.Create(r.Context(), req.Name)
	if err != nil {
		httpx.Error(w, r, h.log, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, toDTO(g))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	g, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httpx.Error(w, r, h.log, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toDTO(g))
}

// OnGreetingCreated là cửa vào từ event "hello.created".
func (h *Handler) OnGreetingCreated(ctx context.Context, ev event.Event) error {
	h.log.InfoContext(ctx, "greeting created", "event_id", ev.ID, "greeting_id", ev.Key)
	return nil
}

func toDTO(g model.Greeting) dto.Greeting {
	return dto.Greeting{ID: g.ID, Name: g.Name, Message: "Hello, " + g.Name + "!", CreatedAt: g.CreatedAt}
}
