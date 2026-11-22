# Stack acordado

Decisiones cerradas en las reuniones de diseño. Lo que aparece como
**"acordado"** no se reabre sin corregir su ADR.

## Decisiones cerradas por el usuario

| Capa | Elección | Alternativa descartada | ADR |
|---|---|---|---|
| Lenguaje y módulo | Última estable de Go, un `go.mod`, `cmd/<servicio>` + `internal/` | Módulos múltiples (fricción sin beneficio) | [0001](./adr/ADR-0001-lenguaje-y-estructura.md) |
| Forma | **Microservicios por bounded context, BD por servicio** | Monolito modular (decisión inicial corregida), un servicio por agregado | [0002](./adr/ADR-0002-microservicios-por-contextos.md) |
| Capa de datos | **`pgx/v5` + `pgxpool` + `sqlc` + `goose`** | `database/sql` a mano, GORM/Ent | [0006](./adr/ADR-0006-pgx-sqlc-goose.md) |
| Async / eventos | **Relay propio (`FOR UPDATE SKIP LOCKED`) + Kafka** | Debezium CDC, bus in-process, Watermill, River | [0007](./adr/ADR-0007-outbox-relay-propio.md) |
| Broker | **Kafka en modo KRaft**, un topic por tipo de evento | NATS JetStream, Postgres `LISTEN/NOTIFY` | [0007](./adr/ADR-0007-outbox-relay-propio.md) |
| Cliente Kafka | **`franz-go`** | `IBM/sarama` | [0007](./adr/ADR-0007-outbox-relay-propio.md) |
| Router HTTP | **`chi/v5`** | `stdlib ServeMux` (Go 1.22+), Gin/Echo | [0003](./adr/ADR-0003-router-http.md) |
| Contrato HTTP | **OpenAPI + `oapi-codegen`**, un spec por servicio | `ogen`, escribir structs a mano | [0004](./adr/ADR-0004-contrato-openapi.md) |

## Resto del stack (aceptado por defecto)

| Necesidad | Elección | Por qué | ADR |
|---|---|---|---|
| IDs | UUID v7 (`google/uuid`) | Ordenables en el tiempo → mejor localidad en índices que v4 | [0011](./adr/ADR-0011-identificadores-y-config.md) |
| Dinero | Value object propio: `int64` cents + `ISO4217` | El tipo garantiza el invariante; sin librería decimal | [0005](./adr/ADR-0005-money-value-object.md) |
| Errores | `AppError` propio + RFC 9457 `problem+json` | Un solo punto mapea dominio → HTTP, sin librería | [0010](./adr/ADR-0010-errores-rfc9457.md) |
| Validación | `validator/v10` **solo en el borde HTTP** | Las invariantes reales van en el constructor del VO | [0010](./adr/ADR-0010-errores-rfc9457.md) |
| Observabilidad | `slog` (stdlib) + OpenTelemetry → OTLP, `traceparent` en headers de Kafka | Logs JSON y trazas que cruzan servicios | [0012](./adr/ADR-0012-observabilidad.md) |
| Auth | `golang-jwt/jwt/v5` + `argon2id` (`x/crypto`) | Estándar actual; argon2 sobre bcrypt | [0011](./adr/ADR-0011-identificadores-y-config.md) |
| Config | `os.Getenv` + parser pequeño por servicio | Pocas variables, todo desde entorno; sin Viper | [0011](./adr/ADR-0011-identificadores-y-config.md) |
| Tests | `testing` + `httptest` + `testcontainers-go` + `go-cmp` | Dominio puro, HTTP en memoria, Postgres y Kafka reales; sin `sqlmock` | [0013](./adr/ADR-0013-estrategia-de-tests.md) |
| Build / deploy | Multi-stage → binario estático distroless + `docker compose` | Imagen por servicio, mismo patrón GHCR conocido; K8s al final | [0014](./adr/ADR-0014-build-y-deploy.md) |

## Servicios del MVP

| Servicio | Contexto | BD | Publica | Consume |
|---|---|---|---|---|
| `catalog-service` | `catalog` | `catalog_db` | `stock.reserved`, `stock.released` | `stock.release_requested` |
| `orders-service` | `orders` + saga | `orders_db` | `order.placed`, `order.paid`, `stock.release_requested` | `payment.processed`, `payment.rejected`, `stock.released` |
| `payments-service` | `payments` | `payments_db` | `payment.processed`, `payment.rejected` | `order.placed`, `payment.void_requested` |

Fase siguiente: `auth-service`, `notification-service` y `api-gateway`.

## Lo que NO se incluye (y por qué)

| Tecnología | Motivo |
|---|---|
| Kubernetes | `docker compose` basta para enseñar el proyecto; se deja como evolución |
| Redis | No hay caché ni rate limiting en el MVP |
| `wire` / `dig` (DI) | Hand-wire en `main.go`; con tres servicios no duele |
| ORM | Pelea contra los agregados: invita a mutar entidades fuera de su raíz |
| Framework HTTP completo (Gin/Echo) | Esconden demasiado `net/http` para un proyecto de aprendizaje |
| MongoDB / otras BD | Postgres por servicio cubre todo el dominio |

## Riesgos conocidos del stack

- **`oapi-codegen` + `chi/v5`**: existen auxiliares para chi, pero hay que
  confirmar compatibilidad de versiones al generar el primer handler.
- **Kafka en local**: KRaft arranca más rápido que ZooKeeper, pero sigue siendo
  un contenedor más; hay que declarar topics antes que consumidores
  (el repo Java ya tuvo un ADR por eso).
- **DLT y lag**: los consumidores necesitan reintentos, DLT por topic y una
  consulta de lag en la demo; se documenta en `docs/eventos-y-comandos.md`.
- **Versiones**: las elecciones son por defecto; se fijarán al crear el `go.mod`.
