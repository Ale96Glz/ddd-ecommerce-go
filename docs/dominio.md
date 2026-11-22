# Diseño del dominio (DDD táctico)

Este documento es la **fuente de verdad del modelo**. El código se escribe para
encajar aquí, no al revés. Los nombres salen del
[lenguaje ubicuo](./lenguaje-ubicuo.md); las invariantes de cada valor están en el
[catálogo de value objects](./value-objects.md) y los mensajes en el
[catálogo de eventos y comandos](./eventos-y-comandos.md).

## Bounded contexts

```
internal/
  catalog/        Product, Stock, Category          (agregado: Product)
  cart/           Cart, CartItem                    (agregado: Cart)
  orders/         Order, OrderItem, PlaceOrderSaga  (agregados: Order, Saga)
  payments/       Payment                           (agregado: Payment)
  notifications/  Notification, WebhookEndpoint     (agregados: Notification, Endpoint)
  shared/
    money/        Money        (value object)
    id/           OrderID, ProductID, ... (IDs tipados)
    events/       DomainEvent, EventBus (interfaces)
    outbox/       OutboxWriter, Relay
    saga/         ProcessManager (interfaz)
```

Cada contexto define **sus propias interfaces de repositorio donde se consumen**
(no en un `domain/` global compartido). No hay capas globales de aplicación.

> Nota: los agregados `Order` y `Saga` están en el mismo contexto `orders`, pero
> son dos raíces distintas con dos repositorios distintos. La saga no es un
> método del agregado. En el MVP, `cart` vive dentro de `orders-service` y
> `notifications` llega en la fase siguiente; `catalog`, `orders` y `payments`
> son servicios con su propia BD. `shared/` es código compartido del monorepo,
> **no** una base de datos compartida.

## Value objects

### `Money`

```go
type Money struct {
    cents    int64
    currency string // ISO 4217, "EUR" por defecto
}

func NewMoney(cents int64, currency string) (Money, error) // cents >= 0, currency != ""
func (m Money) Cents() int64
func (m Money) Add(other Money) (Money, error)             // exige misma moneda
func (m Money) Equals(other Money) bool
```

- **Comparable** por ser solo valores → `amount != order.total` funciona.
- Cero `float64`. El parseo desde HTTP ocurre en un único punto (`money.Parse`).
- El error de moneda distinta es de dominio, no de HTTP.

### IDs tipados

```go
type OrderID uuid.UUID
type ProductID uuid.UUID
type PaymentID uuid.UUID
```

Evitan el clásico "pasar un `string` con el ID equivocado". UUID v7
([ADR-0011](./adr/ADR-0011-identificadores-y-config.md)).

### `Quantity`

`uint32` > 0. Nace en el comando `AddItem`, nunca en el agregado directamente.

## Agregados e invariantes

### `Order`

```go
type Order struct {
    id        OrderID
    userID    UserID
    items     []OrderItem
    total     Money
    status    OrderStatus   // Created | Paid | Cancelled
    version   int           // para ETag / If-Match
    events    []DomainEvent
}
```

**Invariantes:**

1. `total == Σ(item.price × item.qty)` — se calcula en el constructor, nunca se
   recibe del exterior (FR-05).
2. `Pay(amount Money)` falla con `ErrAmountMismatch` si `amount != total`.
   **Aquí muere H-PAGO-01**: no hay `if` en un service que pueda olvidarse.
3. Solo `Created` puede pagarse o cancelarse. `Paid`/`Cancelled` → `ErrInvalidState`.
4. `version` se incrementa en cada mutación; alimenta el `ETag`.

**Eventos que emite:** `OrderPlaced`, `OrderPaid`, `OrderCancelled`.

### `Payment`

```go
type Payment struct {
    id        PaymentID
    orderID   OrderID
    amount    Money      // igual al total por construcción del comando
    status    PaymentStatus // Processed | Rejected
    reason    string
    attempt   int
    idemKey   IdempotencyKey
}
```

**Invariantes:**

1. `amount` nunca se `$set` directamente: el constructor recibe el `total` del
   `Order`, no un `request.monto`.
2. Todo intento queda persistido (auditoría), aprobado o rechazado (FR-07).
3. Un mismo `idemKey` produce un solo `Payment` (FR-11).

### `Product` / `Stock`

