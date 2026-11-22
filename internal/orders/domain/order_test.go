package domain_test

import (
	"errors"
	"testing"

	"github.com/Ale96Glz/ddd-ecommerce-go/internal/orders/domain"
	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/id"
	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/money"
)

func newOrder(t *testing.T, cents int64, quantity uint32) domain.Order {
	t.Helper()

	price, err := money.NewMoney(cents, "EUR")
	if err != nil {
		t.Fatal(err)
	}
	item, err := domain.NewOrderItem(id.NewProductID(), price, quantity)
	if err != nil {
		t.Fatal(err)
	}
	order, err := domain.NewOrder(id.NewOrderID(), id.NewCustomerID(), []domain.OrderItem{item})
	if err != nil {
		t.Fatal(err)
	}
	return order
}

func TestOrder_New_CalculatesTotal(t *testing.T) {
	order := newOrder(t, 3000, 3)

	if got := order.Total().Cents(); got != 9000 {
		t.Fatalf("total = %d, want 9000", got)
	}
	if order.Status() != domain.StatusCreated {
		t.Fatalf("status = %s, want Created", order.Status())
	}
	if order.Version() != 1 {
		t.Fatalf("version = %d, want 1", order.Version())
	}
	events := order.Events()
	if len(events) != 1 || events[0].Type() != "order.placed" {
		t.Fatalf("events = %v, want one order.placed", events)
	}
}

func TestOrder_New_RejectsEmptyItems(t *testing.T) {
	_, err := domain.NewOrder(id.NewOrderID(), id.NewCustomerID(), nil)
	if !errors.Is(err, domain.ErrEmptyOrder) {
		t.Fatalf("got %v, want ErrEmptyOrder", err)
	}
}

func TestNewOrderItem_InvalidQuantity(t *testing.T) {
	price, _ := money.NewMoney(100, "EUR")

	if _, err := domain.NewOrderItem(id.NewProductID(), price, 0); !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Fatalf("got %v, want ErrInvalidQuantity", err)
	}
	if _, err := domain.NewOrderItem(id.NewProductID(), price, 1_000_001); !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Fatalf("got %v, want ErrInvalidQuantity", err)
	}
}

func TestOrder_Pay_EmitsOrderPaid(t *testing.T) {
	order := newOrder(t, 3000, 1)

	if err := order.Pay(order.Total()); err != nil {
		t.Fatal(err)
	}
	if order.Status() != domain.StatusPaid {
		t.Fatalf("status = %s, want Paid", order.Status())
	}
	if order.Version() != 2 {
		t.Fatalf("version = %d, want 2", order.Version())
	}

	events := order.Events()
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	paid, ok := events[1].(domain.OrderPaid)
	if !ok {
		t.Fatalf("second event = %T, want OrderPaid", events[1])
	}
	if paid.Type() != "order.paid" || paid.AggregateID() != order.ID().String() {
		t.Fatalf("event = %s / %s", paid.Type(), paid.AggregateID())
	}
	if paid.EventID() == "" || paid.SchemaVersion() != 1 {
		t.Fatalf("event id = %q, schema = %d", paid.EventID(), paid.SchemaVersion())
	}
}

func TestOrder_Pay_AmountMismatch(t *testing.T) {
	order := newOrder(t, 3000, 1)
	before := len(order.Events())

	wrong, _ := money.NewMoney(order.Total().Cents()+1, "EUR")
	err := order.Pay(wrong)
	if !errors.Is(err, domain.ErrAmountMismatch) {
		t.Fatalf("got %v, want ErrAmountMismatch", err)
	}
	if order.Status() != domain.StatusCreated {
		t.Fatalf("status = %s, want Created", order.Status())
	}
	if order.Version() != 1 {
		t.Fatalf("version = %d, want 1", order.Version())
	}
	if len(order.Events()) != before {
		t.Fatalf("got %d events, want %d", len(order.Events()), before)
	}
}

func TestOrder_Pay_InvalidState(t *testing.T) {
	order := newOrder(t, 3000, 1)
	if err := order.Pay(order.Total()); err != nil {
		t.Fatal(err)
	}

	if err := order.Pay(order.Total()); !errors.Is(err, domain.ErrInvalidState) {
		t.Fatalf("got %v, want ErrInvalidState", err)
	}
}

func TestOrder_Cancel_EmitsOrderCancelled(t *testing.T) {
	order := newOrder(t, 3000, 1)

	if err := order.Cancel(); err != nil {
		t.Fatal(err)
	}
	if order.Status() != domain.StatusCancelled {
		t.Fatalf("status = %s, want Cancelled", order.Status())
	}
	events := order.Events()
	if len(events) != 2 || events[1].Type() != "order.cancelled" {
		t.Fatalf("events = %v, want one order.cancelled", events)
	}
}

func TestOrder_Cancel_AfterPay_InvalidState(t *testing.T) {
	order := newOrder(t, 3000, 1)
	if err := order.Pay(order.Total()); err != nil {
		t.Fatal(err)
	}

	if err := order.Cancel(); !errors.Is(err, domain.ErrInvalidState) {
		t.Fatalf("got %v, want ErrInvalidState", err)
	}
}
