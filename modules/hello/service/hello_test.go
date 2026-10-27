package service

import (
	"context"
	"errors"
	"testing"

	"github.com/luannt2002/go-modular-template/define"
	"github.com/luannt2002/go-modular-template/framework/base/errs"
	"github.com/luannt2002/go-modular-template/framework/base/event"
	"github.com/luannt2002/go-modular-template/framework/dbmem/memory"
	"github.com/luannt2002/go-modular-template/modules/hello/model"
	"github.com/luannt2002/go-modular-template/modules/hello/repo"
)

type fakePub struct {
	topics []string
	events []event.Event
	err    error
}

func (f *fakePub) Publish(_ context.Context, topic string, ev event.Event) error {
	if f.err != nil {
		return f.err
	}
	f.topics = append(f.topics, topic)
	f.events = append(f.events, ev)
	return nil
}

func newSvc(pub *fakePub) *Service {
	return New(repo.NewMemory(memory.New[model.Greeting]()), pub)
}

func TestCreatePublishesEventKeyedByID(t *testing.T) {
	pub := &fakePub{}
	g, err := newSvc(pub).Create(context.Background(), "  An  ")
	if err != nil {
		t.Fatal(err)
	}
	if g.Name != "An" {
		t.Fatalf("name not trimmed: %q", g.Name)
	}
	if len(pub.events) != 1 || pub.topics[0] != define.TopicHello {
		t.Fatalf("want 1 event on %q, got %v", define.TopicHello, pub.topics)
	}
	if ev := pub.events[0]; ev.Name != define.EventHelloCreated || ev.Key != g.ID {
		t.Fatalf("event = %+v", ev)
	}
}

func TestCreateRejectsInvalidName(t *testing.T) {
	for _, name := range []string{"", "   "} {
		_, err := newSvc(&fakePub{}).Create(context.Background(), name)
		if errs.From(err).Kind != errs.KindInvalid {
			t.Fatalf("name %q: want Invalid, got %v", name, err)
		}
	}
}

func TestGetMissingIsNotFound(t *testing.T) {
	_, err := newSvc(&fakePub{}).Get(context.Background(), "x")
	if errs.From(err).Kind != errs.KindNotFound {
		t.Fatalf("want NotFound, got %v", err)
	}
}

func TestPublishFailureIsInternalAndRetriable(t *testing.T) {
	_, err := newSvc(&fakePub{err: errors.New("bus down")}).Create(context.Background(), "An")
	e := errs.From(err)
	if e.Kind != errs.KindInternal || !e.Retriable {
		t.Fatalf("want retriable Internal, got %+v", e)
	}
}
