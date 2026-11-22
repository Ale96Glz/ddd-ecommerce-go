package money_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Ale96Glz/ddd-ecommerce-go/internal/shared/money"
)

func TestNewMoney_RejectsNegativeAmount(t *testing.T) {
	if _, err := money.NewMoney(-1, "EUR"); !errors.Is(err, money.ErrNegativeAmount) {
		t.Fatalf("got %v, want ErrNegativeAmount", err)
	}
}

func TestNewMoney_RejectsEmptyCurrency(t *testing.T) {
	if _, err := money.NewMoney(100, ""); !errors.Is(err, money.ErrInvalidCurrency) {
		t.Fatalf("got %v, want ErrInvalidCurrency", err)
	}
}

func TestMoney_Add_CurrencyMismatch(t *testing.T) {
	a, _ := money.NewMoney(100, "EUR")
	b, _ := money.NewMoney(1, "USD")

	if _, err := a.Add(b); !errors.Is(err, money.ErrCurrencyMismatch) {
		t.Fatalf("got %v, want ErrCurrencyMismatch", err)
	}
}

func TestMoney_AddAndSubtract(t *testing.T) {
	a, _ := money.NewMoney(3000, "EUR")
	b, _ := money.NewMoney(500, "EUR")

	sum, err := a.Add(b)
	if err != nil || sum.Cents() != 3500 {
		t.Fatalf("Add = %v, %v; want 3500", sum, err)
	}
	rest, err := a.Subtract(b)
	if err != nil || rest.Cents() != 2500 {
		t.Fatalf("Subtract = %v, %v; want 2500", rest, err)
	}
}

func TestMoney_Equals(t *testing.T) {
	a, _ := money.NewMoney(3000, "EUR")
	b, _ := money.NewMoney(3000, "EUR")
	c, _ := money.NewMoney(3000, "USD")

	if !a.Equals(b) {
		t.Fatal("same cents and currency must be equal")
	}
	if a.Equals(c) {
		t.Fatal("different currencies must not be equal")
	}
}

func TestMoney_JSONRoundTrip(t *testing.T) {
	original, _ := money.NewMoney(3000, "EUR")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"amount":3000,"currency":"EUR"}`; string(data) != want {
		t.Fatalf("JSON = %s, want %s", data, want)
	}

	var decoded money.Money
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equals(original) {
		t.Fatalf("round trip = %v, want %v", decoded, original)
	}
}

func TestMoney_UnmarshalJSON_RejectsNegativeAmount(t *testing.T) {
	var decoded money.Money
	err := json.Unmarshal([]byte(`{"amount":-1,"currency":"EUR"}`), &decoded)
	if !errors.Is(err, money.ErrNegativeAmount) {
		t.Fatalf("got %v, want ErrNegativeAmount", err)
	}
}
