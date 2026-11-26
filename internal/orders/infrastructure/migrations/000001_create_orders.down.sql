-- +goose Down
DROP INDEX IF EXISTS outbox_events_pending_idx;
DROP TABLE IF EXISTS processed_event;
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
