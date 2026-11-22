# ADR-0015: Lenguaje ubicuo bilingüe y nomenclatura única

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** NFR-02, NFR-05

## Contexto y Problema

En DDD, el vocabulario del dominio debe aparecer **idéntico** en la conversación,
los documentos y el código. El repo Java usaba español en los identificadores
(`pedido`, `monto`, `reponerStock`) y su JSON exponía `monto`. Eso es coherente
con sus hablantes, pero genera dos problemas aquí:

1. Mezcla con un ecosistema Go y un contrato HTTP que son ingleses por defecto
   (`Idempotency-Key`, `problem+json`, `trace_id`).
2. Permitió que `monto` significara a la vez "importe recibido" y "total del
   pedido"; el hallazgo H-PAGO-01 nació de esa ambigüedad.

Hay que decidir el idioma del código **antes** de escribir el primer tipo, y
fijar que un concepto tenga un solo nombre.

## Opciones Evaluadas

1. **Español en documentación, inglés en código, con diccionario explícito
   (elegida)**.
   - Pros: el dominio se discute en español con el usuario; el código encaja con
     Go, OpenAPI y JSON; el glosario ES→EN elimina sinónimos accidentales.
   - Contras: hay que consultar el glosario al principio; rompe el JSON anterior
     (`monto` → `amount`, `pedido` → `order`).
2. **Español también en los identificadores** (como el repo Java).
   - Pros: cero traducción mental para el equipo.
   - Contras: nombres híbridos (`PedidoID`, `montoEUR`), fricción con librerías y
     con el contrato HTTP; no resuelve la ambigüedad `monto`/`total`.
3. **Mezcla por contexto** (catálogo en español, pagos en inglés).
   - Pros: flexibilidad local.
   - Contras: el mismo concepto se llama distinto según dónde esté; descartada.

## Decisión

1. Se publica [`docs/lenguaje-ubicuo.md`](../lenguaje-ubicuo.md) como
   **diccionario oficial**: término en español → identificador en inglés →
   definición → "no usar".
2. Un concepto = **un identificador**. Quedan prohibidos los sinónimos en código:
   `amount`/`importe`, `order`/`pedido`, `release`/`restock` mezclados.
3. Reglas de nombre:
   - Tipos y métodos en inglés, sin sufijos `VO`, `Dto`, `Entity`.
   - Eventos: `<contexto>.<hecho>` en pasado, hecho en `snake_case`.
   - Comandos: imperativo (`ReserveStock`).
   - Columnas SQL: `snake_case` inglés; campos JSON: `camelCase` inglés.
   - Tests: `Test<Aggregate>_<Method>_<Condition>` con términos del glosario.
4. El vocabulario de errores (`amount-mismatch`, `currency-mismatch`, etc.) es
   el mismo en el código, el OpenAPI y los ADRs
   ([ADR-0010](./ADR-0010-errores-rfc9457.md)).
5. **Todo nuevo término se añade al glosario en el mismo commit** que al código.
   Un cambio de contrato que renombre un campo actualiza glosario + OpenAPI +
   escenarios a la vez.

## Consecuencias

### Positivas

- Un reviewer puede leer `internal/orders/domain` y entenderlo sin conocer el
  repo Java.
- `Amount` y `Total` son conceptos distintos con nombres distintos: el bug
  H-PAGO-01 deja de poder expresarse.
- Los mensajes de error, los eventos y los tests usan la misma palabra; no hay
  que traducir durante la revisión.

### Negativas / Compromisos

- El contrato HTTP cambia frente al repo Java (`monto` → `amount`,
  `pedidoId` → `orderId`). Es un corte deliberado, documentado en
  [escenarios.md](../escenarios.md).
- Los mensajes de `detail` dirigidos al usuario siguen en español; solo los
  identificadores técnicos son ingleses. Se testea que no se filtre un término
  prohibido a las respuestas.
- Requiere disciplina: el glosario se actualiza en el mismo commit o se convierte
  en documentación muerta.
