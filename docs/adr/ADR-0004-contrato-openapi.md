# ADR-0004: OpenAPI-first con `oapi-codegen`

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** NFR-03

## Contexto y Problema

En el repo Java el contrato HTTP vivía en anotaciones de controladores y DTOs, y
los cambios se descubrían en tiempo de ejecución (o en la suite SDET). Para un
portafolio moderno el contrato debe ser la fuente de verdad y estar disponible
antes de escribir el handler.

## Opciones Evaluadas

1. **OpenAPI 3.1 + `oapi-codegen` (elegida por el usuario)**.
   - Pros: maduro, muy usado; genera tipos, interfaces de servidor y cliente
     desde el spec; integración con `chi`; `go generate` lo hace parte del build.
   - Contras: código generado que hay que commitear o regenerar en CI.
2. **`ogen`**.
   - Pros: más moderno y rápido, genera validación incluida.
   - Contras: menos adopción; su modelo de handlers es más opinado.
3. **Structs y handlers a mano desde el spec**.
   - Pros: control total, cero generación.
   - Contras: boilerplate y riesgo de desvío entre spec y código.

## Decisión

1. `api/openapi.yaml` es la **única fuente de tipos HTTP**.
2. `oapi-codegen` genera:
   - tipos de request/response,
   - la interfaz `ServerInterface`,
   - un cliente tipado para tests de aceptación.
3. El comando de generación se fija en un `//go:generate` y se ejecuta en CI; el
   diff del código generado se revisa como cualquier otro cambio.
4. El spec define esquemas `problem+json` reutilizables para todos los errores
   ([ADR-0010](./ADR-0010-errores-rfc9457.md)).
5. `Money` se representa en el spec como `{ amount: int64, currency: string }`
   (centavos), nunca como decimal JSON.

## Consecuencias

### Positivas

- Un cambio de contrato rompe la compilación: el error aparece antes del deploy.
- El cliente generado permite tests HTTP con tipos, sin concatenar JSON a mano.
- El spec sirve de documentación publicable sin trabajo extra.

### Negativas / Compromisos

- Hay que mantener el paso de generación sincronizado (CI que falla si el
  generado está desactualizado).
- `oapi-codegen` no valida invariantes de negocio: `amount ≠ total` sigue siendo
  responsabilidad del dominio ([ADR-0005](./ADR-0005-money-value-object.md)).
