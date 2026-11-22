package domain

import (
	"time"

	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/events"
	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/id"
	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/money"
	"github.com/google/uuid"
)

type event struct {
	eventID     string
	aggregateID string
	occurredAt  time.Time
	traceID     string
}

func newEvent(aggregateID string) event {
	return event{
		eventID:     uuid.Must(uuid.NewV7()).String(),
		aggregateID: aggregateID,
		occurredAt:  time.Now().UTC(),
	}
}

func (e event) EventID() string       { return e.eventID }
func (e event) AggregateID() string   { return e.aggregateID }
func (e event) OccurredAt() time.Time { return e.occurredAt }
func (e event) TraceID() string       { return e.traceID }
func (e event) SchemaVersion() int    { return 1 }

type OrderPlaced struct {
	event
	OrderID id.OrderID
	Total   money.Money
}

func (OrderPlaced) Type() string { return "order.placed" }

type OrderPaid struct {
	event
	OrderID id.OrderID
}

func (OrderPaid) Type() string { return "order.paid" }

type OrderCancelled struct {
	event
	OrderID id.OrderID
}

func (OrderCancelled) Type() string { return "order.cancelled" }

var (
	_ events.DomainEvent = OrderPlaced{}
	_ events.DomainEvent = OrderPaid{}
	_ events.DomainEvent = OrderCancelled{}
)
