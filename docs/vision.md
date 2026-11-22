# Visión y objetivos

## Por qué existe este proyecto

Aprender **desarrollo web moderno con Go** y construir algo que aporte valor real
a un portafolio. El motor es aprendizaje, no coste ni rendimiento: por eso se
prioriza **profundidad** en un dominio conocido antes que cantidad de servicios.

## Tesis

> Un e-commerce moderno no se define por tener muchos microservicios, sino por
> tener **contratos claros, límites correctos e invariantes imposibles de violar**.

El repo Java demostró que sabía desplegar microservicios con outbox y sagas. Este
proyecto demuestra que además sé **diseñarlos desde el tipo**: el hallazgo
H-PAGO-01 (`monto ≠ total`) se elimina aquí porque el pago recibe un `Money` que
el compilador iguala al total del agregado. No hay `if` que olvidar.

## Qué quiere demostrar a un reviewer

| Señal | Cómo se demuestra |
|---|---|
| DDD táctico bien aplicado | Value objects con invariantes, agregados con eventos de dominio, repositorios por contexto, glosario oficial |
| Microservicios bien separados | Un proceso y una BD por contexto; cero tablas compartidas; saga que compensa **entre servicios** |
| Concurrencia correcta en Go | Relay con `FOR UPDATE SKIP LOCKED`, consumer groups de Kafka con dedup, graceful shutdown |
| Corrección bajo fallos | Outbox durable con ack del broker, saga reanudable, reintentos y DLT |
| Contratos modernos | OpenAPI-first, `problem+json`, `Idempotency-Key`, `ETag`/`If-Match`, eventos versionados |
| Calidad de ingeniería | Tests puros de dominio + `testcontainers` (Postgres y Kafka) + ADRs que explican el porqué |
| Pragmatismo | Tres servicios en el MVP, `docker compose up`, sin Kubernetes |

## Lo que NO es

- ❌ Un clon del repo Java ni un "port" servicio por servicio.
- ❌ Otro CRUD de e-commerce con "clean architecture" de ocho carpetas.
- ❌ Una demo visual con frontend: es **backend puro**, HTTP API + eventos.
- ❌ Seis microservicios desde el día uno: el MVP son **tres** (`catalog`,
  `orders`, `payments`); auth, notificaciones y gateway llegan después
  ([ADR-0002](./adr/ADR-0002-microservicios-por-contextos.md)).
- ❌ Un proyecto con Kubernetes o Redis solo por Modernidad™.

## Definición de éxito

El proyecto está terminado cuando:

1. La **saga entre servicios** aguanta un fallo inyectado en `payments` y
   compensa el stock de `catalog` exactamente una vez, con el evento reenviado
   por Kafka.
2. Los tests de dominio corren en milisegundos **sin base de datos ni broker**.
3. Un `docker compose up` levanta los tres servicios, Postgres y Kafka, y un
   `curl` recorre el flujo completo.
4. Cada decisión no trivial tiene su ADR y su alternativa descartada documentada.
5. El glosario, los VOs y los contratos de eventos están alineados con el código.
