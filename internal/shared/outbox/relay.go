package outbox

import (
	"context"
	"time"
)

const DefaultBatchSize = 10

type Event struct {
	ID          string
	AggregateID string
	Type        string
	Payload     []byte
	OccurredAt  time.Time
	TraceID     string
}

type MarkFunc func(ctx context.Context, eventID string) error

type Store interface {
	WithPending(ctx context.Context, limit int, fn func([]Event, MarkFunc) error) error
}

type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

type Relay struct {
	store     Store
	publisher Publisher
	batchSize int
}

func New(store Store, publisher Publisher) *Relay {
	return &Relay{store: store, publisher: publisher, batchSize: DefaultBatchSize}
}

// RunOnce publica un lote y marca cada evento dentro de la transacción del almacén.
func (r *Relay) RunOnce(ctx context.Context) (int, error) {
	published := 0
	err := r.store.WithPending(ctx, r.batchSize, func(events []Event, mark MarkFunc) error {
		for _, event := range events {
			if err := r.publisher.Publish(ctx, event); err != nil {
				return err
			}
			if err := mark(ctx, event.ID); err != nil {
				return err
			}
			published++
		}
		return nil
	})
	return published, err
}
