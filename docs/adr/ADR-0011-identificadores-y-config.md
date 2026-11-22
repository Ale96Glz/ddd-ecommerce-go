# ADR-0011: UUID v7, JWT `jwt/v5` + argon2id, config por entorno

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** FR-01, NFR-07

## Contexto y Problema

Tres decisiones pequeñas que conviene fijar juntas porque afectan a los tipos y
al arranque: el tipo de ID, la librería de JWT/hashing y la forma de leer la
configuración. Elegir UUID v4 produces índices con mala localidad; Viper para
cuatro variables es excesivo; bcrypt es correcto pero argon2id es la elección
actual.

## Opciones Evaluadas

1. **UUID v7 + `golang-jwt/jwt/v5` + argon2id + `os.Getenv` (elegida)**.
   - Pros: IDs ordenables en el tiempo (mejor localidad en B-tree que v4);
     JWT v5 es el mantenido y seguro; argon2id resiste GPU mejor que bcrypt;
     el entorno es lo que ya usa el repo Java en K8s.
   - Contras: UUID v7 necesita una librería (`google/uuid`); argon2 consume algo
     más de CPU que bcrypt (params ajustables).
2. **UUID v4**.
   - Pros: más simple.
   - Contras: inserta aleatoriamente en índices → fragmentación.
3. **ULID**.
   - Pros: también ordenable, más compacto en base32.
   - Contras: menos soporte en librerías Go; UUID es más estándar.
4. **Viper para configuración**.
   - Pros: yaml, env, defaults, hot reload.
   - Contras: demasiado para ~10 variables; hot reload no se usa.

## Decisión

1. IDs de dominio como **UUID v7** (`google/uuid`), envueltos en tipos
   `OrderID`, `ProductID`, etc. ([dominio.md](../dominio.md)).
2. Autenticación con `golang-jwt/jwt/v5`; el middleware valida el token y
   mete el `sub` en `context.Context`. Las contraseñas se hashean con
   `argon2id` de `golang.org/x/crypto`.
3. Configuración desde **variables de entorno** con un parser pequeño en
   `internal/platform/config`: `DATABASE_URL`, `JWT_SECRET`, `APPROVAL_LIMIT_CENTS`,
   `OTEL_EXPORTER_OTLP_ENDPOINT`, `PORT`, etc. Sin Viper.
4. Los valores por defecto sirven solo para desarrollo local (`docker compose`);
   producción los inyecta el entorno.
5. Ningún secreto se commitea; el JWT secret se genera en la demo.

## Consecuencias

### Positivas

- Índices más eficientes y logs ordenables por ID.
- Menos dependencias que la alternativa con Viper.
- Mismo modelo de configuración que el stack Java conocido.

### Negativas / Compromisos

- Cambiar de UUID v4 a v7 en una BD ya poblada requeriría migración (no aplica:
  proyecto nuevo).
- Argon2id necesita calibrar parámetros (memoria/tiempo) antes de producción; se
  dejan valores razonables por defecto.
