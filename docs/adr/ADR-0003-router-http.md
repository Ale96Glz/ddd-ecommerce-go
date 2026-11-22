# ADR-0003: `chi/v5` como router HTTP

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** NFR-03

## Contexto y Problema

La capa HTTP necesita enrutado, grupos de rutas y middleware (request ID, logging,
recover, autenticación). Hay que elegir entre el `ServeMux` del stdlib (con
patterns de Go 1.22+) y un router pequeño.

## Opciones Evaluadas

1. **stdlib `http.ServeMux`** (recomendación inicial).
   - Pros: cero dependencias, patterns `GET /v1/...` desde Go 1.22, idiomático.
   - Contras: middleware y grupos hay que escribirlos a mano; con cuatro
     contextos se repite código.
2. **`chi/v5` (elegida por el usuario)**.
   - Pros: `chi.RouteContext`, `r.Use(...)`, `r.Route("/v1", ...)`, muy estable y
     compatible con `net/http` puro (no secuestra `ResponseWriter`). Bien
     soportado por `oapi-codegen`.
   - Contras: una dependencia más.
3. **Gin / Echo**.
   - Pros: mucho middleware incluido.
   - Contras: esconden demasiado `net/http` para un proyecto de aprendizaje;
     sus contextos propios complican `context.Context` y OpenTelemetry.

## Decisión

1. Usar **`chi/v5`** como router.
2. Mantener **todo lo demás en `net/http` puro**: `http.Handler`,
   `http.Request.Context()`, `httptest` para tests.
3. Los handlers se generan desde el contrato OpenAPI
   ([ADR-0004](./ADR-0004-contrato-openapi.md)) y se montan en los grupos de
   `chi`.
4. El middleware se escribe a mano (pocas líneas) en
   `internal/platform/http/`: request ID, `slog`, recover y mapeo de `AppError`.

## Consecuencias

### Positivas

- Grupos por contexto (`/v1/orders`, `/v1/payments`) sin repetir prefijos.
- `httptest.NewRecorder()` funciona igual que con el stdlib.
- No hay framework que oculte el ciclo de vida de la petición.

### Negativas / Compromisos

- Confirmar al generar el código la versión de `oapi-codegen` y su auxiliar
  para chi (riesgo ya anotado en [stack.md](../stack.md)).
- `chi` no valida nada: la validación sigue siendo responsabilidad del borde
  (`validator/v10`) y de los value objects.
