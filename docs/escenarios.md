# Escenarios de compra y contrato HTTP

Derivados de `ecommerce-microservices/docs/escenarios-de-compra.md`, con el
contrato nuevo: rutas `/v1/`, IDs UUID v7, `Money` en centavos por dentro y
`application/problem+json` en los errores.

Convenciones:

| Tema | Valor |
|---|---|
| Base | `http://localhost:8080` |
| Auth | `Authorization: Bearer <jwt>` |
| Rutas públicas | `POST /v1/auth/register`, `POST /v1/auth/login` |
| Errores | `application/problem+json` (RFC 9457) |
| Idempotencia | Header `Idempotency-Key` en todo POST de dinero |
| Concurrencia | `If-Match: <etag>` en mutaciones de carrito/stock |

> Los ejemplos son la **especificación del comportamiento esperado**, no código
> ejecutable todavía. El hito 4 convertirá esto en tests `httptest`.

## Escenario 1 — Compra aprobada (happy path)

1. `POST /v1/auth/register` → `201` con `token`.
2. `POST /v1/categories` y `POST /v1/products` (ADMIN) → `201`.
   El `price` se envía como `{ "amount": 3000, "currency": "EUR" }` (centavos).
3. `POST /v1/cart/items` con `Idempotency-Key` e `If-Match` → `201`, `ETag` nuevo.
4. `POST /v1/orders` → `201`:
   - `total` se **calcula** (`3 × 3000 = 9000` centavos), nunca se envía.
   - stock `50 → 47` en la misma transacción.
   - se escribe `order.placed` en el outbox.
5. La saga `PlaceOrder` avanza: `ReserveStock → AuthorizePayment → ConfirmOrder`.
6. `amount = 9000 ≤ umbral` → pago `PROCESADO`, pedido `PAGADO`.
7. `GET /v1/orders/{id}/events` (SSE) emite `order.placed`, `payment.processed`,
   `order.paid`.
8. La notificación dice: `Tu pedido #<id> fue confirmado. Total: 90.00 EUR`.

Ejemplo de error de pago (imposible en el modelo, se muestra para documentar el
contrato del borde):

```http
POST /v1/orders/{id}/payment
Idempotency-Key: 7f3c...
Content-Type: application/json

{ "amount": { "amount": 9100, "currency": "EUR" } }
```

```http
HTTP/1.1 422 Unprocessable Content
Content-Type: application/problem+json

{
  "type": "https://errors.ecommerce.test/amount-mismatch",
  "title": "El monto no coincide con el total del pedido",
  "status": 422,
  "detail": "El monto 91.00 no coincide con el total del pedido 90.00",
  "instance": "/v1/orders/…/payment"
}
```

> Diferencia clave frente al repo Java: allí el `422` se lograba con un `if` en
> `PagoServiceImpl`; aquí el comando de pago **no acepta un `amount` distinto**
> porque toma el `total` del agregado (FR-06, [ADR-0005](./adr/ADR-0005-money-value-object.md)).

## Escenario 2 — Pago rechazado y compensación (saga)

1. Pedido con `total = 12000` centavos (`> umbral 10000`).
2. `AuthorizePayment` devuelve `RECHAZADO` con
   `reason = "Monto excede el máximo aprobado (100.00)"`.
3. La saga ejecuta la compensación `ReleaseStock`:
   - se persiste el cambio de estado de la saga **y** el evento
     `stock.release_requested` en la misma transacción.
   - el relay publica; el consumidor aplica `Release` y registra `event_id`.
   - stock `43 → 47` **exactamente una vez**, aunque el evento se reentregue.
4. Pedido `CANCELADO`, pago `RECHAZADO` persistido, notificación con el motivo.

Test de aceptación del hito 3: inyectar un fallo en el paso 2 y verificar que
`ReleaseStock` se ejecuta una sola vez y que la saga es reanudable.

## Escenario 3 — Cancelación manual

`DELETE /v1/orders/{id}` (o `POST /cancel`) solo desde `CREATED`:
misma compensación `ReleaseStock`. Desde `PAID` → `409 problem+json`.

## Escenario 4 — Reintentos e idempotencia

| Acción | Resultado esperado |
|---|---|
| Repetir `POST /v1/orders` con el mismo `Idempotency-Key` y payload | Mismo `201` y el mismo `order_id`, sin descontar stock dos veces |
| Repetir con la misma key pero payload distinto | `409 problem+json` |
| Reentregar `stock.release_requested` | Stock unchanged (dedup por `event_id`) |
| `If-Match` obsoleto en el carrito | `412 Precondition Failed` |
| Sin JWT | `401` |
| Rol `USER` creando producto | `403` |
| `amount` con moneda distinta a la del pedido | `422` (`ErrCurrencyMismatch`) |

## Taxonomía de errores (resumen)

| Código | Significado |
|---|---|
| `400` | Forma del payload inválida (JSON, tipos, `validator/v10`) |
| `401` / `403` | Autenticación / autorización |
| `404` | Recurso inexistente |
| `409` | Conflicto de estado del recurso o idem-key reutilizada |
| `412` | `If-Match` obsoleto (concurrencia optimista) |
| `422` | Regla de negocio del contenido (`amount ≠ total`, moneda, etc.) |

## Relación con el repo Java

| Concepto Java | Equivalente aquí |
|---|---|
| `POST /api/v1/pago` con `monto` | Comando derivado del `Order.total` |
| `{message, error, status}` | RFC 9457 `problem+json` |
| Kafka + outbox + DLT | Kafka + outbox + DLT, con `franz-go` y contrato versionado |
| `RESTOCK_REQUIRED` + `restock_event` | `stock.release_requested` + `processed_event` |
| Saga implícita por listeners | `PlaceOrderSaga` explícita y persistida |
| `BigDecimal` | `Money{cents int64, currency}` |
