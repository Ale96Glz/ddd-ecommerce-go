# ADR-0001: Go estable, un solo `go.mod`, `cmd/` + `internal/`

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** NFR-10

## Contexto y Problema

El proyecto es de aprendizaje y portafolio. Hay que elegir la versión del
lenguaje y la estructura de módulos antes de escribir la primera línea. Un repo
con varios `go.mod` obliga a gestionar versiones cruzadas y `replace`, y un
layout plano mezcla código ejecutable con librerías internas.

## Opciones Evaluadas

1. **Última estable de Go, un `go.mod`, `cmd/<servicio>` + `internal/` (elegida)**.
   - Pros: una gráfica de dependencias; `internal/` impide importes desde fuera
     del módulo (el compilador lo garantiza); `cmd/` permite un `main` por
     servicio dentro del mismo monorepo.
   - Contras: un solo módulo significa que todo compila junto (aceptable para un
     monorepo de tres servicios).
2. **Módulos múltiples** (uno por servicio).
   - Pros: límites explícitos y versionado independiente.
   - Contras: fricción (`go.work`, `replace`) sin beneficio a esta escala; los
     servicios comparten `shared/` y conviene que compilen juntos.
3. **Layout `pkg/` para todo**.
   - Pros: familiar para quien viene de Java.
   - Contras: `pkg` es pública por convención; aquí queremos que el código sea
     privado al módulo.

## Decisión

1. Usar la **última estable de Go** disponible al crear el `go.mod`, fijada ahí.
2. **Un solo `go.mod`** en la raíz (monorepo).
3. Estructura:
   ```
   cmd/catalog/main.go    cmd/orders/main.go    cmd/payments/main.go
   internal/catalog/  internal/orders/  internal/payments/
   internal/shared/          (solo código; nunca BD compartida)
   api/catalog.yaml  api/orders.yaml  api/payments.yaml
   ```
4. Cada servicio tiene además un modo `-worker` en su propio `main`; no hay un
   binario global que sirva a todos (ADR-0002).
5. Configuración desde variables de entorno, sin Viper (ver
   [ADR-0011](./ADR-0011-identificadores-y-config.md)).

## Consecuencias

### Positivas

- `go build ./...` y `go test ./...` cubren todo desde la raíz.
- El compilador protege los límites del módulo sin fiestas de convenciones.
- Añadir un servicio nuevo es añadir un `cmd/` y un paquete en `internal/`.

### Negativas / Compromisos

- Un `go test ./...` global mezcla los tres servicios; se separan con
  `make test-unit` / `make test-integration` en el CI.
- Si un servicio necesitara versionar dependencias de forma independiente,
  habría que reabrir esta decisión (no previsto).
- Hay que revisar la versión de Go al retomar el proyecto tras meses.
