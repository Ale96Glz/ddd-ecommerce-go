# Decisiones de arquitectura (ADR)

Registro de decisiones de arquitectura del proyecto `ecommerce-go`. Cada ADR
sigue el formato Michael Nygard (**Contexto → Decisión → Consecuencias**) e
incluye las opciones evaluadas y la alternativa descartada.

Antes de escribir código, una decisión corregida se reescribe en su propio ADR
(es lo que pasó con ADR-0002 y ADR-0007). A partir del hito 1 los ADRs son
**inmutables**: un cambio se registra con un ADR nuevo que supera al anterior
(se marca el viejo como *Superseded*).

| ID | Título | Estatus | Requisitos |
|---|---|---|---|
| [ADR-0001](./ADR-0001-lenguaje-y-estructura.md) | Go estable, un solo `go.mod`, `cmd/` + `internal/` | Aceptado | NFR-10 |
| [ADR-0002](./ADR-0002-microservicios-por-contextos.md) | Microservicios por bounded context, BD por servicio | Aceptado | NFR-11 |
| [ADR-0003](./ADR-0003-router-http.md) | `chi/v5` como router HTTP | Aceptado | NFR-03 |
| [ADR-0004](./ADR-0004-contrato-openapi.md) | OpenAPI-first con `oapi-codegen` | Aceptado | NFR-03 |
| [ADR-0005](./ADR-0005-money-value-object.md) | `Money` como value object (`int64` centavos + ISO 4217) | Aceptado | FR-05, FR-06, NFR-01 |
| [ADR-0006](./ADR-0006-pgx-sqlc-goose.md) | `pgx/v5` + `sqlc` + `goose`, sin ORM | Aceptado | NFR-04 |
| [ADR-0007](./ADR-0007-outbox-relay-propio.md) | Outbox con relay propio y Kafka | Aceptado | FR-09, FR-10, NFR-11 |
| [ADR-0008](./ADR-0008-saga-process-manager.md) | Saga `PlaceOrder` como process manager persistido | Aceptado | FR-08 |
| [ADR-0009](./ADR-0009-idempotencia-y-concurrencia.md) | `Idempotency-Key` + `ETag`/`If-Match` | Aceptado | FR-03, FR-11 |
| [ADR-0010](./ADR-0010-errores-rfc9457.md) | `AppError` + RFC 9457 `problem+json`; taxonomía HTTP | Aceptado | FR-07, NFR-05 |
| [ADR-0011](./ADR-0011-identificadores-y-config.md) | UUID v7, JWT `jwt/v5` + argon2id, config por entorno | Aceptado | FR-01, NFR-07 |
| [ADR-0012](./ADR-0012-observabilidad.md) | `slog` + OpenTelemetry → OTLP, `trace_id` en el outbox | Aceptado | NFR-06 |
| [ADR-0013](./ADR-0013-estrategia-de-tests.md) | `testing` + `httptest` + `testcontainers-go` + `go-cmp` | Aceptado | NFR-08 |
| [ADR-0014](./ADR-0014-build-y-deploy.md) | Multi-stage → distroless + `docker compose` | Aceptado | NFR-09, NFR-10 |
| [ADR-0015](./ADR-0015-lenguaje-ubicuo-y-nomenclatura.md) | Lenguaje ubicuo bilingüe y nomenclatura única | Aceptado | NFR-02, NFR-05 |

## Guías relacionadas

| Guía | Contenido |
|---|---|
| [Visión](../vision.md) | Objetivos y definición de éxito |
| [Requisitos](../requisitos.md) | FR/NFR y trazabilidad hacia los ADRs |
| [Lenguaje ubicuo](../lenguaje-ubicuo.md) | Vocabulario oficial y reglas de nombres |
| [Dominio](../dominio.md) | Bounded contexts, agregados, VOs, eventos, saga |
| [Value objects](../value-objects.md) | Invariantes, representación JSON/SQL y tests de cada VO |
| [Eventos y comandos](../eventos-y-comandos.md) | Mensajes, productores, consumidores e idempotencia |
| [Stack](../stack.md) | Tabla única de decisiones técnicas |
| [Escenarios](../escenarios.md) | Comportamiento HTTP esperado |
| [Roadmap](../roadmap.md) | Hitos y tests de aceptación |

## Convención de numeración

Los ADR-00xx son decisiones fundacionales (stack y diseño). Se numeran
secuencialmente y no se reutilizan IDs. El estado puede ser: `Propuesto`,
`Aceptado`, `Superseded`, `Rechazado`.
