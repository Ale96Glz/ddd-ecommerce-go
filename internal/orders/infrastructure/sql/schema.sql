-- shortcut: duplicado de 000001_create_orders.up.sql porque sqlc no interpreta las anotaciones goose; actualizar ambos hasta migrar a un generador común.

CREATE TABLE orders (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL,
    total_cents BIGINT NOT NULL,
    currency CHAR(3) NOT NULL,
    status TEXT NOT NULL,
    version BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
    order_id UUID NOT NULL REFERENCES orders(id),
    product_id UUID NOT NULL,
    unit_price_cents BIGINT NOT NULL,
    currency CHAR(3) NOT NULL,
    quantity INTEGER NOT NULL,
    PRIMARY KEY (order_id, product_id)
);

CREATE TABLE outbox_events (
    event_id UUID PRIMARY KEY,
    aggregate_id TEXT NOT NULL,
    type TEXT NOT NULL,
    payload JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    trace_id TEXT NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ
);

CREATE INDEX outbox_events_pending_idx ON outbox_events (occurred_at) WHERE published_at IS NULL;

CREATE TABLE processed_event (
    event_id UUID PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
