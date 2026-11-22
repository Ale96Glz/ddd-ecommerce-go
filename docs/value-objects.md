# Catálogo de value objects

Complemento de [dominio.md](./dominio.md): ahí está la estructura; aquí están
**las invariantes exactas** de cada valor. Los nombres siguen el
[lenguaje ubicuo](./lenguaje-ubicuo.md).

## Reglas de oro

Un value object (VO) en este proyecto:

1. Es **inmutable**: sin setters; todo método devuelve un valor nuevo.
2. **Se autovalida en el constructor**. No existe una instancia inválida; si el
   constructor no devuelve error, el valor es bueno.
3. Tiene **igualdad por valor**, nunca por identidad. Si solo contiene valores
   comparables, el tipo es comparable y se usa `==`.
4. **No depende de infraestructura**: ni `net/http`, ni `pgx`, ni `chi`, ni
   reloj global. (Un ID se genera fuera y se pasa ya construido.)
5. Se llama como el concepto, **sin sufijo** `VO`, `Dto` ni `Entity`.
6. Su representación JSON y SQL está definida aquí, en un solo sitio.
7. Test mínimo por VO: **invariante rechazada + igualdad + round-trip**
   (JSON/SQL). Sin DB para el test de dominio.

## Catálogo de dominio

### `Money`

| | |
|---|---|
| **Campos** | `cents int64`, `currency string` (ISO 4217) |
| **Invariantes** | `cents >= 0`; `currency` de 3 letras mayúsculas soportada |
| **Error** | `ErrNegativeAmount`, `ErrInvalidCurrency` |
| **Constructor** | `NewMoney(cents int64, currency string) (Money, error)` |
| **Operaciones** | `Add`, `Subtract` → exigen misma moneda; `Equals`, `Compare`, `IsZero` |
| **Igualdad** | Por valor (`==`): `30.00 == 30.0 == 3000 centavos` |
| **JSON** | `{ "amount": 3000, "currency": "EUR" }` — **centavos**, nunca decimal |
| **SQL** | `amount_cents BIGINT`, `currency CHAR(3)` |
| **Uso** | `Product.price`, `Order.total`, `OrderItem.unitPrice`, `Payment.amount`, `ApprovalLimit` |
| **No confundir con** | `Price` (es el mismo tipo con otro nombre en el código; no se crea un tipo aparte) |

> El parseo de texto humano (`"30.00"`) ocurre **solo** en el adaptador HTTP
> (`money.Parse`), nunca en el dominio. Ver [ADR-0005](./adr/ADR-0005-money-value-object.md).

### `Currency`

| | |
|---|---|
| **Campos** | `code string` |
| **Invariantes** | Exactamente 3 letras mayúsculas; incluida en la lista soportada (EUR por ahora) |
| **Error** | `ErrInvalidCurrency`, `ErrUnsupportedCurrency` |
| **Constructor** | `NewCurrency(code string) (Currency, error)` |
| **Igualdad** | Por código |
| **JSON** | String: `"EUR"` |
| **SQL** | `CHAR(3)` |
| **Uso** | Dentro de `Money`; no se serializa por separado |
| **No confundir con** | Tipo de cambio (no modelado) |

### `Quantity`

| | |
|---|---|
| **Campos** | `value uint32` |
| **Invariantes** | `value > 0`; límite superior `1_000_000` para evitar desbordes al reservar |
| **Error** | `ErrInvalidQuantity` |
| **Constructor** | `NewQuantity(n uint32) (Quantity, error)` |
| **Operaciones** | `Add`, `Subtract` → `ErrInsufficientStock` si el resultado sería 0 |
| **Igualdad** | Por valor |
| **JSON** | Entero: `3` |
| **SQL** | `INTEGER` |
| **Uso** | `CartItem.quantity`, `OrderItem.quantity`, comandos `ReserveStock`/`ReleaseStock` |
| **No confundir con** | `Stock` (disponibilidad), que es estado del agregado `Product` |

### IDs tipados (`OrderID`, `ProductID`, `PaymentID`, `SagaID`, `CustomerID`)

| | |
|---|---|
| **Campos** | UUID v7 envuelto (`type OrderID uuid.UUID`) |
| **Invariantes** | UUID válido **y versión 7** |
| **Error** | `ErrInvalidID` |
| **Constructor** | `ParseOrderID(s string) (OrderID, error)`, `NewOrderID() OrderID` |
| **Igualdad** | Por valor (UUID es comparable) |
| **JSON** | String canónico: `"0192c1a4-..."` |
| **SQL** | `UUID` (o `TEXT` con el mismo formato en SQLite de test) |
| **Uso** | Claves de agregados, `aggregate_id` del outbox, rutas `/v1/orders/{id}` |
| **No confundir con** | `IdempotencyKey` (opaco, no es identidad de recurso) |

