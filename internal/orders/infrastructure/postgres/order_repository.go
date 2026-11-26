package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Ale96Glz/ddd-ecommerce-go/internal/orders/domain"
	"github.com/Ale96Glz/ddd-ecommerce-go/internal/orders/infrastructure/sql/db"
	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/events"
	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/money"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type OrderRepository struct {
	q *db.Queries
}

func NewOrderRepository(dbtx db.DBTX) *OrderRepository {
	return &OrderRepository{q: db.New(dbtx)}
}

func (r *OrderRepository) Save(ctx context.Context, order domain.Order) error {
	orderID, err := toPgUUID(order.ID().String())
	if err != nil {
		return err
	}
	customerID, err := toPgUUID(order.CustomerID().String())
	if err != nil {
		return err
	}

	if err := r.q.InsertOrder(ctx, db.InsertOrderParams{
		ID:         orderID,
		CustomerID: customerID,
		TotalCents: order.Total().Cents(),
		Currency:   order.Total().Currency(),
		Status:     string(order.Status()),
		Version:    int64(order.Version()),
	}); err != nil {
		return err
	}

	for _, item := range order.Items() {
		productID, err := toPgUUID(item.ProductID().String())
		if err != nil {
			return err
		}
		if err := r.q.InsertOrderItem(ctx, db.InsertOrderItemParams{
			OrderID:        orderID,
			ProductID:      productID,
			UnitPriceCents: item.UnitPrice().Cents(),
			Currency:       item.UnitPrice().Currency(),
			Quantity:       int32(item.Quantity()),
		}); err != nil {
			return err
		}
	}

	for _, event := range order.Events() {
		payload, err := eventPayload(event)
		if err != nil {
			return err
		}
		eventID, err := toPgUUID(event.EventID())
		if err != nil {
			return err
		}
		if err := r.q.InsertOutboxEvent(ctx, db.InsertOutboxEventParams{
			EventID:     eventID,
			AggregateID: event.AggregateID(),
			Type:        event.Type(),
			Payload:     payload,
			OccurredAt:  pgtype.Timestamptz{Time: event.OccurredAt(), Valid: true},
			TraceID:     event.TraceID(),
		}); err != nil {
			return err
		}
	}
	return nil
}

type orderEventPayload struct {
	OrderID string       `json:"orderId"`
	Total   *money.Money `json:"total,omitempty"`
}

func eventPayload(event events.DomainEvent) ([]byte, error) {
	switch e := event.(type) {
	case domain.OrderPlaced:
		return json.Marshal(orderEventPayload{OrderID: e.OrderID.String(), Total: &e.Total})
	case domain.OrderPaid:
		return json.Marshal(orderEventPayload{OrderID: e.OrderID.String()})
	case domain.OrderCancelled:
		return json.Marshal(orderEventPayload{OrderID: e.OrderID.String()})
	default:
		return nil, fmt.Errorf("Tipo de evento no soportado: %q", event.Type())
	}
}

func toPgUUID(s string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}
