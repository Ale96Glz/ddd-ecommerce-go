# ADR-0014: Multi-stage → distroless + `docker compose`

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** NFR-09, NFR-10

## Contexto y Problema

El entregable debe arrancar con un comando y ser reproducible. El repo Java
desplegaba seis imágenes en GHCR y manifiestos K8s; para un proyecto de
portafolio eso es ruido. Pero sí queremos un build de producción creíble: binario
pequeño, sin shell, y arranque/apagado limpio.

## Opciones Evaluadas

1. **Dockerfile multi-stage → binario estático distroless + `docker compose`
   (elegida)**.
   - Pros: imagen mínima y reproducible; `docker compose up` levanta API + worker
     + Postgres (+ colector OTLP opcional); mismo patrón que el repo Java usa en
     GHCR, pero sin K8s.
   - Contras: distroless complica depurar dentro del contenedor (no hay shell).
2. **Imagen con Go y `go run`**.
   - Pros: simple.
   - Contras: imagen enorme y compilación en arranque; descartado.
3. **Kubernetes desde el día uno**.
   - Pros: realismo operativo.
   - Contras: fuera del objetivo de aprendizaje; se deja como evolución.

## Decisión

1. `Dockerfile` multi-stage por servicio: `golang` para compilar (CGO
   desactivado), etapa final `gcr.io/distroless/static-debian12` con el binario.
2. **Una imagen por servicio**, con dos modos de arranque: servidor HTTP y
   `-worker` (relay + consumidores de su BD). El worker escala de forma
   independiente porque usa Kafka (ADR-0007).
3. `docker compose.yml` con:
   - Postgres 16 con tres BDs lógicas (`catalog_db`, `orders_db`, `payments_db`),
   - Kafka en modo KRaft con los topics declarados al arrancar,
   - `catalog-service`, `orders-service`, `payments-service` y sus workers,
   - colector OTLP opcional.
4. Arranque y apagado con `context.WithCancel` y `signal.NotifyContext` para
   `SIGINT`/`SIGTERM`: el relay termina su lote, los consumidores Kafka hacen
   commit y el pool se libera.
5. CI: `go generate` (OpenAPI + sqlc) → `go test ./...` → build → publicar imagen
   en GHCR. Mismo patrón de release que el repo Java, simplificado.

## Consecuencias

### Positivas

- `docker compose up` es suficiente para la demo y para los tests E2E.
- Imagen pequeña y sin shell: superficie de ataque mínima.
- El apagado graceful se testea en el hito 5.

### Negativas / Compromisos

- Distroless dificulta el debugging en vivo; se compensa con logs estructurados y
  trazas ([ADR-0012](./ADR-0012-observabilidad.md)).
- Cada servicio añade una imagen y un healthcheck; el compose local crece frente
  al monolito inicial.
- Publicar en GHCR requiere el mismo token/permisos que el repo Java; se configura
  al final.
