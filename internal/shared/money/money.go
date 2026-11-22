package money

import (
	"cmp"
	"encoding/json"
	"errors"
)

var (
	ErrNegativeAmount   = errors.New("amount must not be negative")
	ErrInvalidCurrency  = errors.New("invalid currency")
	ErrCurrencyMismatch = errors.New("currency mismatch")
)

// Money is an immutable amount in cents with an ISO 4217 currency.
type Money struct {
	cents    int64
	currency string
}

func NewMoney(cents int64, currency string) (Money, error) {
	if cents < 0 {
		return Money{}, ErrNegativeAmount
	}
	if currency == "" {
		return Money{}, ErrInvalidCurrency
	}
	return Money{cents: cents, currency: currency}, nil
}

func (m Money) Cents() int64     { return m.cents }
func (m Money) Currency() string { return m.currency }

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrCurrencyMismatch
	}
	return NewMoney(m.cents+other.cents, m.currency)
}

func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrCurrencyMismatch
	}
	return NewMoney(m.cents-other.cents, m.currency)
}

func (m Money) Equals(other Money) bool { return m == other }
func (m Money) IsZero() bool            { return m.cents == 0 }

func (m Money) Compare(other Money) int {
	if c := cmp.Compare(m.currency, other.currency); c != 0 {
		return c
	}
	return cmp.Compare(m.cents, other.cents)
}

type moneyJSON struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(moneyJSON{Amount: m.cents, Currency: m.currency})
}

func (m *Money) UnmarshalJSON(data []byte) error {
	var raw moneyJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	parsed, err := NewMoney(raw.Amount, raw.Currency)
	if err != nil {
		return err
	}
	*m = parsed
	return nil
}
