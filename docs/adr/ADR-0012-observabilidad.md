# ADR-0012: `slog` + OpenTelemetry → OTLP, `trace_id` en el outbox

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** NFR-06

## Contexto y Problema

El repo Java tenía Zipkin, Prometheus y Grafana. En Go queremos lo mismo con
menos piezas: logs estructurados que encajen con el JSON de `slog`, trazas
distribuidas y correlación entre una petición HTTP y los eventos que genera la
saga. Si el `trace_id` no viaja dentro del evento, depurar una compensación es
imposible.

## Opciones Evaluadas

1. **`slog` (stdlib) + OpenTelemetry → OTLP (elegida)**.
   - Pros: `slog` ya está en la stdlib y saca JSON; OTel es el estándar y permite
     exportar a cualquier backend (Jaeger, Tempo, el que sea); el `trace_id` se
     inyecta en los eventos del outbox y en los logs.
   - Contras: más configuración inicial que Zipkin directo.
2. **Zipkin directo**.
   - Pros: conocido del repo Java.
   - Contras: formato propietario; menos futuro.
3. **Solo logs JSON sin trazas**.
   - Pros: cero infra.
   - Contras: no se puede seguir una saga de 3 pasos; descartado.

## Decisión

1. `slog.NewJSONHandler(os.Stdout, ...)` como logger global inyectado por
   `context.Context`. Cada request lleva `request_id` y `trace_id`.
2. OpenTelemetry SDK con exporter OTLP; instrumentación HTTP de `chi` y `pgx`.
3. El middleware de trazas extrae el `trace_id` del `context` y lo añade a:
   - cada log de la petición,
   - cada `DomainEvent` (`trace_id` en la interfaz, ver
     [dominio.md](../dominio.md)),
   - la columna `trace_id` de `outbox_events`,
   - el header `traceparent` (W3C) de cada mensaje Kafka.
4. El consumidor extrae `traceparent` del mensaje y reanuda la traza, de modo que
   petición en un servicio → saga → compensación en otro aparecen en la misma
   traza distribuida.
5. Métricas básicas de OTel (latencia HTTP, tamaño del outbox, reintentos). Sin
   Prometheus/Grafana en el MVP: la demo corre con `docker compose` y un
   colector OTLP opcional.

## Consecuencias

### Positivas

- Una sola búsqueda por `trace_id` reconstruye el flujo completo.
- Cambiar de backend de trazas no cambia el código instrumentado.
- Logs legibles por máquina desde el día uno.

### Negativas / Compromisos

- El colector OTLP es un servicio más en `docker compose` (opcional, puede
  desactivarse).
- Instrumentar `pgx` y `chi` añade algo de sobrecarga y dependencias; se mide en
  el hito 4.
- Sin dashboards prefabricados: se documenta cómo consultar trazas en el README.
