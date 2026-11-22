# ADR-0006: `pgx/v5` + `sqlc` + `goose`, sin ORM

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** NFR-04

## Contexto y Problema

La persistencia debe respetar los agregados: el `Order` controla sus propias
escrituras y el outbox se escribe en la **misma transacción** que el cambio de
estado. Un ORM tienta a mutar entidades fuera de su raíz y oculta el SQL;
`database/sql` a mano obliga a repetir `Scan` y pierde seguridad de tipos.

## Opciones Evaluadas

1. **`pgx/v5` + `pgxpool` + `sqlc` + `goose` (elegida por el usuario)**.
   - Pros: `sqlc` genera structs y métodos tipados desde SQL real; el SQL es
     visible y revisable; `pgx` es el driver más capaz de Postgres en Go;
     `goose` lleva migraciones SQL planas y reversibles. Errores de columna o
     tipo se detectan al compilar/generar.
   - Contras: hay que escribir el SQL y mantener el paso de generación.
2. **`database/sql` a mano**.
   - Pros: cero generación, control total.
   - Contras: boilerplate de `Scan`, más riesgo de bugs de tipo, peor experiencia
     con `JSONB` y `TIMESTAMPTZ`.
3. **GORM / Ent**.
   - Pros: se escribe menos código.
   - Contras: pelea con DDD táctico; oculta transacciones y hace fácil violar
     invariantes del agregado. Descartada.

## Decisión

1. **`pgxpool`** como pool global, inyectado en los repositorios.
2. **`sqlc`** genera el código de consultas desde `internal/*/infrastructure/sql/`.
   Todo acceso a tablas pasa por funciones generadas.
3. **`goose`** gestiona migraciones en `migrations/`; cada cambio de esquema es
   un par `.sql` (up/down) aplicado al arrancar en desarrollo y en `testcontainers`.
   Cada servicio ejecuta **solo sus propias migraciones** en su BD
   (`catalog_db`, `orders_db`, `payments_db`, ADR-0002).
4. **Una transacción por caso de uso**: repositorios reciben un `pgx.Tx` o un
   `DBTX` (interfaz mínima de `pgx`), nunca un pool suelto cuando hay que
   coordinar agregado + outbox.
5. El mapeo agregado ↔ fila vive en el repositorio (no en el dominio): el
   dominio no conoce `sqlc`.
6. No se usa `sqlmock`: los tests de persistencia corren contra Postgres real
   ([ADR-0013](./ADR-0013-estrategia-de-tests.md)).

## Consecuencias

### Positivas

- Errores de SQL se detectan en build, no en producción.
- La transacción agregado + outbox es explícita y auditable en el código.
- Sin magia: un reviewer ve exactamente qué consulta se ejecuta.

### Negativas / Compromisos

- `sqlc` no cubre consultas dinámicas complejas (no hay ninguna prevista).
- Hay que ejecutar `sqlc generate` tras cambiar migraciones o queries; se añade
  al CI para que un diff olvidado falle el build.
