package postgres

import (
	"testing"

	"github.com/Ale96Glz/ddd-ecommerce-go/internal/orders/domain"
	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/id"
	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/money"
)

func TestEventPayload_OrderPlaced(t *testing.T) {
	price, _ := money.NewMoney(3000, "EUR")
	item, err := domain.NewOrderItem(id.NewProductID(), price, 1)
	if err != nil {
		t.Fatal(err)
	}
	order, err := domain.NewOrder(id.NewOrderID(), id.NewCustomerID(), []domain.OrderItem{item})
	if err != nil {
		t.Fatal(err)
	}

	payload, err := eventPayload(order.Events()[0])
	if err != nil {
		t.Fatal(err)
	}
	want := `{"orderId":"` + order.ID().String() + `","total":{"amount":3000,"currency":"EUR"}}`
	if string(payload) != want {
		t.Fatalf("Payload = %s, se esperaba %s", payload, want)
	}
}
