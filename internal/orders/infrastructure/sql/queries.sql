-- name: InsertOrder :exec
INSERT INTO orders (id, customer_id, total_cents, currency, status, version)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: InsertOrderItem :exec
INSERT INTO order_items (order_id, product_id, unit_price_cents, currency, quantity)
VALUES ($1, $2, $3, $4, $5);

-- name: InsertOutboxEvent :exec
INSERT INTO outbox_events (event_id, aggregate_id, type, payload, occurred_at, trace_id)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: FetchUnpublishedEvents :many
SELECT event_id, aggregate_id, type, payload, occurred_at, trace_id
FROM outbox_events
WHERE published_at IS NULL
ORDER BY occurred_at
FOR UPDATE SKIP LOCKED
LIMIT $1;

-- name: MarkOutboxPublished :exec
UPDATE outbox_events SET published_at = now() WHERE event_id = $1;

-- name: InsertProcessedEvent :exec
INSERT INTO processed_event (event_id) VALUES ($1) ON CONFLICT DO NOTHING;
