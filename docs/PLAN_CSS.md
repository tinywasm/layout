---
PLAN: "`tinywasm/css`: cerrar el catálogo de tokens y retirar el DSL que refleja CSS"
---

> Depende de: [`PLAN.md`](PLAN.md) (§1 diagnóstico), [`PLAN_WIDGET`](PLAN_WIDGET.md)
> (§3 tabla de `Surface`, que define exactamente qué tokens faltan).
> Rompe API en la Etapa 5, y solo cuando ya no queda ningún consumidor del DSL viejo.

---

## Qué se conserva y qué se retira

`tinywasm/css` hace hoy **dos** cosas y solo una está bien:

| Responsabilidad | Veredicto |
|---|---|
| Catálogo `Token` tipado (`tokens.go`) | **Correcto.** Es un catálogo de design tokens de manual, alineable con W3C DTCG. Se conserva y se completa. |
| DSL espejo de CSS (`Display`, `Background`, `Padding`, `RawRule`, `Str`…) | **Se retira.** Es la fuente de todos los defectos de [`PLAN.md` §1](PLAN.md). Pasa a ser el motor de emisión **interno** que consume `widget/style`. |

La distinción importa: no se está tirando la librería, se está **quitándole la superficie
pública que permite escribir lo incorrecto**. `Token`, `Class`, `Stylesheet` y el emisor
siguen; lo que desaparece es la capacidad de que un consumidor escriba una propiedad CSS a
mano.

---

## Etapa 1 — Los tokens que faltan (aditivo, no rompe nada)

Derivados de la tabla `Surface` de [`PLAN_WIDGET` §3](PLAN_WIDGET.md). Son los que `crudview`
inventó con nombres de Material y que, por tanto, **nunca resuelven**: hoy el marco gris, las
hairlines y el estado deshabilitado de `crudview` ignoran el tema por completo, incluido el
modo oscuro.

```go
// Superficies — capa activa
ColorSurfaceSunken = Token{"--color-surface-sunken", "#E5E5EA"} // pozo dentro de un panel
ColorOutline       = Token{"--color-outline", "#D1D1D6"}        // hairline
ColorDisabled      = Token{"--color-disabled", "#E5E5EA"}
ColorOnDisabled    = Token{"--color-on-disabled", "#8E8E93"}

// Pares que faltaban en estados semánticos
ColorOnSuccess = Token{"--color-on-success", "#FFFFFF"}
ColorOnError   = Token{"--color-on-error", "#FFFFFF"}

// Capa fuente (gemelos claro/oscuro, como el resto del catálogo)
ColorSurfaceSunkenLight = Token{"--color-surface-sunken-light", "#E5E5EA"}
ColorSurfaceSunkenDark  = Token{"--color-surface-sunken-dark", "#21262D"}
ColorOutlineLight       = Token{"--color-outline-light", "#D1D1D6"}
ColorOutlineDark        = Token{"--color-outline-dark", "#30363D"}
ColorDisabledLight      = Token{"--color-disabled-light", "#E5E5EA"}
ColorDisabledDark       = Token{"--color-disabled-dark", "#21262D"}
ColorOnDisabledLight    = Token{"--color-on-disabled-light", "#8E8E93"}
ColorOnDisabledDark     = Token{"--color-on-disabled-dark", "#6E7681"}
```

Correspondencia con lo que hoy está hardcodeado en `crudview/css.go`:

| Constante local de `crudview` | Token que la reemplaza |
|---|---|
| `cInset  = "var(--color-surface-variant, #d7d7dd)"` | `ColorSurfaceSunken` |
| `cBorder = "var(--color-outline-variant, #cfcfd6)"` | `ColorOutline` |
| `cDisBg  = "var(--color-outline-variant, #c2c1c1)"` | `ColorDisabled` |
| `cDisFg  = "var(--color-on-surface-variant, #6e6e73)"` | `ColorOnDisabled` |
| `cOnAcc  = "#ffffff"` | `ColorOnPrimary`, tras la Etapa 2 |

También falta `Space0` (valor `0`): hoy el "sin gap" se escribe `Zero`, un valor genérico
fuera de la escala.

---

## Etapa 2 — Los pares deben contrastar, y hay que probarlo

`crudview/css.go:18-21` descarta un token del catálogo con esta razón:

> *"Not `--color-on-primary`: some themes set on-primary to a near black that is unreadable on
> the primary fill; white is the safe universal."*

Y hardcodea `#ffffff`. Eso no es un parche del consumidor: **es un defecto del catálogo**. Con
los valores actuales, `ColorPrimary` (`#00ADD8`) y `ColorOnPrimary` (`#1C1C1E`) forman un par
cuyo contraste no es el que un par `X`/`on-X` promete.

Dos cambios:

