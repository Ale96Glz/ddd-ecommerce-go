# Requisitos

Los requisitos se derivan del dominio del repo Java (`docs/escenarios-de-compra.md`,
ADRs 0013–0020) pero se **rediseñan** con el contrato nuevo. Identificadores
estables (`FR-xx`, `NFR-xx`) para poder citarlos desde los ADRs y los tests.

## Requisitos funcionales

| ID | Requisito | Criterio de aceptación |
|---|---|---|
| **FR-01** | Registro y login de usuario con JWT | `POST /v1/auth/register` y `/login` devuelven un token; el resto de rutas lo exigen |
| **FR-02** | Catálogo: categorías y productos | Solo `ADMIN` crea; `GET` público; el producto tiene `precio` (`Money`) y `stock` |
| **FR-03** | Carrito con concurrencia optimista | `POST /v1/cart/items` exige `Idempotency-Key` e `If-Match` (ETag); un ETag obsoleto → `412` |
| **FR-04** | Crear pedido reserva stock | `POST /v1/orders` descuenta stock **en la misma transacción** y publica `OrderPlaced` en el outbox |
| **FR-05** | El total del pedido se calcula, nunca se recibe | El cliente envía cantidades; el servidor calcula `total = Σ(precio × cantidad)` |
| **FR-06** | Pago con `monto = total` por construcción | El comando de pago no acepta un `monto` arbitrario: se deriva del pedido. Un `amount` distinto es imposible en el tipo → `422 problem+json` si se intenta en el borde |
| **FR-07** | Decisión de aprobación | `amount ≤ umbral` → `PROCESADO`; si no, `RECHAZADO` con `motivo`. El pago se persiste siempre (auditoría) |
| **FR-08** | Saga `PlaceOrder` como process manager | `ReserveStock → AuthorizePayment → ConfirmOrder`; cada paso tiene compensación declarada (`ReleaseStock`, `VoidPayment`); el estado se persiste y es reanudable |
| **FR-09** | Outbox con relay propio | Eventos y cambios de estado de la saga en la misma transacción; el relay publica a Kafka con `FOR UPDATE SKIP LOCKED` y marca `PUBLICADO` solo tras el ack del broker |
| **FR-10** | Idempotencia de consumidores | Tabla `processed_event(event_id PK)`; reentrega/dedup exactamente una vez por evento |
| **FR-11** | Idempotencia de HTTP | `Idempotency-Key` en todo POST de dinero; se persiste y se replaya la respuesta original, incluido el status |
| **FR-12** | Cancelación de pedido | Un pedido `CREADO` se puede cancelar → compensación de stock; `PAGADO` no se cancela → `409` |
| **FR-13** | Estado en vivo por SSE | `GET /v1/orders/{id}/events` emite cambios de estado del pedido |
| **FR-14** | Webhooks salientes | Registro de endpoints del cliente; entregas firmadas HMAC, reintentos y log de entregas |
| **FR-15** | Notificaciones al usuario | El usuario ve el resultado del pedido/pago con el motivo legible |

## Requisitos no funcionales

| ID | Requisito | Criterio de aceptación |
|---|---|---|
| **NFR-01** | Dinero correcto | `amount_cents BIGINT` + `currency CHAR(3)`; cero `float`, cero `BigDecimal`, cero errores de redondeo |
| **NFR-02** | Invariantes en el dominio | Ninguna regla de negocio depende de la capa HTTP; los tests de dominio no importan `net/http` ni DB |
| **NFR-03** | Contrato primero | `api/openapi.yaml` es la única fuente de tipos HTTP; `oapi-codegen` genera handlers y clientes en el build |
| **NFR-04** | SQL tipado | `sqlc` genera código en tiempo de build; cero ORM; migraciones `goose` versionadas y reversibles |
| **NFR-05** | Errores consistentes | Un único `AppError{Code, Message}` se mapea a `application/problem+json`; taxonomía `400` forma / `409` estado / `422` regla de negocio / `412` ETag |
| **NFR-06** | Observabilidad | Logs JSON con `slog`, trazas OpenTelemetry → OTLP, `trace_id` propagado dentro de los eventos del outbox |
| **NFR-07** | Seguridad | JWT `jwt/v5`, passwords con `argon2id`, validación solo en el borde HTTP |
| **NFR-08** | Tests por capa | Dominio: `testing` puro. HTTP: `httptest`. Integración: `testcontainers-go`. Aserciones: `go-cmp`. Sin `sqlmock` |
| **NFR-09** | Arranque y apagado limpio | `context` + señales `SIGINT`/`SIGTERM`; el relay termina su lote y el bus se cierra sin perder eventos |
| **NFR-10** | Reproducibilidad | Un `go.mod`, `Dockerfile` multi-stage → binario estático distroless, `docker compose up` levanta API + worker + Postgres |
| **NFR-11** | Aislamiento por servicio | Un proceso y una BD por contexto; cero tablas compartidas; tres servicios en el MVP (ADR-0002) |
| **NFR-12** | Documentación viva | Todo cambio de diseño actualiza su ADR y este fichero en el mismo commit |

## Trazabilidad

Cada ADR declara qué requisitos toca. El mapa inverso se mantiene en
[`docs/adr/README.md`](./adr/README.md).

| Requisito | ADR que lo decide |
|---|---|
| FR-04, FR-08, FR-09, FR-10 | [ADR-0007](./adr/ADR-0007-outbox-relay-propio.md), [ADR-0008](./adr/ADR-0008-saga-process-manager.md) |
| FR-05, FR-06, NFR-01 | [ADR-0005](./adr/ADR-0005-money-value-object.md) |
| FR-11, FR-03, NFR-05 | [ADR-0009](./adr/ADR-0009-idempotencia-y-concurrencia.md) |
| FR-07, NFR-05 | [ADR-0010](./adr/ADR-0010-errores-rfc9457.md) |
| NFR-03 | [ADR-0004](./adr/ADR-0004-contrato-openapi.md) |
| NFR-04 | [ADR-0006](./adr/ADR-0006-pgx-sqlc-goose.md) |
| NFR-06 | [ADR-0012](./adr/ADR-0012-observabilidad.md) |
| NFR-08 | [ADR-0013](./adr/ADR-0013-estrategia-de-tests.md) |
| NFR-02 | [ADR-0015](./adr/ADR-0015-lenguaje-ubicuo-y-nomenclatura.md) |
| NFR-11 | [ADR-0002](./adr/ADR-0002-microservicios-por-contextos.md), [ADR-0007](./adr/ADR-0007-outbox-relay-propio.md) |
