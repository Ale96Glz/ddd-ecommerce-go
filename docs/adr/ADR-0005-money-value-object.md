# ADR-0005: `Money` como value object (`int64` centavos + ISO 4217)

- **Estatus:** Aceptado
- **Fecha:** 2026-11
- **Requisitos:** FR-05, FR-06, NFR-01

## Contexto y Problema

En el repo Java, `BigDecimal` + un `monto` recibido del cliente permitieron el
hallazgo H-PAGO-01: pagar `0.01` por un pedido de `30.00` y cerrarlo como
`PAGADO`. La validación era un `compareTo` en `PagoServiceImpl` (ADR-0020): un
`if` que cualquiera puede olvidar al añadir un nuevo camino de pago.

Aquí queremos que ese bug **no sea representable**.

## Opciones Evaluadas

1. **Value object propio, `int64` centavos + `currency` string (elegida)**.
   - Pros: comparable (`==`), sin errores de coma flotante, el constructor
     centraliza las invariantes, cero dependencias.
   - Contras: no soporta subdivisiones mayores que centavos (p. ej. cripto) ni
     redondeos complejos; aceptable para el dominio.
2. **`shopspring/decimal` u otra librería decimal**.
   - Pros: precisión decimal arbitraria.
   - Contras: no comparable directamente, añade dependencia, y **no impide** que
     el cliente envíe un monto distinto al total.
3. **`float64` + formateo**.
   - Pros: simple.
   - Contras: descartado. Dinero en float es un error clásico.
4. **Un tipo `Amount` genérico con escala configurable**.
   - Pros: flexible.
   - Contras: sobre-ingeniería; el dominio es EUR con 2 decimales.

## Decisión

1. ```go
   type Money struct {
       cents    int64
       currency string
   }
   ```
   - `NewMoney(cents, currency)` rechaza `cents < 0` y `currency == ""`.
   - `Add` / `Subtract` exigen la **misma moneda** → `ErrCurrencyMismatch`.
   - `Equals` usa comparación de structs.
2. El JSON expone centavos (`{"amount": 3000, "currency": "EUR"}`). El parseo
   desde texto (`"30.00"`) ocurre en un único adaptador del borde HTTP, nunca en
   el dominio.
3. **El comando de pago no recibe un `monto` del cliente**: recibe el `orderID` y
   el dominio toma `order.total` ([dominio.md](../dominio.md)). Si el borde HTTP
   recibe un `amount` explícito, debe validarlo contra el total y devolver `422`
   antes de construir el comando; es un caso de defensa en profundidad, no la
   fuente de la invariante.
4. `Order.Pay(amount Money)` exige `amount.Equals(o.total)` →
   `ErrAmountMismatch`, sin persistir ni emitir eventos.
5. Las columnas SQL son `amount_cents BIGINT` + `currency CHAR(3)`.

## Consecuencias

### Positivas

- H-PAGO-01 desaparece **por tipos**: no hay `BigDecimal.compareTo` que olvidar.
- Sumas y comparaciones son exactas y baratas.
- Los tests de dominio son puros y rápidos.

### Negativas / Compromisos

- La conversión a/decimal humano vive en el borde y hay que testearla una vez.
- No soporta monedas sin centavos ni criptomonedas; si llegara, se escribe un ADR
  nuevo con un tipo distinto, no se parchea `Money`.
- El `422` sigue existiendo en el contrato HTTP por compatibilidad conceptual con
  el repo Java, pero ya no es la única línea de defensa.