1. **Un tipo para el par**, de modo que un token `on-*` no pueda existir suelto:
   ```go
   // Pair es una decisión de superficie completa. Un fondo nunca se declara sin su
   // texto — impide por tipo el bug de rightpanel/css.go:17, donde ColorOnSurface
   // (un color de TEXTO) se declaró como fondo de panel.
   type Pair struct{ Bg, Fg Token }

   var (
       SurfacePrimary  = Pair{ColorPrimary, ColorOnPrimary}
       SurfacePanel    = Pair{ColorSurface, ColorOnSurface}
       SurfaceSunken   = Pair{ColorSurfaceSunken, ColorOnSurface}
       SurfaceSelected = Pair{ColorSelection, ColorOnSelection}
       SurfaceDanger   = Pair{ColorError, ColorOnError}
       SurfaceSuccess  = Pair{ColorSuccess, ColorOnSuccess}
       SurfaceDisabled = Pair{ColorDisabled, ColorOnDisabled}
   )
   ```

2. **`contrast_test.go` en la propia `css`**: para cada `Pair`, y para cada gemelo claro y
   oscuro, el ratio WCAG entre `Bg.Fallback` y `Fg.Fallback` debe ser ≥ 4.5:1 (≥ 3:1 para los
   pares que solo se usan bajo texto grande). Ajustar los fallbacks que fallen — empezando por
   `ColorOnPrimary`.

Con esto, el motivo por el que `crudview` hardcodeó blanco desaparece en su origen.

---

## Etapa 3 — Arreglar los defectos ya documentados por los consumidores

Ambos están descritos *en comentarios del código de este repo*, que es la definición de deuda
detectada y no reportada aguas arriba.

1. **`joinValues()` pierde los separadores.** `crudview/css.go:174-181`:
   *"`Padding(a,b,c,d)`'s output loses its spaces somewhere in the CSS pipeline:
   `padding:var(...)var(...)var(...)`, no separators — both verified via the live
   stylesheet"*. Es un bug silencioso: produce CSS sintácticamente inválido que el navegador
   descarta sin avisar. Arreglar `joinValues` y añadir un test de mesa con 2, 3 y 4 valores.

2. **`RawRule`s adyacentes se concatenan sin `;`.** Documentado también en
   `crudview/css.go:59-61` como un *"acuérdate de…"*. `RawRule` desaparece en la Etapa 5, pero
   mientras siga viva debe emitir su propio `;` — un escape que corrompe la salida en silencio
   es peor que no tenerlo.

---

## Etapa 4 — `Class` pasa a `widget` mediante alias (aditivo)

`widget.Class` debe ser el tipo canónico, porque `dom`, `form` y `view` lo necesitan sin
depender de una librería de estilo ([`PLAN.md` §5.3](PLAN.md)). Para no romper a nadie:

```go
// css/tokens.go
type Class = widget.Class   // ALIAS de tipo, no un tipo nuevo: todo el código
                            // existente sigue compilando sin tocar una línea.
```

`css` pasa a importar `widget` (que solo depende de `fmt`). La dirección es correcta y no hay
ciclo.

---

## Etapa 5 — Cerrar el DSL (rompe API; última etapa)

Solo cuando `layout`, `components` y `form` ya no lo usen — es decir, después de
[`PLAN.md` §8](PLAN.md) Etapa 4 y de [`PLAN_COMPONENTS`](PLAN_COMPONENTS.md).

**Se retira de la superficie pública:**

- `Str`, `Px`, `Em`, `Rem`, `Vh`, `Vw`, `Pct`, `Zero` — todo constructor de valor libre. Son
  el mecanismo exacto por el que un color o una longitud arbitraria entra al sistema.
- `RawRule` — el agujero sin tipar. **Sin sustituto y sin excepciones.**
- Las ~70 funciones espejo de propiedades CSS (`Display`, `Background`, `Padding`, `Position`,
  `Overflow`, `GridArea`, …).
- `Media`, `Selector`, `At` — la construcción de selectores y consultas arbitrarias.

**Permanece público:** `Token`, `Pair`, el catálogo, `Class` (alias), `Stylesheet` y
`NewStylesheet`.

**Permanece, pero como interno del módulo:** el emisor de reglas, consumido únicamente por
`widget/style`. Deja de ser API para el autor de widgets y pasa a ser el backend de la API de
intención.

Migración de un consumidor que se quede fuera de plazo: no hay ruta automática, y es
deliberado. Cada `RawRule` es una decisión que debe releerse una vez, porque casi todas
esconden un token fantasma o una unidad de viewport.

---

## Etapa 6 — Alineación con W3C DTCG (opcional, aditivo)

Con el catálogo completo y los pares probados, exportar/importar el formato DTCG cuesta un
archivo:

```go
//go:build !wasm
func (t Token) DTCG() json.Object  // {"$type":"color","$value":"#00ADD8"}
```

Beneficio: el catálogo se vuelve intercambiable con Figma, Style Dictionary y Tokens Studio —
el tema deja de ser algo que solo existe dentro de Go. No bloquea nada; se hace cuando haga
falta.

---

## Criterios de aceptación

1. `contrast_test.go` en verde para todo `Pair`, en claro y en oscuro.
2. Ningún `Token` referenciado desde el ecosistema queda sin declarar — test que cruza los
   `var(--…)` emitidos contra el catálogo.
3. `Padding` con 2, 3 y 4 valores emite CSS válido (test de mesa).
4. Tras la Etapa 5: `grep -rn "RawRule\|Str(" --include=*.go` vacío en `layout`, `components`
   y `form`.
5. `gotest` en verde en `css` y en todos sus consumidores.
