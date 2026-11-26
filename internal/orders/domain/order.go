package domain

import (
	"errors"

	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/events"
	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/id"
	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/money"
)

var (
	ErrInvalidQuantity = errors.New("La cantidad debe estar entre 1 y 1000000")
	ErrEmptyOrder      = errors.New("El pedido debe tener al menos un ítem")
	ErrAmountMismatch  = errors.New("El monto no coincide con el total del pedido")
	ErrInvalidState    = errors.New("El pedido no está en un estado modificable")
)

type OrderStatus string

const (
	StatusCreated   OrderStatus = "Created"
	StatusPaid      OrderStatus = "Paid"
	StatusCancelled OrderStatus = "Cancelled"
)

type OrderItem struct {
	productID id.ProductID
	unitPrice money.Money
	quantity  uint32
}

func NewOrderItem(productID id.ProductID, unitPrice money.Money, quantity uint32) (OrderItem, error) {
	if quantity == 0 || quantity > 1_000_000 {
		return OrderItem{}, ErrInvalidQuantity
	}
	return OrderItem{productID: productID, unitPrice: unitPrice, quantity: quantity}, nil
}

type Order struct {
	id         id.OrderID
	customerID id.CustomerID
	items      []OrderItem
	total      money.Money
	status     OrderStatus
	version    int
	events     []events.DomainEvent
}

func NewOrder(orderID id.OrderID, customerID id.CustomerID, items []OrderItem) (Order, error) {
	if len(items) == 0 {
		return Order{}, ErrEmptyOrder
	}

	total, err := money.NewMoney(0, items[0].unitPrice.Currency())
	if err != nil {
		return Order{}, err
	}
	for _, item := range items {
		lineTotal, err := money.NewMoney(item.unitPrice.Cents()*int64(item.quantity), item.unitPrice.Currency())
		if err != nil {
			return Order{}, err
		}
		total, err = total.Add(lineTotal)
		if err != nil {
			return Order{}, err
		}
	}

	o := Order{
		id:         orderID,
		customerID: customerID,
		items:      append([]OrderItem(nil), items...),
		total:      total,
		status:     StatusCreated,
		version:    1,
	}
	o.events = append(o.events, OrderPlaced{
		event:   newEvent(orderID.String()),
		OrderID: orderID,
		Total:   total,
	})
	return o, nil
}

func (o *Order) Pay(amount money.Money) error {
	if o.status != StatusCreated {
		return ErrInvalidState
	}
	if !amount.Equals(o.total) {
		return ErrAmountMismatch
	}
	o.status = StatusPaid
	o.version++
	o.events = append(o.events, OrderPaid{
		event:   newEvent(o.id.String()),
		OrderID: o.id,
	})
	return nil
}

func (o *Order) Cancel() error {
	if o.status != StatusCreated {
		return ErrInvalidState
	}
	o.status = StatusCancelled
	o.version++
	o.events = append(o.events, OrderCancelled{
		event:   newEvent(o.id.String()),
		OrderID: o.id,
	})
	return nil
}

func (o Order) ID() id.OrderID               { return o.id }
func (o Order) CustomerID() id.CustomerID    { return o.customerID }
func (o Order) Items() []OrderItem           { return append([]OrderItem(nil), o.items...) }
func (o Order) Total() money.Money           { return o.total }
func (o Order) Status() OrderStatus          { return o.status }
func (o Order) Version() int                 { return o.version }
func (o Order) Events() []events.DomainEvent { return append([]events.DomainEvent(nil), o.events...) }

func (i OrderItem) ProductID() id.ProductID { return i.productID }
func (i OrderItem) UnitPrice() money.Money  { return i.unitPrice }
func (i OrderItem) Quantity() uint32        { return i.quantity }
