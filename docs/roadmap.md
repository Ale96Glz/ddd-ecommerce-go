# Roadmap / hitos

Regla: **un hito terminado gana siempre sobre seis a medias**. Cada hito tiene un
test de aceptación observable. No se empieza el siguiente hasta que el anterior
pasa.

## Hito 0 — Documentación ✅

**Entregable:** visión, requisitos, lenguaje ubicuo, dominio, catálogo de value
objects, catálogo de eventos/comandos, stack, escenarios, ADRs.

**Aceptación:** cada ADR tiene contexto → decisión → consecuencias; cada
requisito tiene un ADR que lo decide; cada VO tiene invariantes + JSON + SQL;
cada término del glosario aparece con un único identificador en código.

## Hito 1 — Núcleo de dominio puro

**Alcance:** `Money`, IDs tipados, `Order` con `Pay`, eventos de dominio, errores
de dominio. Sin DB, sin HTTP, sin bus.

**Test de aceptación:**

```bash
go test ./internal/shared/money ./internal/orders/domain
```

- `NewMoney(-1, "EUR")` → error.
- `NewMoney(100, "EUR").Add(NewMoney(1, "USD"))` → `ErrCurrencyMismatch`.
- `order.Pay(total)` → `OrderPaid` emitido.
- `order.Pay(total+1)` → `ErrAmountMismatch`, **cero eventos**, estado intacto.
- Los tests no importan `net/http`, `pgx` ni `chi`.

## Hito 2 — Persistencia y outbox

**Alcance:** un servicio por contexto con su `goose` + `sqlc` + `pgxpool`,
repositorios de `Order` y `Payment`, tabla `outbox_events`, relay con
`FOR UPDATE SKIP LOCKED` publicando a Kafka (`franz-go`) y `processed_event`
para dedup en la BD del consumidor.

**Test de aceptación:** `testcontainers-go` levanta Postgres y Kafka (KRaft); un
test guarda un `Order` y su evento en la misma transacción, ejecuta el relay y
comprueba que Kafka recibió el evento exactamente una vez y que `published_at`
quedó fijado. Un segundo relay no reenvía nada.

## Hito 3 — Saga `PlaceOrder` (el corazón)

**Alcance:** `PlaceOrderSaga` con `ReserveStock → AuthorizePayment →
ConfirmOrder`, compensaciones `ReleaseStock` / `VoidPayment`, estado persistido.

**Test de aceptación:** inyectar fallo en `AuthorizePayment` → la saga ejecuta
`ReleaseStock` una sola vez y termina en `Compensated`; reanudar la saga desde el
estado persistido tras simular crash no duplica la compensación.

## Hito 4 — Capa HTTP contract-first

**Alcance:** `api/openapi.yaml`, `oapi-codegen`, `chi/v5`, middleware (request ID,
`slog`, recover, `AppError` → `problem+json`), `Idempotency-Key`, `If-Match`/ETag.

**Test de aceptación:** `httptest` sobre el handler generado:

- happy path del Escenario 1 → `201`.
- `amount ≠ total` → `422 problem+json` con `type` y `detail`.
- mismo `Idempotency-Key` → misma respuesta; payload distinto → `409`.
- `If-Match` obsoleto → `412`.

## Hito 5 — Worker, SSE y webhooks

**Alcance:** modo `-worker` por servicio (relay + consumidores de su BD);
`GET /v1/orders/{id}/events` por SSE; registro de webhooks con firma HMAC y
reintentos; graceful shutdown de consumidores Kafka.

**Test de aceptación:** dos peticiones SSE reciben el mismo evento; un webhook
firmado con el secret correcto se entrega y se registra; `SIGTERM` termina el lote
del relay sin eventos pendientes huérfanos.

## Fuera del roadmap (por ahora)

- Redis, Kubernetes, frontend, multimoneda con FX, membership de saga.
- Métricas de negocio avanzadas (solo las de OTel incluidas en NFR-06).
