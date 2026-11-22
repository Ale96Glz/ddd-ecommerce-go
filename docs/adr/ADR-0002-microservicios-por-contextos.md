# ADR-0002: Microservicios por bounded context, BD por servicio

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** NFR-11
- **Nota:** corrige la decisión inicial de monolito modular, antes de escribir código.

## Contexto y Problema

El dominio tiene cuatro contextos con vocabulario e invariantes propios
([lenguaje ubicuo](../lenguaje-ubicuo.md)). El repo Java ya demostró que sabemos
desplegar seis microservicios; aquí queremos **microservicios pero mejor
construidos por dentro**: DDD táctico, outbox, sagas explícitas y contratos
generados, no una traducción de sus servicios.

La decisión de forma hay que cerrarla antes del hito 1: determina el transporte
de eventos, la propiedad de la saga y el modelo de despliegue.

## Opciones Evaluadas

1. **Microservicios por bounded context + BD por servicio (elegida)**.
   - Pros: despliegue y escalado independientes; fallos aislados; esquemas que
     evolucionan por separado; obliga a diseñar de verdad outbox, saga,
     idempotencia y trazas distribuidas —justo lo que se quiere demostrar—.
   - Contras: más infra (Kafka, varios contenedores, healthchecks, secretos) y
     consistencia eventual en todo el flujo de compra.
2. **Monolito modular** (decisión inicial, corregida).
   - Pros: menos infra y transacciones locales para la saga.
   - Contras: no cumple el objetivo de microservicios; la saga no ejercita
     compensaciones entre procesos.
3. **Un microservicio por agregado**.
   - Pros: aislamiento máximo.
   - Contras: sobre-división; `Order`, `OrderItem` y `PlaceOrderSaga` pertenecen
     al mismo contexto y deben desplegarse juntos.

## Decisión

1. **MVP: tres servicios** — `catalog-service`, `orders-service`,
   `payments-service`. Fase siguiente: `auth-service`, `notification-service` y
   `api-gateway` (que valida JWT y enruta).
2. Cada servicio es **su propio proceso**, con su `cmd/<service>/main.go` y su
   base de datos lógica: `catalog_db`, `orders_db`, `payments_db`. En local, un
   solo contenedor Postgres con tres BDs; en producción, instancias separadas.
   **Ningún servicio lee las tablas de otro.**
3. **Un `go.mod` monorepo**: `internal/catalog`, `internal/orders`,
   `internal/payments`, `internal/shared`. Un servicio solo puede importar su
   propio contexto y `shared`; nunca el dominio de otro contexto.
4. Comunicación:
   - **HTTP** para consultas síncronas puntuales (p. ej. `orders` consulta
     disponibilidad a `catalog`).
   - **Kafka** para todos los eventos de dominio
     ([ADR-0007](./ADR-0007-outbox-relay-propio.md)).
5. Cada servicio tiene **su propio outbox y sus propios consumidores** con dedup
   en su BD. El productor nunca llama al consumidor directamente.
6. La saga `PlaceOrder` **vive en `orders-service`**, que es el dueño del
   proceso; su estado se persiste en `orders_db` y envía comandos a `catalog` y
   `payments` por Kafka ([ADR-0008](./ADR-0008-saga-process-manager.md)).
7. Los datos que otro servicio necesita se **congelan en el agregado**: el
   `unitPrice` ya se guarda en `OrderItem`, así que `payments` no consulta
   `catalog` para validar el total.
8. Contrato OpenAPI propio por servicio
   ([ADR-0004](./ADR-0004-contrato-openapi.md)) y `chi` como router
   ([ADR-0003](./ADR-0003-router-http.md)).

## Consecuencias

### Positivas

- Compilación y despliegue independientes por servicio; un fallo en `payments`
  no tira `catalog`.
- La saga ejerce compensaciones **reales entre procesos**, con dedup y DLT.
- Cada servicio elige su propio esquema y evoluciona sin coordinar migraciones.
- Trazas distribuidas de extremo a extremo (HTTP → Kafka → otro servicio).

### Negativas / Compromisos

- Infra nueva: Kafka en modo KRaft, varios contenedores, healthchecks y
  configuración por servicio.
- No hay commits que abarquen varios servicios: todo es eventualmente consistente
  y la saga es el único mecanismo de corrección.
- Las consultas que antes eran un JOIN ahora son datos congelados o llamadas HTTP
  (se decide caso por caso y se documenta en el ADR correspondiente).
- Más superficie de depuración; se compensa con `trace_id` propagado en headers
  de Kafka ([ADR-0012](./ADR-0012-observabilidad.md)).
