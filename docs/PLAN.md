---
PLAN: "layout: migrar crudview, platformd y rightpanel al contrato visual widget/style"
EXECUTOR: jules
---

> Este plan se despacha con el flujo CodeJob. Ver skill: **agents-workflow**.
> Documento de diseño de fondo (lectura opcional):
> [`VISUAL_CONTRACT_MASTER_PLAN.md`](VISUAL_CONTRACT_MASTER_PLAN.md) §8.

# Plan — `tinywasm/layout`: migración al contrato visual

`layout` es el **último consumidor** de la cadena. Cuando este plan cierre, la verificación
visual en vivo la hace un humano sobre la app real: es el único punto donde se comprueba que
todo el árbol migrado se ve bien de verdad.

---

## 🚦 0. Bloqueo previo — no empieces sin esto

`layout` va **el último**, detrás de otros tres repos que publican en este orden: `ssr` (gate) →
`components` y `widget` (en paralelo) → `layout`.

Este plan requiere **dos publicaciones previas**:

1. **`github.com/tinywasm/components`** con su anatomía nueva: clases derivadas de `widget.Name`
   y su hoja expuesta como **`RenderCSS() *css.Stylesheet`**. Si migras `layout` contra una
   versión anterior, `crudview` compone widgets sin la anatomía nueva y el resultado no compila
   o se ve roto.
2. **`github.com/tinywasm/widget`** sin `style.Styler` y **con** la escala `Motion` /
   `Animate(m)`, que `platformd` consume en la etapa 5. Ese chequeo está en la etapa 1.1.

