package events

import "time"

// DomainEvent is the envelope shared by every context. It is immutable.
type DomainEvent interface {
	EventID() string
	AggregateID() string
	OccurredAt() time.Time
	TraceID() string
	SchemaVersion() int
	Type() string
}
