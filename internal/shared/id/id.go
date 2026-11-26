package id

import (
	"errors"

	"github.com/google/uuid"
)

var ErrInvalidID = errors.New("El identificador no es válido")

type OrderID uuid.UUID
type ProductID uuid.UUID
type CustomerID uuid.UUID

func NewOrderID() OrderID { return OrderID(uuid.Must(uuid.NewV7())) }

func ParseOrderID(s string) (OrderID, error) {
	u, err := parseUUIDv7(s)
	if err != nil {
		return OrderID{}, err
	}
	return OrderID(u), nil
}

func (v OrderID) String() string { return uuid.UUID(v).String() }

func NewProductID() ProductID { return ProductID(uuid.Must(uuid.NewV7())) }

func ParseProductID(s string) (ProductID, error) {
	u, err := parseUUIDv7(s)
	if err != nil {
		return ProductID{}, err
	}
	return ProductID(u), nil
}

func (v ProductID) String() string { return uuid.UUID(v).String() }

func NewCustomerID() CustomerID { return CustomerID(uuid.Must(uuid.NewV7())) }

func ParseCustomerID(s string) (CustomerID, error) {
	u, err := parseUUIDv7(s)
	if err != nil {
		return CustomerID{}, err
	}
	return CustomerID(u), nil
}

func (v CustomerID) String() string { return uuid.UUID(v).String() }

func parseUUIDv7(s string) (uuid.UUID, error) {
	u, err := uuid.Parse(s)
	if err != nil || u.Version() != 7 {
		return uuid.Nil, ErrInvalidID
	}
	return u, nil
}
