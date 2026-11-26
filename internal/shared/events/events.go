package events

import "time"

// DomainEvent es la envelope compartida por todos los contextos. Es inmutable.
type DomainEvent interface {
	EventID() string
	AggregateID() string
	OccurredAt() time.Time
	TraceID() string
	SchemaVersion() int
	Type() string
}