> **`Style()` no existe y no debe existir.** Hubo una ventana en la que `components` expuso su
> hoja como `Style() *style.Sheet` y `ssr` reconocía **tres** entradas de CSS
> (`RootCSS`, `RenderCSS`, `Style`). Esa tercera vía se eliminó: el contrato SSR de CSS es
> `RootCSS()` para los tokens de documento y `RenderCSS()` para la hoja de un componente, y nada
> más. Ver
> [`components/docs/PLAN.md`](https://github.com/tinywasm/components/blob/main/docs/PLAN.md) y
> [`ssr/docs/PLAN.md`](https://github.com/tinywasm/ssr/blob/main/docs/PLAN.md).
>
> Consecuencia directa para este plan: **los `css.go` de `layout` conservan su firma actual**
> `func (x *T) RenderCSS() *Stylesheet`. Lo que cambia es el **cuerpo** — pasa a construirse con
> el DSL `style.Of(...)` y termina en `.Stylesheet()`. Si en algún momento te ves renombrando un
> `RenderCSS` a `Style`, has leído un documento obsoleto: **para y repórtalo**.

**Comprobación obligatoria antes de la etapa 1:**

```bash
go list -m -versions github.com/tinywasm/components
```

Toma la versión **mayor que v0.1.9** (la última publicada antes de la migración) y confirma que
expone `RenderCSS()`, no `Style()`:

```bash
go doc github.com/tinywasm/components/fieldset.Fieldset.RenderCSS   # debe existir
go doc github.com/tinywasm/components/fieldset.Fieldset.Style       # debe FALLAR
```

Si el primero falla, si el segundo **no** falla, o si no hay versión posterior a v0.1.9,
**detente y repórtalo**. No migres `layout` contra un `components` sin migrar, y no añadas un
`replace` local para esquivarlo.

---

## ⚠️ 1. Alcance — LEE ESTO ANTES DE TOCAR NADA

Se migran **tres paquetes en el mismo cambio**: `rightpanel`, `crudview` y `platformd`. No es por
etapas con esperas entre medio; el gate es uno solo, `gotest` en verde al final.

**PROHIBIDO — no hagas nada de esto:**

| Prohibición | Motivo |
|---|---|
| Usar `RawRule(` | Es el agujero sin tipar que este plan cierra. Si el vocabulario no alcanza, **para y repórtalo**: es un defecto aguas arriba en `widget`/`css`, no algo a parchear aquí. |
| Usar `Str(` para colores, longitudes o breakpoints | Misma razón. Si no existe el tipo, no se escribe el valor. |
| Escribir un literal de color (`#rrggbb`, `rgb(`, `hsl(`) o un nombre de color CSS | La paleta la posee `css` v0.3.0. |
| Usar unidades de viewport (`vw`, `vh`, `vmin`, `vmax`) | Toda medida es relativa al contenedor. Lo resuelven las primitivas `Flow` y `@container`. |
| Escribir `Media(` | Responsivo es el default. Las primitivas ya se adaptan. |
| Declarar un `Token` o un bloque `Root(Declare(...))` | La escala la posee `css`. Un componente que declara tokens crea un catálogo paralelo. |
| Escribir una `Class` a mano | Toda clase se deriva de un `widget.Name` con `.Root()` / `.Class(Part)`. |
| Añadir un `replace` nuevo en `go.mod` | Se migra contra versiones publicadas. |
| Tocar la lógica de negocio (`crud.go`, `factory.go`, `platformd.go`, `rightpanel.go`, `crudview.go`) más allá de las clases que emiten | Este plan es visual. |
| Usar `go test` | En este repo se usa `gotest`. |

**Anti-footgun:** los archivos `css.go` y `svg.go` llevan `//go:build !wasm` y **deben
conservarlo**. `widget/style` y `css` no pueden entrar al binario WASM; hay un criterio de
aceptación que lo verifica (§8.2).

---

## 2. Punto de partida medido

Estado real de este repo, contado antes de escribir el plan:

| Paquete | `css.go` | `RawRule(` | `Media(` | `vw`/`vh` | literales de color | `Str(` | tokens propios |
|---|---|---|---|---|---|---|---|
| `rightpanel` | 166 líneas | 11 | 1 | 11 | 0 | 9 | 10 `--rp-*` |
| `crudview` | 369 líneas | 9 | 1 | 8 | 8 | 61 | 0 (3 fantasma `--cv-*`) |
| `platformd` | 521 líneas | 11 | 2 | 11 | 4 | 51 | 7 `--pd-*` en `tokens.go` |

Ninguno importa `github.com/tinywasm/widget` todavía.

---

## 3. Vocabulario disponible — úsalo, no inventes

Todo esto ya está publicado. Si crees que falta algo, **para y repórtalo** (§1).

La firma que envuelve todo esto **no cambia**: cada `css.go` sigue declarando
`func (x *T) RenderCSS() *Stylesheet`, y el DSL se cierra con `.Stylesheet()`:

```go
//go:build !wasm

func (v *CrudView) RenderCSS() *Stylesheet {
	return style.Of(nameCrudView).
		Root(style.Split(style.RatioTwoThirds, style.Space2), style.On(style.Accent)).
		Part(partDetail, style.Stack(style.Space2), style.Fill()).
		Stylesheet()   // <- cierra la cadena; la firma sigue siendo RenderCSS
}
```

```go
// github.com/tinywasm/widget — identidad, anatomía, estados
type Name string          // .Root() -> "nombre" · .Class(Part) -> "nombre__parte"
type Part string
Kind: Region · Listbox · Menu · Dialog · Disclosure · Tabs · Toolbar · Grid · Combobox · Form · Alert
State: Selected · Disabled · Locked · Invalid · Busy · Open · Current   // -> data-*="true"
Cue: Hover · Focus · Press · Target

// github.com/tinywasm/widget/style — solo !wasm
style.Of(name).Root(opts...).Part(p, opts...).When(state, part, opts...).Cue(cue, part, opts...)

// Disposición (todas fluidas, sin media queries)
Stack(gap) · Row(gap) · Split(ratio, gap) · Grid(track, gap)
Center() · Cover() · Reel(gap) · Frame(ratio)

// Superficie: fondo + texto + borde SIEMPRE juntos
On(Page|Panel|Sunken|Accent|Selected|Danger|Success|Muted|Disabled|PanelHover)

// Escalas cerradas
Pad(Space) · Round(Radius) · Raise(Elevation) · Text(TextSize) · FontWeight(Weight) · Width(Size)

// Las excepciones — lo único que se declara explícitamente
Fill() · Scrolls() · Fixed() · Flush() · Clip()

// Superposición — el vocabulario que sustituyó al último escape hatch (css.Raw)
Backdrop(Scope) · Above() · Scrim() · Hidden() · Shown()

// Movimiento — escala cerrada; el easing no se elige y prefers-reduced-motion
// lo emite la librería sola
Animate(MotionNone|MotionFast|MotionBase|MotionSlow)
```

---

## Etapa 1 — `go.mod`

Archivo: `go.mod`

### 1.1 Fijar versiones publicadas

```
github.com/tinywasm/css        v0.3.0    // era v0.1.4
github.com/tinywasm/widget     v0.3.0+   // NUEVA dependencia directa
github.com/tinywasm/form       v0.3.0    // era v0.2.25
github.com/tinywasm/view       v0.1.10   // era v0.1.2
github.com/tinywasm/components <la del gate §0>   // era v0.1.8
```

**No copies estos números a ciegas.** Son el suelo, no el techo: `widget` y `components`
publican versiones nuevas al cerrar sus propios planes. Para cada uno, resuelve la versión real
con `go list -m -versions <módulo>` y toma la **más alta publicada**. `widget` debe ser una
posterior a la que eliminó `style.Styler`; compruébalo con:

```bash
go doc github.com/tinywasm/widget/style.Styler   # debe FALLAR  (interfaz eliminada)
go doc github.com/tinywasm/widget/style.Animate  # debe existir (escala Motion, etapa 5)
```

Si el primero **tiene éxito**, o si el segundo **falla**, `widget` aún no ha publicado su plan:
**detente y repórtalo**.

### 1.2 Desactivar los `replace` de desarrollo local — **comentándolos, no borrándolos**

Hay cinco, todos marcados `TEMP`: `dom`, `components`, `form`, `css`, `view`.

**No los borres. Coméntalos con `//`.** Esta migración toca los tres paquetes a la vez y es
seguro que queden bugs por depurar; volver a apuntar a un checkout local tiene que costar
descomentar una línea, no reconstruir la directiva de memoria.

Deja el bloque así, al final de `go.mod`, conservando el comentario original de cada uno para
que se sepa por qué existía:

```go
// ── replaces de desarrollo local ─────────────────────────────────────────────
// Desactivados en la migración al contrato visual: se migra contra versiones
// publicadas. Descomenta el que necesites para depurar contra un checkout local
// y vuelve a comentarlo antes de cerrar el PR.
//
// TEMP: local dom checkout carries the BindChildren initial-row wiring fix.
// replace github.com/tinywasm/dom => ../dom
//
// TEMP: local components checkout so demo edits render live via the dev server.
// replace github.com/tinywasm/components => ../components
//
// TEMP: local form checkout — field-wrapper/input id collision fix.
// replace github.com/tinywasm/form => ../form
//
// TEMP: local css checkout — brand palette set to the Pa100T reference.
// replace github.com/tinywasm/css => ../css
//
// TEMP: local view checkout — conformance.Driver gained New/Edit/FocusedFieldID.
// replace github.com/tinywasm/view => ../view
```

Procedimiento: coméntalos **uno a uno** y ejecuta `gotest` tras cada uno. Si al comentar uno los
tests fallan, **no lo reactives en silencio: detente y repórtalo**, indicando cuál y qué falla.
Un `replace` activo convierte en mentira la frase "migrado contra versiones publicadas".

Tras `go mod tidy`, **verifica que el bloque comentado sigue en el archivo**. Si la herramienta
lo elimina, vuelve a añadirlo al final de `go.mod` — es documentación operativa, no residuo.

Atención especial al de `css`, cuyo comentario dice *"brand palette set to the Pa100T reference
(steel blue + white text)"*: si al comentarlo la paleta cambia, es que esa marca **no está** en
`css` v0.3.0 — repórtalo, no la re-declares localmente aquí.

---

## Etapa 2 — `layout_conformance_test.go` (se escribe PRIMERO)

Archivo: `layout_conformance_test.go`, en la raíz del módulo. **Escríbelo antes de migrar nada.**
Va a fallar al principio; ése es el objetivo: fija la meta y mide el avance.

Recorre todos los `.go` del repo (excluyendo `docs/`, `_test.go` y `web/`) por AST y falla si
encuentra:

1. Un literal de color: `#rgb`, `#rrggbb`, `rgb(`, `hsl(`, o un nombre de color CSS.
2. Una llamada a `RawRule(`.
3. Una llamada a `Media(`.
4. Una unidad de viewport: `vw`, `vh`, `vmin`, `vmax`.
5. Una llamada a `Str(`.
6. Una llamada a `Declare(` o a `RootCSS(`.
7. Una `var(--…)` cuyo nombre no exista en el catálogo de `css` v0.3.0.
8. Una constante de tipo `Class` asignada desde un literal de string.

Detecta los paquetes por su **ruta de import resuelta en el bloque de imports**, nunca por el
texto del selector: este repo usa dot-imports (`. "github.com/tinywasm/css"`) y el emparejamiento
textual daría falsos negativos.

Este test queda en CI como guardia permanente: cualquier regla nueva que hardcodee un color falla
el build.

---

## Etapa 3 — `rightpanel` (el más pequeño: prueba del vocabulario)

166 líneas, 10 tokens `--rp-*`. Es el primero **porque es la prueba real de si el vocabulario
alcanza**. Si aquí falta algo, se reporta aguas arriba (§1) — no se abre un escape hatch.

1. `Name` + `Kind` (`widget.Region`) + partes nombradas, derivadas de las clases actuales.
2. Borrar el bloque `Root(Declare(...))` y las 10 constantes `Token`. Sustituciones:
   - `tokenAsideBg`, `tokenBorderColor`, `tokenBg` → `On(Panel)`.
     **Nota:** hoy `tokenAsideBg` vale `ColorOnSurface.Var()` — usa el color **de texto** como
     color **de fondo**. Es un bug preexistente; `On(Panel)` lo corrige por construcción.
   - `tokenMainWidth` (`66vw`) + `tokenAsideWidth` (`30vw`) → `Split(style.RatioTwoThirds, Space2)`.
   - `tokenTitleHeight`, `tokenContentHeight`, `tokenControlsHeight` (`8vh`/`89vh`/`3vh`) →
     `Fill()` sobre las partes que deben tomar el alto; nada de alturas explícitas.
   - `tokenGap` → el `gap` de la primitiva.
   - `tokenTitleColor` → `On(...)` de la superficie que corresponda.
3. Borrar el bloque `Media("(max-width: 640px)")`: lo cubre `Split`.
4. Conservar el comentario que dice que la apariencia del campo de formulario **no** se define
   aquí (vive en `components/fieldset`). Sigue siendo cierto.

---

## Etapa 4 — `crudview`

369 líneas, 8 literales de color, 61 `Str(`, 3 tokens fantasma `--cv-*` (usados en `RawRule` y
**nunca declarados**, así que hoy resuelven siempre al fallback).

### 4.1 Anatomía

`Name` = `crudview`, `Kind` = `widget.Region`. Mapeo de clases actuales a partes:

| Clase actual (`crudview.go`) | Parte | Disposición |
|---|---|---|
| `cv-module-content` | *root* | `Split(style.RatioTwoThirds, Space2)`, `On(Accent)`, `Flush()` |
| `cv-article-contend` | `detail` | `Stack(Space2)`, `Fill()` |
| `cv-box-content` | `fields` | `On(Sunken)`, `Pad(Space2)`, `Scrolls()`, `Round(RadiusMd)` |
| `cv-aside-wrap` | `aside` | `Stack(Space1)`, `On(Panel)`, `Pad(Space1)`, `Fill()` |
| `cv-lista-box` | `list` | `On(Sunken)`, `Scrolls()`, `Round(RadiusMd)` |
| `cv-aside-actions` | `actions` | `Row(Space1)` |
| `cv-title-container` / `cv-title` | `title` | `Row(Space1)`, `Fixed()` |
| `cv-btn-crud` | `action` | `On(Accent)`, `Round(RadiusMd)` |
| `cv-btn-crud-icon-hidden` | — | **desaparece**: es `widget.Open`, no una clase |
| `cv-back` | — | **desaparece**: con `Split` el reflow no necesita botón de vuelta |

Las clases restantes de `crudview.go` que no estén en la tabla se convierten en partes con el
mismo criterio: nombre en inglés, sin el prefijo `cv-`.

### 4.2 Borrar las compensaciones de internos ajenos

Dos hacks que existen solo porque a los componentes les faltaba anatomía. **Ambos se borran**:

1. **`css.go:164-175`** — los cuatro longhands de padding con `PaddingTop(Space3)` y el
   comentario de 8 líneas que explica que el chip de `fieldset` invade su propio borde. Con la
   anatomía de `fieldset`, ese ajuste vive en `fieldset`. Aquí se usa `Pad(Space2)` y punto.
2. **`css.go:292`** — el bloque *"Fixed, not static: takes this mount point (and modaldialog's
   own hidden-state anchor div inside it) out of `clsModuleContent`'s grid item participation"*.
   `modaldialog` declara `Kind = widget.Dialog` y se posiciona solo. El punto de montaje deja de
   ser un concepto aquí.

Si al borrarlos algo se ve mal, **el defecto es de `components`** y se reporta allí; no se
recompensa desde `layout`.

### 4.3 Bloque `Media`

El `Media("(max-width: 640px)")` completo se borra. Lo cubre `Split`, que se apila bajo su propio
ancho vía `@container`. No hay `direction:rtl` que razonar.

---

## Etapa 5 — `platformd` (el más grande)

521 líneas, 11 `vw`/`vh`, 7 tokens en `tokens.go`.

1. **Borrar `platformd/tokens.go` entero** y el bloque `Root(Declare(...))`.
   - `tokenMenuSize` (`4vw`), `tokenHeaderHeight` (`3vh`) → escala `Space`.
   - `tokenContentHeight` (`97vh`, y su `calc(100vh - 2.8rem)`) → `Cover()` + `Fill()`.
   - `tokenFontSizeNormal`/`tokenFontSizeSmall` → `Text(TextBase)` / `Text(TextXs)`.
   - `tokenSlideDur` (`0.6s`) → **`Animate(MotionSlow)`**.
   - `tokenTransitionWait` (`0s`) → **desaparece**. Sin retardo es el default; no se escribe.

     Este hueco existía cuando se escribió la primera versión de este plan y **ya está
     cerrado**: `widget/style` expone ahora la escala `Motion`
     (`MotionNone` · `MotionFast` · `MotionBase` · `MotionSlow`) y el `Opt` `Animate(m Motion)`,
     que consume los tokens `--duration-*` / `--ease-in-out` que `css` ya poseía. El easing no
     se elige, y la supresión por `prefers-reduced-motion` la emite `widget/style` sola: aquí no
     se escribe nada de eso.

     **`0.6s` pasa a 400ms y eso es correcto**, no un redondeo a corregir. La escala es cerrada;
     las duraciones elegidas a mano dejan de existir. Si echas de menos un peldaño intermedio,
     **para y repórtalo** — se amplía el enum aguas arriba en `widget`, se publica y se sigue.
     Lo que no se hace nunca es recuperarlo con un `Token` local, un `RawRule(` o un `Str(`.

     Requisito de versión: el `widget` que fijes en la etapa 1 debe exponerlo. Compruébalo con
     `go doc github.com/tinywasm/widget/style.Animate` — si falla, `widget` aún no ha publicado
     su plan: **detente y repórtalo**.
2. `Name` + `Kind` + partes, mismo criterio que §4.1.
3. Borrar los dos bloques `Media`.
4. `svg.go` (`IconSvg()`) **no se toca**. No es un hueco pendiente: `IconSvg() *sprite.Sprite`
   ya es un contrato tipado, propiedad de `tinywasm/svg`, y es el que `ssr` recolecta — el
   equivalente exacto de `RenderCSS()` para iconos. Los iconos no entran en el contrato visual.

---

## Etapa 6 — Test de forma-consumidor

Archivo: `crudview/consumer_test.go` (ya existe y recorre la pila real).

Extenderlo para aseverar la **hoja emitida**, no solo el markup:

1. No contiene `!important`.
2. Sus `@layer` aparecen en el orden `tokens, primitives, widgets, states`.
3. **Cada clase presente en el markup renderizado existe en la hoja, y cada selector de clase de
   la hoja aparece en el markup.** Este es el par que hoy nadie verifica y que permitió que
   `cv-btn-crud-icon-hidden` fuera un estado disfrazado de clase. Escríbelo en el mismo cambio,
   no como paso posterior.

---

## 7. Lo que NO entra en este plan

- **`tinywasm/view`** no se migra aquí. No importa `widget` todavía y no bloquea: los estados que
  `crudview` necesita los emiten los widgets de `components`.
- **`svg.go`** de `crudview` y `platformd`: `IconSvg()` sigue igual.
- **La verificación visual en vivo** (§9) la hace un humano después de que este PR cierre. El
  agente termina en `gotest` verde.

---

## 8. Criterios de aceptación — verificables con grep

1. `gotest` en verde en `layout`, incluido `layout_conformance_test.go`.
2. `GOOS=js GOARCH=wasm go list -deps ./...` **no** contiene `tinywasm/widget/style` ni
   `tinywasm/css`.
3. `grep -rn "RawRule(\|Str(" --include='*.go' .` → **vacío** (fuera de `_test.go`).
4. `grep -rnE '#[0-9a-fA-F]{3,6}' --include='*.go' .` → **vacío** (fuera de `_test.go`).
5. `grep -rnE '[0-9.]+(vw|vh|vmin|vmax)' --include='*.go' .` → **vacío**.
6. `grep -rn "Media(\|Declare(\|RootCSS(" --include='*.go' .` → **vacío**.
7. `ls platformd/tokens.go` → **no existe**.
8. `grep -rn -- "--pd-\|--rp-\|--cv-" .` → **vacío**.
9. `grep -nE '^[[:space:]]*replace' go.mod` → **vacío**: ningún `replace` **activo**. El bloque
   comentado de la etapa 1.2 debe seguir presente y **no** cuenta como violación —
   `grep -c '// replace' go.mod` debe dar **5**. Si algún `replace` tuvo que quedarse activo, va
   acompañado de su justificación reportada en la etapa 1.2.
10. `grep -rn "Class = \"" --include='*.go' .` → **vacío**: toda clase se deriva de un `widget.Name`.

---

## 9. Verificación visual en vivo (la hace un humano, tras el merge)

Este es el motivo de que `layout` sea el último: es donde se comprueba de verdad. Con la app
corriendo, revisar en escritorio y en móvil emulado, en claro y en oscuro:

1. Los tres paneles rellenan el alto de su contenedor, sin franjas grises.
2. `crudview` se apila correctamente bajo su propio ancho, sin botón de vuelta.
3. El chip de etiqueta de `fieldset` no queda cortado ni empuja el contenido.
4. El menú ⋮ de `targetlist` abre hacia arriba en la última fila y se cierra al hacer clic fuera.
5. `modaldialog` se centra y no duplica el margen inferior.
6. **El marco gris cambia en modo oscuro.** Hoy no lo hace: usa `var()` con fallback a un hex
   claro. Es la comprobación que valida todo el trabajo de tokens.

---

## 10. Checklist de calidad Go (obligatorio)

- **Sin strings repetidos**: toda clase sale de `widget.Name`; ningún literal `"cv-..."`,
  `"--rp-..."` ni `"#..."` en la lógica.
- **Errores** con `github.com/tinywasm/fmt` (`fmt.Err(...)`), nunca `errors`/`fmt` de stdlib.
- **`//go:build !wasm`** se conserva en todo `css.go` y `svg.go`.
- **Cero `any`, cero `map`** en API nueva.
- Comentarios que expliquen un *"acuérdate de…"* se borran: si hay que recordarlo, es un hueco
  del arnés, no una línea de manual.

---

## 11. Tabla de etapas

| # | Etapa | Archivos | Gate |
|---|---|---|---|
| 0 | *(bloqueo)* `components` con `RenderCSS()` **y** `widget` con `Animate`/sin `Styler`, ambos publicados | — | §0: `go doc …/fieldset.Fieldset.RenderCSS` existe y `….Style` falla · §1.1: `go doc …/style.Animate` existe y `….Styler` falla |
| 1 | Versiones y `replace` | `go.mod`, `go.sum` | `go build ./...` |
| 2 | Auditoría ejecutable | `layout_conformance_test.go` (nuevo) | compila (falla a propósito) |
| 3 | `rightpanel` | `rightpanel/css.go`, `rightpanel/rightpanel.go` | compila |
| 4 | `crudview` | `crudview/css.go`, `crudview/crudview.go` | compila |
| 5 | `platformd` | `platformd/css.go`, `platformd/platformd.go`; borrar `platformd/tokens.go` | compila |
| 6 | Test de forma-consumidor | `crudview/consumer_test.go` | `gotest` verde |

Secuenciales. La 6 es el gate real. La §9 es posterior al merge y la hace un humano.
