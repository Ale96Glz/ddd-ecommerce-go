# ADR-0009: `Idempotency-Key` + `ETag`/`If-Match`

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** FR-03, FR-11

## Contexto y Problema

Un cliente que reintenta `POST /v1/orders` por un timeout no sabe si la orden se
creó: repetirla descontaría stock dos veces. Y dos pestañas editando el carrito
pueden sobrescribirse sin querer. El repo Java tenía idempotencia en el consumidor
(`restock_event`) pero no en el borde HTTP.

## Opciones Evaluadas

1. **`Idempotency-Key` persistida + `ETag`/`If-Match` (elegida)**.
   - Pros: semántica HTTP estándar (estilo Stripe) para reintentos; la respuesta
     original se replaya **incluido el status**; el ETag evita writes perdidos.
   - Contras: una tabla más y cabeceras que el cliente debe enviar.
2. **Solo dedup en el consumidor**.
   - Pros: menos código.
   - Contras: no protege el POST inicial; el cliente seguiría creando órdenes
     duplicadas.
3. **Bloqueo pesimista (`SELECT FOR UPDATE`) en todas las mutaciones**.
   - Pros: simple de razonar.
   - Contras: serializa todo y no arregla el reintento del cliente; descartado.

## Decisión

1. Todo POST que mueve dinero o stock exige `Idempotency-Key`.
2. Se persiste `(key, request_hash, response_status, response_body, created_at)`.
   - Misma key + mismo hash → **replay** de la respuesta original (mismo status).
   - Misma key + hash distinto → `409 problem+json`.
   - El guardado y el efecto del comando ocurren en la **misma transacción**.
3. La clave se guarda también en el evento de dominio, para que el outbox y la
   saga hereden la idempotencia de extremo a extremo.
4. Los recursos mutables (`Cart`, `Product`, `Order`) exponen `ETag` derivado de
   su `version`. Las mutaciones exigen `If-Match`; desalineado → `412`.
5. Los consumidores de eventos usan `processed_event(event_id PK)` en la misma
   transacción del efecto ([ADR-0007](./ADR-0007-outbox-relay-propio.md)). La tabla
   vive en la BD del **consumidor**, nunca en la del productor.
6. Las claves y ETags son opacos para el cliente; se devuelven en la respuesta.

## Consecuencias

### Positivas

- Reintentos seguros de verdad: ni cobros ni stock duplicados.
- `412` da una señal clara al cliente de que debe re-leer y reintentar.
- La idempotencia atraviesa HTTP → dominio → outbox → consumidor.

### Negativas / Compromisos

- La tabla de idempotencia crece; se purgará por `created_at` en el futuro (fuera
  del MVP).
- El cliente **debe** enviar la cabecera; se documenta en el OpenAPI como
  `required`.
- Requiere tests explícitos de replay y de key reutilizada con payload distinto.
