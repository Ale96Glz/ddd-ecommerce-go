# Lenguaje ubicuo

> **Regla de oro:** el equipo habla y documenta en **español**; el código, el
> JSON, los eventos y la BD usan **inglés**. Esta tabla es el diccionario
> oficial que traduce entre ambos. Si una palabra no está aquí, no se usa.

El lenguaje ubicuo no es un glosario decorativo: es el contrato entre el dominio
y el código. Un término = un identificador = un concepto. Si el código dice
`importe` en un sitio y `amount` en otro, hay dos conceptos donde solo debía
haber uno, y las invariantes se pierden por el camino.

## Convención de traducción

| Nivel | Idioma | Ejemplo |
|---|---|---|
| Conversación, ADRs, docs, commits | Español | "El pedido se paga con el total calculado" |
| Identificadores Go, paquetes, tipos, métodos | Inglés | `Order`, `Pay`, `AmountMismatch` |
| Campos JSON / OpenAPI | Inglés (`camelCase`) | `amount`, `orderId` |
| Eventos de dominio | Inglés, pasado, `contexto.hecho` (hecho en `snake_case`) | `order.placed`, `stock.release_requested` |
| Comandos | Inglés, imperativo | `ReserveStock` |
| Columnas SQL | Inglés, `snake_case` | `amount_cents`, `order_id` |
| Códigos de error `problem+json` | Inglés, `kebab-case` | `amount-mismatch` |
| Nombres de tests | Inglés, terminología del glosario | `TestOrder_Pay_AmountMismatch` |

## Glosario por contexto

### Catálogo (`catalog`)

| Español | Código | Definición | No usar |
|---|---|---|---|
| Producto | `Product` | Artículo vendible, raíz de agregado; tiene precio y stock | Artículo, Item |
| Categoría | `Category` | Agrupación de productos | Tipo, Clasificación |
| Precio | `Price` | Valor monetario de **una unidad** del producto (= `Money`) | Coste, Tarifa, Importe |
| Stock | `Stock` | Unidades disponibles del producto | Inventario, Existencias |
| Reserva | `Reservation` | Unidades apartadas para un pedido en curso | Apartado, Hold |
| Reponer stock | `ReleaseStock` | Compensación que devuelve las unidades reservadas | Restock, Devolver (en código) |

### Carrito (`cart`)

| Español | Código | Definición | No usar |
|---|---|---|---|
| Carrito | `Cart` | Selección **temporal** de ítems antes del pedido | Basket, Cesta |
| Ítem de carrito | `CartItem` | Producto + cantidad elegida | Línea, Detalle |
| Versión | `Version` | Contador de cambios del agregado → `ETag` | Revision (en código) |

### Pedidos (`orders`)

| Español | Código | Definición | No usar |
|---|---|---|---|
| Pedido | `Order` | Intención de compra con ítems y total calculado; raíz de agregado | Compra, Orden |
| Ítem de pedido | `OrderItem` | Producto + cantidad + **precio congelado** al crear el pedido | Detalle, Línea |
| Total | `Total` | `Σ(precio × cantidad)` calculado por el servidor | Monto, Importe, Suma |
| Cliente | `Customer` | Quien realiza el pedido (rol en este contexto) | Usuario (aquí) |
| Confirmar pedido | `ConfirmOrder` | Paso final de la saga: pedido → `Paid` | Aprobar (en código) |
| Cancelar pedido | `CancelOrder` | Solo desde `Created`; dispara compensación | Anular |

### Pagos (`payments`)

| Español | Código | Definición | No usar |
|---|---|---|---|
| Pago | `Payment` | Intento de cobro **registrado**, aprobado o rechazado; raíz de agregado | Cobro, Transacción |
| Monto | `Amount` | Valor monetario del pago; **debe igualar el Total** | Importe, Money (como campo) |
| Aprobación | `Approval` | Resultado `Processed` del pago | Éxito, OK |
| Rechazo | `Rejection` | Resultado `Rejected` con motivo legible | Fallo, Error |
| Motivo | `Reason` | Explicación obligatoria de un rechazo | Mensaje, Descripción |
| Umbral de aprobación | `ApprovalLimit` | `Amount` máximo aprobado sin revisión | Límite, Tope, `monto-maximo-aprobado` |
| Intento | `Attempt` | Número de intento de pago del pedido | Retry, Reintento |

### Saga e infraestructura (`shared`)

