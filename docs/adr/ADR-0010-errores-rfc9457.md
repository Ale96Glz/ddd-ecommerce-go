# ADR-0010: `AppError` + RFC 9457 `problem+json`; taxonomía HTTP

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** FR-07, NFR-05

## Contexto y Problema

En Go no hay excepciones: `if err != nil` se reparte por todos los handlers y es
fácil que cada uno decida un status distinto para el mismo error de dominio. El
repo Java usaba un `{message, error, status}` artesanal y tuvo que añadir un
código nuevo (`422`) para H-PAGO-01. Hay que fijar el mapeo **antes** del primer
handler.

## Opciones Evaluadas

1. **`AppError{Code, Message}` + middleware → `problem+json` (elegida)**.
   - Pros: un solo punto de mapeo; el dominio no conoce HTTP; formato estándar
     RFC 9457 (`type`, `title`, `status`, `detail`, `instance`) con `type` estable
     para que los clientes ramifiquen por URL, no por texto.
   - Contras: define un tipo propio (son ~20 líneas).
2. **Mapear errores directamente en cada handler**.
   - Pros: cero abstracción.
   - Contras: inconsistencia garantizada al crecer.
3. **Librería de problems**.
   - Pros: gestión de cabeceras y serialización.
   - Contras: dependencia innecesaria para un formato pequeño.

## Decisión

1. Errores de dominio tipados (`ErrAmountMismatch`, `ErrInvalidState`,
   `ErrCurrencyMismatch`, `ErrNotFound`, `ErrStaleVersion`, ...).
2. Un `AppError{Code string, Message string, Cause error}` en
   `internal/platform/errors`. La capa de aplicación traduce errores de dominio
   a `AppError` (o los pasa con un `Code` por defecto).
3. Middleware de `chi` que intercepta `AppError` y escribe
   `application/problem+json` con `type = https://errors.ecommerce.test/<code>`.
4. Taxonomía:

   | Situación | Status | Código |
   |---|---|---|
   | Forma del payload / validación HTTP | `400` | `invalid-request` |
   | Sin auth / sin permiso | `401` / `403` | `unauthenticated` / `forbidden` |
   | Recurso inexistente | `404` | `not-found` |
   | Conflicto de estado o idem-key reutilizada | `409` | `conflict` |
   | `If-Match` obsoleto | `412` | `precondition-failed` |
   | Regla de negocio del contenido (`amount ≠ total`) | `422` | `amount-mismatch` |

5. `validator/v10` se usa **solo en el borde HTTP**; las invariantes reales están
   en los value objects ([ADR-0005](./ADR-0005-money-value-object.md)). No se
   duplica la validación en dos sitios.
6. El `detail` nombra los valores implicados, p. ej.
   `"El monto 91.00 no coincide con el total del pedido 90.00"`.

## Consecuencias

### Positivas

- Un único sitio decide status; añadir un error nuevo es una línea.
- `type` es estable: los clientes no parsean mensajes en español.
- El dominio sigue siendo ignorante de HTTP y de JSON.

### Negativas / Compromisos

- Hay que documentar cada `code` en el OpenAPI para que el cliente conozca los
  `type` posibles.
- El `422` se mantiene por claridad contractual, aunque el modelo ya hace
  imposible el caso en el camino interno; es defensa en profundidad.
