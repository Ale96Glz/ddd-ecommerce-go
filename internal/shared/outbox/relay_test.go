package outbox_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/outbox"
)

type fakeStore struct {
	events []outbox.Event
	marked []string
}

func (s *fakeStore) WithPending(ctx context.Context, limit int, fn func([]outbox.Event, outbox.MarkFunc) error) error {
	batch := s.events
	if len(batch) > limit {
		batch = batch[:limit]
	}
	return fn(batch, func(ctx context.Context, id string) error {
		s.marked = append(s.marked, id)
		return nil
	})
}

type fakePublisher struct {
	published []string
	err       error
}

func (p *fakePublisher) Publish(ctx context.Context, event outbox.Event) error {
	if p.err != nil {
		return p.err
	}
	p.published = append(p.published, event.ID)
	return nil
}

func TestRelay_RunOnce_PublishesThenMarks(t *testing.T) {
	store := &fakeStore{events: []outbox.Event{{ID: "e-1"}, {ID: "e-2"}}}
	publisher := &fakePublisher{}

	count, err := outbox.New(store, publisher).RunOnce(context.Background())
	if err != nil || count != 2 {
		t.Fatalf("RunOnce = %d, %v; se esperaban 2, nil", count, err)
	}
	if len(publisher.published) != 2 || len(store.marked) != 2 {
		t.Fatalf("Publicados = %v, marcados = %v", publisher.published, store.marked)
	}
	for i, id := range publisher.published {
		if id != store.marked[i] {
			t.Fatalf("Se marcó %q tras publicar %q", store.marked[i], id)
		}
	}
}

func TestRelay_RunOnce_PublishErrorLeavesEventPending(t *testing.T) {
	store := &fakeStore{events: []outbox.Event{{ID: "e-1"}}}
	publisher := &fakePublisher{err: errors.New("El broker no está disponible")}

	count, err := outbox.New(store, publisher).RunOnce(context.Background())
	if err == nil || count != 0 {
		t.Fatalf("RunOnce = %d, %v; se esperaban 0, error", count, err)
	}
	if len(store.marked) != 0 {
		t.Fatalf("Marcados = %v, se esperaba ninguno", store.marked)
	}
}

func TestRelay_RunOnce_SecondRunPublishesNothing(t *testing.T) {
	store := &fakeStore{events: []outbox.Event{{ID: "e-1"}}}
	publisher := &fakePublisher{}
	relay := outbox.New(store, publisher)

	if _, err := relay.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	store.events = nil // ya publicado: la consulta no devuelve pendientes

	count, err := relay.RunOnce(context.Background())
	if err != nil || count != 0 {
		t.Fatalf("Segundo RunOnce = %d, %v; se esperaban 0, nil", count, err)
	}
	if len(publisher.published) != 1 {
		t.Fatalf("Publicados = %v, se esperaba un evento", publisher.published)
	}
}
