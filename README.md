# ddd-commerce-go — DDD táctico en Go moderno

Backend de e-commerce de **aprendizaje y portafolio**, escrito en Go con DDD
táctico (value objects, agregados, eventos de dominio, repositorios), **outbox
transaccional** y **sagas** como process manager explícito.

El proyecto nace con una tesis clara: un e-commerce no se define por cuántos
servicios tiene, sino por **contratos claros, límites correctos e invariantes
imposibles de violar**. Por eso las reglas de negocio viven en el tipo: el pago
recibe un `Money` igual al total del pedido y la diferencia de importes deja de
ser representable, no depende de un `if` que alguien pueda olvidar.

## Estado

| Fase | Estado |
|---|---|
| 0 — Documentación (visión, requisitos, ADRs, dominio) | ✅ |
| 1 — Núcleo de dominio puro (`Money`, `Order`, eventos) | ✅ |
| 2 — Persistencia y outbox (`pgx`/`sqlc`/`goose`, Kafka) | ⬜ |
| 3 — Saga `PlaceOrder` con compensaciones | ⬜ |
| 4 — Capa HTTP (OpenAPI + `chi`) | ⬜ |
| 5 — Worker, SSE y webhooks | ⬜ |

Código del hito 1 en `internal/`: dominio puro, sin DB ni HTTP.

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

## Qué demuestra

- **DDD táctico real:** value objects con invariantes, agregados que emiten
  eventos, repositorios por contexto y un glosario oficial.
- **Microservicios separados:** un proceso y una BD por contexto, cero tablas
  compartidas y una saga que compensa entre servicios.
- **Corrección bajo fallos:** outbox durable, relay con `FOR UPDATE SKIP LOCKED`,
  Kafka con dedup por `event_id` y DLT.
- **Concurrencia en Go:** consumer groups, idempotencia y apagado graceful.
- **Calidad de ingeniería:** tests puros de dominio, `testcontainers` y ADRs que
  explican el porqué de cada decisión.

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
