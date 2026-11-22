# ADR-0008: Saga `PlaceOrder` como process manager persistido

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** FR-08

## Contexto y Problema

El repo Java implementó la compensación de stock con un listener que reaccionaba
a `payment-processed` y encolaba `RESTOCK_REQUIRED` (ADR-0013). Funciona, pero el
flujo completo está repartido entre tres servicios y no hay un objeto que
represente "la operación de comprar". Es difícil responder: *¿en qué paso está
esta compra y qué compensación le toca?*

## Opciones Evaluadas

1. **Process manager explícito y persistido (elegida)**.
   - Pros: la máquina de estados es un tipo con tests puros; cada transición y su
     compensación están declaradas juntas; sobrevive a un crash y es reanudable.
   - Contras: una tabla más y algo más de código que un listener.
2. **Coreografía por eventos** (cada servicio reacciona).
   - Pros: menos código central.
   - Contras: el flujo es implícito; con 3+ pasos se vuelve imposible de razonar
     y de testear sin integración.
3. **Transacción distribuida / 2PC**.
   - Pros: atomicidad.
   - Contras: no existe entre procesos de forma práctica; descartada.

## Decisión

1. `PlaceOrderSaga` es una **raíz de agregado propia** en `internal/orders`, con
   repositorio e ID propios, distinta de `Order`.
2. Estados: `StockReserved → PaymentAuthorized → Confirmed`, más terminales
   `Completed` y `Compensated`.
3. API: `func (s *PlaceOrderSaga) On(e DomainEvent) []Command`.
   - Cada `Command` se ejecuta **una vez** y su resultado es un evento que
     avanza la saga.
   - Compensaciones declaradas: `ReleaseStock`, `VoidPayment`.
4. El cambio de estado de la saga **y** el evento resultante se escriben en la
   misma transacción ([ADR-0007](./ADR-0007-outbox-relay-propio.md)).
5. Los comandos son idempotentes por `sagaID + step`; los efectos por
   `event_id` ([ADR-0009](./ADR-0009-idempotencia-y-concurrencia.md)).
   `ReserveStock` y `ReleaseStock` viajan por Kafka a `catalog-service`;
   `AuthorizePayment` y `VoidPayment` a `payments-service`.
6. El TTL de la reserva de stock es una **red de seguridad**, no el mecanismo de
   compensación: la saga sigue siendo la fuente de la decisión.
7. Test obligatorio: fallo en `AuthorizePayment` → `ReleaseStock` exactamente
   una vez; reanudar tras crash no duplica la compensación.

## Consecuencias

### Positivas

- Se puede explicar la compra con un dibujo de 4 estados y sus compensaciones.
- Los tests de la saga son unitarios y deterministas (inyectando eventos).
- Reanudable: responde a "¿qué le pasó al pedido X?" con el estado persistido.

### Negativas / Compromisos

- Más código que la coreografía; se justifica porque el dominio tiene dinero y
  stock de por medio.
- La saga no es atómica con los servicios externos: dependemos de comandos
  idempotentes para que el reintento sea seguro.
- Sin un dashboard, inspeccionar sagas requiere consultar su tabla (aceptable en
  el MVP; SSE del pedido da la vista al usuario).