| Español | Código | Definición | No usar |
|---|---|---|---|
| Saga | `Saga` | Proceso con pasos y compensaciones; raíz de agregado propia | Transacción distribuida |
| Paso | `Step` | Unidad de avance de la saga | Fase, Etapa |
| Compensación | `Compensation` | Acción que deshace un paso ya ejecutado | Rollback (no es una BD) |
| Evento de dominio | `DomainEvent` | Hecho ocurrido, inmutable, con `eventID` | Mensaje, Notificación |
| Comando | `Command` | Intención de cambiar estado; puede fallar | Request, Petición |
| Outbox | `Outbox` | Tabla durable de eventos pendientes de publicar | Buzón, Cola, Queue |
| Relay | `Relay` | Proceso que publica el outbox y marca `published_at` | Publisher, Publicador |
| Consumidor | `Consumer` | Suscriptor que aplica un evento con dedup | Listener, Handler |
| Idempotencia | `Idempotency` | Repetir una operación sin duplicar efectos | — |
| Clave de idempotencia | `IdempotencyKey` | Identificador opaco del reintento | Token, Correlation ID |
| Entrega | `Delivery` | Intento de envío de un webhook | Envío, Push |

### Notificaciones (`notifications`)

| Español | Código | Definición | No usar |
|---|---|---|---|
| Notificación | `Notification` | Mensaje al cliente sobre un resultado | Aviso, Alert |
| Webhook | `WebhookEndpoint` | Endpoint del cliente al que se entregan eventos | Callback, Listener |
| Firma | `Signature` | HMAC que autentica una entrega | Token, Hash |

## Distinciones que se confunden (y importan)

| Par | Diferencia |
|---|---|
| **Total** vs **Monto** | El total lo calcula el servidor desde el pedido; el monto es lo que se cobra. El pago exige `amount == total`. `Total` no se envía en el POST; `Amount` sí existía en el contrato Java y aquí se deriva. |
| **Precio** vs **Total** | Precio es por unidad (y se congela en el ítem); total es la suma. |
| **Stock** vs **Reserva** | Stock es lo disponible; reserva es lo apartado por un pedido en curso. |
| **Pedido** vs **Pago** | Pedido es la intención; pago es el intento de cobro. Un pedido `Paid` implica un pago `Processed`, no al revés. |
| **Usuario** vs **Cliente** | Misma persona física, distinto contexto: `User` en `auth` (credenciales), `Customer` en `orders` (comprador). No se unifica. |
| **Evento** vs **Comando** | El evento cuenta un hecho ya ocurrido (no puede fallar); el comando pide una acción (puede fallar). Nombres en pasado vs imperativo. |
| **Compensación** vs **Rollback** | Rollback deshace una transacción local; compensación deshace un paso ya confirmado en otro agregado. |
| **Outbox** vs **Cola** | Outbox es durabilidad dentro de la BD; la cola/bus es transporte. No son lo mismo. |
| **Rechazo** vs **Error** | Rechazo es una decisión legítima del pago (se persiste, dispara compensación); error es una petición inválida (`422`, sin persistir). |

## Palabras prohibidas en código

`monto`, `importe`, `orden`, `pedido`, `producto`, `stock`, `restock`, `manager`,
`helper`, `util`, `bo`, `vo`, `dto`, `entity`, `data`, `info`, `obj`.

- El código va **siempre** en inglés; `monto` solo aparece en documentación
  histórica del repo Java y en el `detail` de los mensajes al usuario.
- Sufijos `VO` / `Dto` / `Entity` prohibidos: `Money`, no `MoneyVO`;
  `Order`, no `OrderEntity`.
- `Service`, `Manager`, `Helper` prohibidos en `domain/`. En `application/` se
  permite `Service` para los casos de uso.
- Los nombres de tests usan el término del glosario:
  `TestOrder_Pay_AmountMismatch`, `TestMoney_Add_CurrencyMismatch`.

## Vocabulario de errores

| Código `problem+json` | HTTP | Cuándo se usa (término del glosario) |
|---|---|---|
| `invalid-request` | 400 | Forma del payload inválida |
| `unauthenticated` / `forbidden` | 401 / 403 | Sin JWT / sin permiso |
| `not-found` | 404 | Pedido, producto o pago inexistente |
| `conflict` | 409 | Pedido no `Created`, o clave de idempotencia reutilizada |
| `precondition-failed` | 412 | `If-Match` obsoleto (versión del carrito/producto) |
| `amount-mismatch` | 422 | El **monto** no coincide con el **total** |
| `currency-mismatch` | 422 | Monedas distintas dentro de un `Money` |

Detalle completo en [ADR-0010](./adr/ADR-0010-errores-rfc9457.md).

## Cómo se aplica

1. Todo ADR, requisito y escenario usa estos términos sin sinónimos.
2. Toda nueva struct, método, evento o columna se nombra desde la columna
   **Código** de este documento.
3. Si el dominio necesita un término nuevo: se añade aquí **en el mismo commit**
   que al código, con su definición y su "no usar".
4. Las descripciones del OpenAPI usan el término en español y el campo en inglés,
   p. ej. *"monto del pago (`amount`): debe coincidir con el total del pedido"*.