Cada ID es un **tipo distinto**: pasar un `ProductID` donde se espera un
`OrderID` no compila. Ver [ADR-0011](./adr/ADR-0011-identificadores-y-config.md).

### `IdempotencyKey`

| | |
|---|---|
| **Campos** | `value string` |
| **Invariantes** | No vacío; 1–255 caracteres; caracteres visibles |
| **Error** | `ErrInvalidIdempotencyKey` |
| **Constructor** | `NewIdempotencyKey(s string) (IdempotencyKey, error)` |
| **Igualdad** | Por valor exacto (case-sensitive) |
| **JSON** | Header `Idempotency-Key` (no va en el body) |
| **SQL** | `VARCHAR(255)` con índice único `(key, request_hash)` |
| **Uso** | Todo POST que mueve dinero o stock; viaja al evento de dominio |
| **No confundir con** | `TraceID` (diagnóstico, no control de duplicados) |

### `Version`

| | |
|---|---|
| **Campos** | `value uint64` |
| **Invariantes** | `value >= 1`; solo aumenta, nunca se decrementa a mano |
| **Error** | `ErrStaleVersion` al comparar con un `If-Match` desactualizado |
| **Constructor** | `NewVersion(n uint64) (Version, error)`, `Next() Version` |
| **Igualdad** | Por valor |
| **JSON** | Header `ETag`: `"W/\"7\""` |
| **SQL** | `BIGINT` |
| **Uso** | `Cart`, `Product`, `Order` → concurrencia optimista |
| **No confundir con** | `Attempt` (intento de pago, distinto contador) |

### `Reason`

| | |
|---|---|
| **Campos** | `text string` |
| **Invariantes** | No vacío; máximo 500 caracteres |
| **Error** | `ErrInvalidReason` |
| **Constructor** | `NewReason(s string) (Reason, error)` |
| **Igualdad** | Por valor |
| **JSON** | String: `"Monto excede el máximo aprobado (100.00)"` |
| **SQL** | `VARCHAR(500)` |
| **Uso** | `Payment.reason` cuando el resultado es `Rejected`; `null` si `Processed` |
| **No confundir con** | `AppError.Message` (errores de petición, no decisiones de pago) |

### `Email`

| | |
|---|---|
| **Campos** | `address string` |
| **Invariantes** | Formato válido según `net/mail.ParseAddress`; longitud ≤ 254 |
| **Error** | `ErrInvalidEmail` |
| **Constructor** | `NewEmail(s string) (Email, error)` |
| **Igualdad** | Por valor en minúsculas (el constructor normaliza) |
| **JSON** | String |
| **SQL** | `CITEXT` o `VARCHAR(254)` con `LOWER()` en el índice único |
| **Uso** | Contexto `auth` / `notifications`, nunca en `orders` (ahí se usa `CustomerID`) |
| **No confundir con** | Identidad de usuario |

## VOs de plataforma (no de dominio)

Viven en `internal/platform/`, no en `shared/`. No expresan negocio, pero también
son inmutables y por valor:

| Tipo | Invariante | Uso |
|---|---|---|
| `TraceID` | 32 hex (W3C trace context) | `DomainEvent.traceID`, logs, `outbox_events.trace_id` |
| `RequestID` | UUID v4 | Header de respuesta, logs de la petición |

## Lo que deliberadamente NO es un VO (todavía)

| Concepto | Por qué no |
|---|---|
| `Price` | Es `Money`; un tipo aparte no añade invariantes |
| `Percentage`, `Discount` | No hay descuentos en el MVP |
| `Address`, `Sku`, `DateRange` | No existen en el dominio modelado |
| `MoneyVO`, `OrderDto`, `PaymentEntity` | Prohibidos por la convención de nombres |
| `Stock` | Es **estado del agregado** `Product`, no un valor suelto |

## Checklist antes de añadir un VO

- [ ] El término está en [lenguaje-ubicuo.md](./lenguaje-ubicuo.md).
- [ ] Tiene al menos una invariante que el tipo pueda hacer cumplir.
- [ ] No depende de BD, HTTP ni framework.
- [ ] Definida su representación JSON y SQL.
- [ ] Test de invariante + igualdad + round-trip.
- [ ] ¿Realmente no es un agregado ni estado de uno? (Si tiene identidad o
      ciclo de vida, no es un VO.)
