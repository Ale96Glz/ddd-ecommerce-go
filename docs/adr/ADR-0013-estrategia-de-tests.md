# ADR-0013: `testing` + `httptest` + `testcontainers-go` + `go-cmp`

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** NFR-08

## Contexto y Problema

El repo Java usaba Mockito y H2. Para este proyecto importa que los tests
**demuestren** las invariantes: los de dominio deben correr en milisegundos sin
framework, y los de persistencia deben usar Postgres real para no ocultar bugs
de SQL. Un `sqlmock` puede dar verde mientras la consulta real falla.

## Opciones Evaluadas

1. **`testing` + `httptest` + `testcontainers-go` + `go-cmp` (elegida)**.
   - Pros: todo en la stdlib salvo dos dependencias pequeñas; fakes escritos a
     mano (adiós Mockito); Postgres real en integración; diffs legibles con
     `go-cmp`.
   - Contras: `testcontainers` necesita Docker en CI y arranca más lento.
2. **`sqlmock` + mocks generados**.
   - Pros: tests rápidos y sin Docker.
   - Contras: comprueba el SQL esperado, no el ejecutado; oculta errores de
     esquema y de transacción. Descartado.
3. **H2 / SQLite en memoria**.
   - Pros: rápido.
   - Contras: no es Postgres; `JSONB`, `SKIP LOCKED` y `TIMESTAMPTZ` se comportan
     distinto. Descartado.

## Decisión

1. **Dominio**: `go test` puro, tabla de casos, fakes minimalistas. Sin importar
   `net/http`, `pgx` ni `chi`. Es la joya del portafolio y debe correr en <100 ms.
2. **HTTP**: `httptest.NewServer` / `NewRecorder` sobre el handler generado
   ([ADR-0004](./ADR-0004-contrato-openapi.md)). Cubre los escenarios de
   [escenarios.md](../escenarios.md).
3. **Integración**: `testcontainers-go` levanta Postgres y Kafka (KRaft), aplica
   migraciones `goose` y ejecuta repositorios + relay reales. Cubre transacción
   agregado+outbox, `SKIP LOCKED`, publicación con ack y dedup del consumidor.
4. **Aserciones**: `github.com/google/go-cmp/cmp` para estructuras; `errors.Is`
   para errores de dominio.
5. **Saga**: tests de unidad con eventos inyectados + un test de integración con
   fallo inyectado en `AuthorizePayment`.
6. Cobertura objetivo: dominio ≥ 90 %, capa HTTP sencilla (no se obsesiona con
   porcentajes globales).

## Consecuencias

### Positivas

- Un fallo en una query se ve en CI, no en producción.
- Los tests de dominio explican el modelo mejor que cualquier diagrama.
- Sin magia de mocks: un fake de 5 líneas es legible.

### Negativas / Compromisos

- CI necesita Docker; se documenta el requisito y se separan los tests rápidos
  (`make test-unit`) de los de integración (`make test-integration`).
- Los tests de integración son más lentos; se limitan a los caminos que realmente
  tocan BD y concurrencia.