```go
type Product struct {
    id      ProductID
    price   Money
    stock   uint32
    version int
}
```

`Reserve(qty)` y `Release(qty)` son métodos del agregado: `Reserve` falla si no
hay stock suficiente. La compensación `Release` es idempotente por
`eventID` (FR-10).

### `Cart`

Agregado pequeño con `version` para `If-Match`. No emite eventos de dominio; es
transient until checkout.

## Eventos de dominio

```go
type DomainEvent interface {
    EventID() string      // UUID v7, dedup
    AggregateID() string
    OccurredAt() time.Time
    TraceID() string      // propagado desde el contexto HTTP
    SchemaVersion() int   // empieza en 1
    Type() string
}
```

Convención de nombres en inglés, pasado: `order.placed`, `payment.processed`,
`stock.released`. El payload es JSON versionado (`schemaVersion: 1`). Catálogo
completo en [eventos-y-comandos.md](./eventos-y-comandos.md).

**Importante:** cada contexto se despliega como un servicio distinto
([ADR-0002](./adr/ADR-0002-microservicios-por-contextos.md)), así que **todos** los
eventos salen por Kafka: el productor los escribe en su **outbox** en la misma
transacción que el cambio de agregado, el relay los publica y el consumidor los
deduplica en su propia BD ([ADR-0007](./adr/ADR-0007-outbox-relay-propio.md)).
Todos comparten el mismo `DomainEvent`.

## Outbox

Tabla `outbox_events`:

| Columna | Tipo | Uso |
|---|---|---|
| `event_id` | UUID PK | dedup y idempotencia del relay |
| `aggregate_id` | TEXT | trazabilidad |
| `type` | TEXT | despacho por tipo |
| `payload` | JSONB | evento serializado |
| `occurred_at` | TIMESTAMPTZ | orden |
| `trace_id` | TEXT | correlación con OTel |
| `published_at` | TIMESTAMPTZ NULL | `NULL` = pendiente |

El relay hace `SELECT ... FOR UPDATE SKIP LOCKED LIMIT N`, publica al bus y
marca `published_at` **solo tras el ack** (equivalente al ADR-0019 del repo Java).
Los reintentos son seguros porque el `event_id` es estable.

## Saga `PlaceOrder` (process manager)

No es un flujo implícito de mensajes: es una **máquina de estados persistida**.

```
StockReserved ──► PaymentAuthorized ──► Confirmed
      │                  │
      └── ReleaseStock   └── VoidPayment   (compensaciones)
```

```go
type PlaceOrderSaga struct {
    id      SagaID
    state   State
    orderID OrderID
}

func (s *PlaceOrderSaga) On(e DomainEvent) []Command { ... }
```

- Cada `Command` se ejecuta una vez; su resultado es un evento que avanza la saga.
- El estado se persiste **en la misma transacción** que el evento que lo disparó.
- Tras un crash, la saga se reanuda desde su último estado.
- `ReleaseStock` y `VoidPayment` son compensaciones **declaradas**, no código
  suelto en un listener.
- El TTL de la reserva es una **red de seguridad**, no un sustituto de la saga.

Tests obligatorios del hito 3: fallo en el paso 2 → se ejecuta `ReleaseStock`
exactamente una vez, incluso si el evento se reentrega.

## Modelo de errores de dominio → HTTP

| Dominio | HTTP (`problem+json`) |
|---|---|
| `ErrInvalidInput` (forma del JSON) | `400` |
| `ErrAmountMismatch` (`amount != total`) | `422` |
| `ErrInvalidState` (pedido ya pagado) | `409` |
| `ErrStaleVersion` (`If-Match` obsoleto) | `412` |
| `ErrNotFound` | `404` |
| `ErrDuplicateIdemKey` con payload distinto | `409` |
| `ErrDuplicateIdemKey` con mismo payload | `200/201` (replay) |

Decisiones en [ADR-0010](./adr/ADR-0010-errores-rfc9457.md) y
[ADR-0009](./adr/ADR-0009-idempotencia-y-concurrencia.md).

## Lo que deliberadamente NO modelamos (todavía)

- Descuentos, cupones, impuestos, devoluciones, multimoneda real con FX.
- `Money` ya tiene `currency`, pero solo se soporta una por sesión de pedido.
- Reparto, inventario por almacén, pagos parciales.
- Membership changes ni log compaction (no hay consenso aquí).
