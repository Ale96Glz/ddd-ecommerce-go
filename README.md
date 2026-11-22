# ecommerce-go — DDD táctico en Go moderno

Proyecto de **aprendizaje y portafolio**: un backend de e-commerce escrito en Go,
usando DDD táctico (value objects, agregados, eventos de dominio, repositorios),
**outbox** y **sagas** como process manager explícito.

No es un port ni una traducción del repo Java `ecommerce-microservices`. Es el
mismo dominio, rediseñado como se haría hoy: menos servicios, contratos mejores y
**corrección por diseño** (el bug `monto ≠ total` deja de ser posible por tipos).

## Estado

| Fase | Estado |
|---|---|
| 0 — Documentación (visión, requisitos, ADRs, dominio) | ✅ En curso |
| 1 — Núcleo de dominio puro (`Money`, `Order`, eventos) | ⬜ |
| 2 — Persistencia + outbox (`pgx`/`sqlc`/`goose`) | ⬜ |
| 3 — Saga `PlaceOrder` con compensaciones | ⬜ |
| 4 — Capa HTTP (OpenAPI + `chi`) | ⬜ |
| 5 — Worker, SSE y webhooks | ⬜ |

Sin código todavía. Esta carpeta contiene solo documentación.

## Principios rectores

1. **Las invariantes viven en el tipo, no en un `if` de un service.** `Money` no
   se puede construir con centavos negativos; `Order.Pay(amount)` exige que
   `amount == total`.
2. **Contrato primero.** El `openapi.yaml` manda; el código se genera.
3. **Sin ORM.** SQL tipado con `sqlc`; el agregado controla sus propias escrituras.
4. **Microservicios por bounded context**, cada uno con su BD y su outbox. Tres en el MVP.
5. **Errores como valor.** Un `AppError` → `application/problem+json` (RFC 9457)
   en un único punto del borde HTTP.
6. **Todo POST de dinero es idempotente** por `Idempotency-Key`.
7. **La documentación es parte del entregable.** Cada decisión tiene su ADR.

## Documentación

| Documento | Contenido |
|---|---|
| [Visión](./docs/vision.md) | Por qué existe el proyecto y qué demuestra |
| [Requisitos](./docs/requisitos.md) | Funcionales (FR) y no funcionales (NFR) |
| [Lenguaje ubicuo](./docs/lenguaje-ubicuo.md) | Glosario oficial ES→EN, convenciones y palabras prohibidas |
| [Diseño del dominio](./docs/dominio.md) | Bounded contexts, agregados, VOs, eventos, saga |
| [Value objects](./docs/value-objects.md) | Catálogo de VOs: invariantes, JSON, SQL, tests |
| [Eventos y comandos](./docs/eventos-y-comandos.md) | Catálogo de mensajes, productores, consumidores, idempotencia |
| [Stack](./docs/stack.md) | Decisiones técnicas acordadas y alternativas descartadas |
| [Escenarios](./docs/escenarios.md) | Flujos de compra y contrato HTTP esperado |
| [Roadmap](./docs/roadmap.md) | Hitos, alcance y test de aceptación de cada uno |
| [ADRs](./docs/adr/README.md) | Decisiones de arquitectura (formato Michael Nygard) |

## Inspiración (no copia)

El repo Java `ecommerce-microservices` aporta el **dominio y los escenarios**, no
el diseño. Sus ADR-0013 (compensación), 0014 (contrato 201), 0016 (reintento),
0019 (outbox ack) y 0020 (`monto = total`) se releen aquí y se rediseñan con las
herramientas que Go ofrece hoy. La suite SDET de ese repo sirve de referencia de
comportamiento observable, no de contrato a replicar literalmente (el contrato
nuevo usa RFC 9457 e `Idempotency-Key`).
