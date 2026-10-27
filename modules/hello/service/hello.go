// Package service: nghiệp vụ của module hello. Chỉ phụ thuộc interface (Repo, Publisher), không biết HTTP hay DB cụ thể.
package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/luannt2002/go-modular-template/define"
	"github.com/luannt2002/go-modular-template/framework/base/errs"
	"github.com/luannt2002/go-modular-template/framework/base/event"
	"github.com/luannt2002/go-modular-template/framework/pubsub"
	"github.com/luannt2002/go-modular-template/framework/utils"
	"github.com/luannt2002/go-modular-template/modules/hello/model"
)

type Repo interface {
	Create(ctx context.Context, g model.Greeting) error
	Get(ctx context.Context, id string) (model.Greeting, bool, error)
}

type Service struct {
	repo  Repo
	pub   pubsub.Publisher
	newID func() string
	now   func() time.Time
}

func New(repo Repo, pub pubsub.Publisher) *Service {
	return &Service{repo: repo, pub: pub, newID: utils.NewID, now: time.Now}
}

const maxNameLen = 100

func (s *Service) Create(ctx context.Context, name string) (model.Greeting, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Greeting{}, errs.Invalid("hello.name_required", "name is required")
	}
	if utf8.RuneCountInString(name) > maxNameLen {
		return model.Greeting{}, errs.Invalid("hello.name_too_long", fmt.Sprintf("name must be at most %d characters", maxNameLen))
	}

	g := model.Greeting{ID: s.newID(), Name: name, CreatedAt: s.now().UTC()}
	if err := s.repo.Create(ctx, g); err != nil {
		return model.Greeting{}, errs.Internal("hello.save", err)
	}

	ev, err := event.New(s.newID(), define.EventHelloCreated, g.ID, g, s.now().UTC())
	if err != nil {
		return model.Greeting{}, errs.Internal("hello.event", err)
	}
	// shortcut: ghi kho rồi mới publish nên hai bước có thể lệch nhau khi publish lỗi;
	// hệ thật dùng Transactional Outbox (ghi event cùng transaction với dữ liệu).
	if err := s.pub.Publish(ctx, define.TopicHello, ev); err != nil {
		return model.Greeting{}, errs.Internal("hello.publish", err)
	}
	return g, nil
}

func (s *Service) Get(ctx context.Context, id string) (model.Greeting, error) {
	g, found, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Greeting{}, errs.Internal("hello.load", err)
	}
	if !found {
		return model.Greeting{}, errs.NotFound("hello.not_found", "greeting not found")
	}
	return g, nil
}
