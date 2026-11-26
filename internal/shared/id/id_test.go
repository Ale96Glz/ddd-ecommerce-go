package id_test

import (
	"errors"
	"testing"

	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/id"
)

func TestParseOrderID_ValidV7(t *testing.T) {
	generated := id.NewOrderID()

	parsed, err := id.ParseOrderID(generated.String())
	if err != nil {
		t.Fatal(err)
	}
	if parsed != generated {
		t.Fatalf("Analizado %v, se esperaba %v", parsed, generated)
	}
}

func TestParseOrderID_RejectsV4(t *testing.T) {
	_, err := id.ParseOrderID("f47ac10b-58cc-4372-a567-0e02b2c3d479")
	if !errors.Is(err, id.ErrInvalidID) {
		t.Fatalf("Se obtuvo %v, se esperaba ErrInvalidID", err)
	}
}

func TestParseOrderID_RejectsInvalidString(t *testing.T) {
	_, err := id.ParseOrderID("not-a-uuid")
	if !errors.Is(err, id.ErrInvalidID) {
		t.Fatalf("Se obtuvo %v, se esperaba ErrInvalidID", err)
	}
}

func TestNewProductIDAndCustomerID_AreV7(t *testing.T) {
	if _, err := id.ParseProductID(id.NewProductID().String()); err != nil {
		t.Fatal(err)
	}
	if _, err := id.ParseCustomerID(id.NewCustomerID().String()); err != nil {
		t.Fatal(err)
	}
}
