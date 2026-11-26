package postgres

import (
	"context"

	"github.com/Ale96Glz/ddd-ecommerce-go/internal/orders/infrastructure/sql/db"
	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/outbox"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxStore struct {
	pool *pgxpool.Pool
}

func NewOutboxStore(pool *pgxpool.Pool) *OutboxStore {
	return &OutboxStore{pool: pool}
}

// WithPending mantiene abierta la transacción con SKIP LOCKED mientras se publica.
func (s *OutboxStore) WithPending(ctx context.Context, limit int, fn func([]outbox.Event, outbox.MarkFunc) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	rows, err := db.New(tx).FetchUnpublishedEvents(ctx, int32(limit))
	if err != nil {
		return err
	}

	pending := make([]outbox.Event, 0, len(rows))
	for _, row := range rows {
		pending = append(pending, outbox.Event{
			ID:          uuid.UUID(row.EventID.Bytes).String(),
			AggregateID: row.AggregateID,
			Type:        row.Type,
			Payload:     row.Payload,
			OccurredAt:  row.OccurredAt.Time,
			TraceID:     row.TraceID,
		})
	}

	mark := func(ctx context.Context, eventID string) error {
		id, err := toPgUUID(eventID)
		if err != nil {
			return err
		}
		return db.New(tx).MarkOutboxPublished(ctx, id)
	}
	if err := fn(pending, mark); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
