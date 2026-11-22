# Catálogo de eventos y comandos

Complemento de [dominio.md](./dominio.md). Fija el nombre, el productor, el
consumidor y la clave de idempotencia de cada mensaje. Los nombres salen del
[lenguaje ubicuo](./lenguaje-ubicuo.md).

## Convenciones

**Evento de dominio** (`DomainEvent`): hecho ya ocurrido, inmutable, no puede
fallar. Nombre `<contexto>.<hecho>` en pasado, con el hecho en `snake_case` cuando
tiene varias palabras.

```text
order.placed                 stock.release_requested
payment.processed            webhook.delivery_failed
```

Todo evento lleva esta envelope (ver [dominio.md](./dominio.md)):

| Campo | Tipo | Uso |
|---|---|---|
| `eventID` | UUID v7 | **Estable**: dedup del consumidor y del relay |
| `aggregateID` | string | Trazabilidad del agregado origen |
| `occurredAt` | timestamp | Ordenación |
| `traceID` | W3C trace | Correlación HTTP → saga → compensación |
| `schemaVersion` | entero | Evolución del payload (empieza en 1) |

**Comando** (`Command`): intención de cambiar estado, puede fallar. Nombre en
imperativo. Se ejecuta **una vez por paso de saga** y su resultado es un evento.

## Comandos

| Comando | Contexto | Emitido por | Efecto | Idempotencia |
|---|---|---|---|---|
| `AddItemToCart` | `cart` | HTTP | Añade un ítem; exige `If-Match` | `IdempotencyKey` |
| `PlaceOrder` | `orders` | HTTP | Crea `Order`, reserva stock, escribe `order.placed` | `IdempotencyKey` |
| `CancelOrder` | `orders` | HTTP | `Created → Cancelled`, pide `ReleaseStock` | `IdempotencyKey` |
| `ReserveStock` | `catalog` | Saga, paso 1 | Descuenta stock del producto | `sagaID + step` |
| `AuthorizePayment` | `payments` | Saga, paso 2 | Crea `Payment` con `amount = order.total` | `sagaID + step` |
| `ConfirmOrder` | `orders` | Saga, paso 3 | `Created → Paid` | `sagaID + step` |
| `ReleaseStock` | `catalog` | Compensación | Repone el stock reservado | `eventID` |
| `VoidPayment` | `payments` | Compensación | Anula la autorización | `eventID` |

> `AuthorizePayment` **no recibe un monto del llamador**: recibe el `orderID` y
> toma `order.total`. Es el punto donde desaparece H-PAGO-01
> ([ADR-0005](./adr/ADR-0005-money-value-object.md)).

## Eventos

| Evento | Productor | Consumidor | Payload de dominio | Efecto |
|---|---|---|---|---|
| `order.placed` | `orders` | `payments`, saga | `orderID`, `customerID`, `total`, `items[]` | Arranca la saga |
| `stock.reserved` | `catalog` | saga | `orderID`, `productID`, `qty` | Avanza a `PaymentAuthorized` |
| `payment.processed` | `payments` | `orders`, `notifications` | `paymentID`, `orderID`, `amount`, `attempt` | Confirma el pedido |
| `payment.rejected` | `payments` | saga, `notifications` | `paymentID`, `orderID`, `reason` | Dispara compensación |
| `order.paid` | `orders` | `notifications`, SSE | `orderID`, `paymentID`, `total` | Notifica al cliente |
| `order.cancelled` | `orders` | `notifications`, SSE | `orderID`, `reason` | Notifica al cliente |
| `stock.release_requested` | saga | `catalog` | `orderID`, `productID`, `qty` | Pide la compensación |
| `stock.released` | `catalog` | saga | `orderID`, `productID`, `qty` | Cierra la compensación |
| `payment.void_requested` | saga | `payments` | `paymentID`, `orderID` | Anula autorización |
| `payment.voided` | `payments` | saga | `paymentID`, `orderID` | Cierra la compensación |
| `notification.created` | `notifications` | SSE | `customerID`, `message` | Entrega al usuario |
| `webhook.delivery_failed` | `notifications` | — | `endpointID`, `eventID`, `attempt` | Reintento con backoff |

## Transporte Kafka

Un topic por tipo de evento, un consumer group por servicio y `aggregateID` como
clave de partición (orden garantizado por agregado). Headers por mensaje:
`eventID`, `eventType`, `traceparent` (W3C) y `schemaVersion`. Tras 3 reintentos
el evento va a `<topic>.dlt`. Kafka corre en modo KRaft. Detalle en
[ADR-0007](./adr/ADR-0007-outbox-relay-propio.md).

## Reglas de publicación

1. El evento se escribe en `outbox_events` **en la misma transacción** que el
   cambio de agregado o de estado de la saga; el relay lo publica a Kafka y fija
   `published_at` solo con el ack del broker
   ([ADR-0007](./adr/ADR-0007-outbox-relay-propio.md)).
2. `eventID` se genera **una sola vez** y viaja con el evento; reintentos y
   reentregas reutilizan el mismo ID.
3. El consumidor registra `eventID` en `processed_event` **en la misma
   transacción** que su efecto. Reentrega ⇒ cero efectos nuevos.
4. Ningún consumidor publica un evento de dominio directamente: lo escribe en su
   outbox. Así la traza no se rompe.
5. Los eventos `*.requested` describen una **intención duradera**, no una
   confirmación. Su contrapartida `*.released` / `*.voided` cierra el ciclo.
6. Los payloads son JSON con `schemaVersion`. Un cambio incompatible obliga a un
   nuevo `schemaVersion` y a un ADR; nunca se reescribe un evento ya publicado.

## Relación con la saga

```
order.placed ──► ReserveStock ──► stock.reserved ──► AuthorizePayment
                                                        │
                                    ┌───────────────────┴────────────────┐
                              payment.processed                    payment.rejected
                                    │                                    │
                              ConfirmOrder                        ReleaseStock
                                    │                                    │
                                order.paid                         stock.released
                                                                  (+ order.cancelled)
```

Cada flecha es un evento; cada verbo es un comando. Detalle de estados y
compensaciones en [ADR-0008](./adr/ADR-0008-saga-process-manager.md).

## Lo que NO se modela como evento (todavía)

- `cart.item_added` / `cart.cleared`: el carrito no emite eventos de dominio; es
  estado temporal hasta `PlaceOrder`.
- Eventos de auditoría de login/logout: son logs de seguridad, no eventos de
  negocio.
- Eventos de métricas o de health check: son telemetría
  ([ADR-0012](./adr/ADR-0012-observabilidad.md)).
