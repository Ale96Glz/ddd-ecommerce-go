# ADR-0007: Outbox con relay propio y Kafka

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** FR-09, FR-10, NFR-11
- **Nota:** corrige la decisión inicial de bus in-process; con microservicios
  (ADR-0002) el transporte tiene que cruzar límites de proceso.

## Contexto y Problema

Cuando un agregado cambia de estado y debe avisar a otro servicio, escribir en la
BD y publicar son dos operaciones que pueden fallar por separado. El repo Java
resolvió esto con outbox + Kafka (sus ADR-0013/0019). Aquí queremos **el patrón a
la vista**, no delegarlo a una herramienta, pero con microservicios el bus
in-process ya no es una opción.

## Opciones Evaluadas

1. **Relay propio + Kafka (elegida por el usuario)**.
   - Pros: ~180 líneas legibles con `FOR UPDATE SKIP LOCKED`; durabilidad en la
     BD del productor; ack del broker antes de marcar publicado; escalable a
     varias réplicas del relay; orden por agregado usando la clave de partición.
   - Contras: Kafka añade topics, consumer groups, lag y DLT que hay que operar.
2. **Debezium CDC sobre el WAL**.
   - Pros: cero código de relay.
   - Contras: esconde justo el patrón que se quiere aprender y añade Kafka
     Connect. Descartado.
3. **Bus in-process por canales** (decisión inicial, corregida).
   - Pros: cero infra.
   - Contras: no cruza procesos; incompatible con ADR-0002.
4. **Watermill / River**.
   - Pros: routers y colas maduros.
   - Contras: añaden conceptos antes de haber escrito el patrón. Descartados.

## Decisión

1. Cada servicio tiene su tabla `outbox_events`:
   `event_id` (PK, UUID v7), `aggregate_id`, `type`, `payload` JSONB,
   `occurred_at`, `trace_id`, `published_at NULL`.
2. El caso de uso escribe el evento **en la misma transacción** que el cambio del
   agregado y del estado de la saga.
3. El relay (mismo binario, modo `-worker`) hace
   `SELECT ... FOR UPDATE SKIP LOCKED LIMIT N`, publica a Kafka y marca
   `published_at = now()` **solo tras el ack del broker**. Reintento con backoff;
   si agota los intentos, el evento queda pendiente y se alerta (no hay DLT en el
   productor).
4. **Un topic por tipo de evento**: `order.placed`, `stock.reserved`,
   `payment.processed`, `payment.rejected`, `stock.release_requested`, etc.
   Catálogo completo en
   [eventos-y-comandos.md](../eventos-y-comandos.md).
5. La **clave de partición es `aggregateID`**, no el `eventID`: garantiza orden
   de los eventos de un mismo agregado dentro de su partición y permite
   múltiples particiones para escalar.
6. Headers Kafka por mensaje: `eventID`, `eventType`, `traceparent` (W3C),
   `schemaVersion`. El `eventID` es estable y viaja en el payload de la envelope.
7. Cliente Kafka: **`franz-go`** (`github.com/twmb/franz-go/pkg/kgo`), puro Go y
   moderno. Alternativa si hiciera falta: `IBM/sarama`.
8. Consumidores: un **consumer group por servicio**, `auto.offset.reset=earliest`,
   3 reintentos con backoff y DLT `<topic>.dlt`. Dedup con
   `processed_event(event_id PK)` en la BD del consumidor, aplicado **en la misma
   transacción** que el efecto (equivalente al `restock_event` del repo Java).
9. Kafka corre en **modo KRaft** (sin ZooKeeper) en `docker compose`.
10. Contrato de eventos versionado con `schemaVersion`; un cambio incompatible
    exige versión nueva y un ADR, nunca reescribir un evento publicado.

## Consecuencias

### Positivas

- Patrón didáctico completo: transacción local → relay → broker ack → dedup.
- Durabilidad y reanudación tras crash; reentregas con efecto único.
- Orden garantizado por agregado y escalado horizontal del relay.
- Trazas que atraviesan servicios mediante `traceparent` en headers.

### Negativas / Compromisos

- Kafka es infra nueva: hay que declarar topics antes de los consumidores,
  vigilar lag y operar DLTs (el repo Java ya documentó este problema en su
  ADR-0018).
- Consistencia eventual: entre el commit local y la compensación hay latencia;
  los tests y escenarios lo asumen con polling.
- El `published_at` se fija tras el ack, así que un fallo justo después deja un
  evento duplicado en el broker; la dedup por `eventID` lo absorbe.
- Sin esquema global de registro de eventos: se sustituye por contratos JSON
  versionados en `contracts/`.
