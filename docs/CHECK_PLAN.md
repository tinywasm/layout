# Plan — `github.com/tinywasm/layout`: contrato `UIModule` + tema compartido + distribución de panel

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Gate 2a del [PLAN_TEMA_Y_CAPACIDADES](../../docs/PLAN_TEMA_Y_CAPACIDADES.md). Depende de **Gate 1 (`tinywasm/css`
> con `Theme`) publicado**. Corre en paralelo con `components`. Bloquea a `mjosefa-cms`.
>
> Pieza de lego con **responsabilidad única**: aportar chasis de UI reutilizables (esqueletos
> que definen *dónde* van los elementos, no *qué* contienen). Autocontenido.

---

## Reglas de Desarrollo

Las reglas del arnés (tipado sobre `any`, estados ilegales no representables, reutilizar tipos
ya declarados, superficie mínima, aislamiento WASM/assets, pruebas con `gotest`) viven en el
**`AGENTS.md` de la raíz de esta librería** — léelo antes de cualquier cambio. Este PLAN
describe el *cómo* del cambio.

Alcance de la pieza: esta librería solo describe *estructura visual*; no sabe de datos, de red
ni de dominio.

Este plan tiene **cuatro frentes** en el mismo repo, independientes entre sí:

- **B. Contrato de vista tipado (`UIModule`)** — Cambios 1–3.
- **A. platformd theme-agnostic** — Cambio 4.
- **C. Distribución de rightpanel = módulo pa100t `#device`** — Cambio 5.
- **D. Paridad de geometría del shell con pa100t** — Cambio 6.

> **Fuente de verdad del layout de pa100t:** [IE_LAYOUT.md](IE_LAYOUT.md) — ingeniería inversa del
> shell (`http://192.168.122.10:1100`, ver.3.0.1): estructura HTML, clases, tokens de geometría y
> flujos JS. El **Cambio 6** aplica sus diferencias pendientes (§4, D1–D6). Sus filas de **color**
> (§2) quedan superadas por el Frente A (los colores dejan de ser `--pd-*`).

---

## Estado de partida

La librería expone hoy dos chasis:

- **Chasis de plataforma (`platformd`):** esqueleto raíz con cabecera, barra de navegación y
  escenario. Registra módulos con un **struct público que mezcla tres cosas distintas**:

  ```go
  // Descriptor público actual: mezcla identidad + datos + componentes pre-armados.
  type Module struct {
      ID      string    // identidad — duplica lo que ya da ModelName()
      Label   string    // dato de presentación
      Icon    Component // componente ya renderizado (debería ser un svg.Icon)
      View    Component // componente ya construido (contenido del módulo)
      Default bool      // composición de la app, no del módulo
  }
  ```

- **Chasis de panel de dos columnas (`rightpanel`):** esqueleto con zona principal y panel
  lateral. Toma el id de su elemento raíz de un método `ModelName()`, pero **declara su propia**
  `type Module interface { ModelName() string }`, duplicando un contrato ya existente.

Además `platformd` **inventa una paleta de color propia** (`--pd-color-*`, hex hardcodeados) en
vez de consumir los tokens de `tinywasm/css`, por lo que sus hijos (rightpanel, components) no
respetan su tema.

---

## Cambio 1 — Eliminar la duplicación de identidad  (Frente B)

El chasis de panel **no debe** declarar su propia `Module interface { ModelName() string }`. La
identidad ya está declarada como pieza cero-dep del ecosistema (interfaz externa con
`ModelName() string`). En el código se **importa y se usa**:

```go
// Identidad reutilizada desde la pieza de identidad del ecosistema:
//     type Module interface { ModelName() string }
// El chasis de panel deja de declarar interfaz propia y usa el contrato externo
// para el id de su elemento raíz.
```

Resultado: un único tipo de identidad; los modelos que ya exponen `ModelName()` encajan sin
cambios.

---

## Cambio 2 — El módulo se describe por interfaz; sin descriptor público  (Frente B)

Hoy el host arma a mano un struct descriptor **público** (`ID`, `Label`, `Icon`, `View`,
`Default`). Es un **hueco del arnés** (#3): se puede construir `Module{}` a medias y compila; y
`ID` duplica `ModelName()`. Solución: **eliminar el descriptor** y que el módulo se describa
**por interfaz**. Así no hay dato construible mal, y el compilador exige cada pieza.

```go
// Identidad reutilizada (embebida): type Module interface { ModelName() string }

// UIModule es un módulo que aporta su UI al chasis de plataforma. Lo consume el punto de
// entrada del cliente (wasm). El chasis toma id/hash/ruta de ModelName() y el resto de la
// presentación de estos métodos. No hay struct descriptor público construible a medias.
type UIModule interface {
    Module           // identidad: ModelName() → el chasis lo usa como id
    Label() string   // texto en la barra de navegación
    Icon() svg.Icon  // ícono SIN renderizar; el chasis lo pinta con su clase
    View() Component // contenido del módulo (a menudo un panel rightpanel)
}
```

El chasis compone internamente: `id := m.ModelName(); nav.add(id, m.Label(), m.Icon());
stage.add(id, m.View())`.

**Por qué por interfaz y no por struct:**

- **Estados ilegales no representables (#3).** No existe un `descriptor{}` incompletable. Si un
  módulo no implementa `Icon()`, **no compila** como `UIModule`.
- **Firmas autodescriptivas (#7).** Cada método declara su tipo real; un agente sin contexto
  implementa `Label`/`Icon`/`View` guiado solo por la interfaz.
- **Una sola fuente de identidad.** El id sale de `ModelName()`; imposible que discrepe.

> **`Icon()` devuelve `svg.Icon`, no `Component`** (#1, #2): un ícono de nav es un SVG, no
> "cualquier renderizable", y va **sin renderizar** — el chasis lo pinta con *su* clase
> (`m.Icon().Render(ClsNavIcon)`), cerrando el hueco de "recuerda pasar la clase". `layout` ya
> depende de `tinywasm/svg`. `svg.Icon` es por valor: "sin ícono" es el valor cero (id vacío),
> que el chasis detecta y no pinta. `View()` sí es `Component` (contenido arbitrario).

> **`Default` sale del módulo.** Que un módulo se declare "el default" implica conocer la
> composición de la app. El default es decisión del host → Cambio 3.

**Construcción del módulo (en su propio paquete, no aquí):** constructor con parámetros que
devuelve la **interfaz** sobre un tipo **no exportado**:

```go
// paquete del módulo, fuera de layout
type mod struct{ /* … */ }             // no exportado
func New(deps …) platformd.UIModule {   // función con parámetros; solo puntero tras interfaz
    return &mod{ /* … */ }
}
```

> Cuando un módulo aporta **más de una** capacidad (vista + tools/apis/DB), su `New`/`Load`
> devuelve `[]any` (el bag del Gate 3), del que la vista es **uno** de los elementos tipados como
> `platformd.UIModule`. El caso de arriba (una sola capacidad de vista) devuelve `UIModule` directo.

> **Interacción con el contrato `[]any` de la app (Gate 3).** El módulo de la app expone su
> vista implementando `platformd.UIModule`; ese valor es **uno** de los elementos del `[]any`
> que devuelve `modules/x/init.go`. `web/client.go` recorre el `[]any` y **asierta
> `platformd.UIModule`** para poblar el shell; `web/server.go` asierta `mcp.ToolProvider`. La
> aserción y el registro NO viven en `layout` — solo el tipo `UIModule`.

---

## Cambio 3 — El módulo por defecto es decisión del host  (Frente B)

Se elimina `Default bool` del descriptor. El chasis de plataforma recibe el id por defecto como
ajuste propio: campo **`DefaultID string`** en `Platform`, que el host fija con el `ModelName()`
del módulo a mostrar primero. Si no se fija, el chasis usa el primer módulo. Así "cuál es el
default" queda en el host, no en el contrato del módulo.

### Superficie pública de `Platform` (cómo el host aporta la lista)

`Platform` sigue siendo un **struct** (raíz de composición de la app). Cambia su campo de módulos:
`Modules []Module` → **`Modules []UIModule`**, y se añade `DefaultID string`. **No** hay
constructor aparte (`FromViews` u otro): una sola forma de armar el shell — el struct literal.

```go
p := &platformd.Platform{
    AppName:   "MJosefa CMS",
    Modules:   views,               // []UIModule — cada uno un módulo completo
    DefaultID: someModule.ModelName(),
}
```

Esto **no** reintroduce el hueco #3: lo eliminado fue el *descriptor por módulo* (`Module` con
`ID/Label/Icon/View`), construible a medias. `Platform.Modules []UIModule` guarda **interfaces
completas**, no descriptores con campos a medio llenar; `Platform` como struct de composición es
correcto. `AppName` sigue siendo campo público (como hoy).

---

## Cambio 4 — platformd deja de inventar tokens de color  (Frente A)

`platformd/tokens.go` define una paleta paralela con hex hardcodeados (`--pd-color-primary`
`#ffffff`, `--pd-color-secondary` `#3f88bf`, `--pd-color-gray` `#e9e9e9`, tertiary, quaternary,
selection, hover, success, error). `css.go` las `Declare`. Los hijos leen `--color-*` de
`tinywasm/css` que nadie declara con el tema → no respetan el estilo del shell.

### Tarea

1. **Elimina** de `tokens.go` **todos los tokens de color** `--pd-color-*` y sus `Declare` en
   `css.go`. Sustitúyelos por los símbolos de `tinywasm/css` (`import . "github.com/tinywasm/css"`):

   | Antes (`--pd-color-*`) | Ahora (`tinywasm/css`) |
   |---|---|
   | primary (#ffffff, texto sobre acento) | `ColorOnSecondary` |
   | secondary (#3f88bf, acento/paneles) | `ColorSecondary` |
   | gray (#e9e9e9, fondo shell) | `ColorSurface` / `ColorBackground` según uso |
   | tertiary (#c2c1c1, bordes) | `ColorMuted` |
   | quaternary (#000000, texto) | `ColorOnSurface` |
   | error | `ColorError` |
   | success | `ColorSuccess` |
   | selection/hover (naranjo #ff9300) | activo → `ColorSecondary`; hover → `ColorHover` |

   El naranjo de selección es **decisión de marca**, no de layout: no lo hardcodees. El color
   concreto lo fija la app con `css.Theme(...)` (Gate 3). Hasta entonces se ve el default del
   design-system — correcto.

2. **Conserva** los tokens de **geometría** (no colores) como `--pd-*`: `tokenMenuSize`,
   `tokenHeaderHeight`, `tokenContentHeight`, `tokenContentWidth`, `tokenSlideDur`,
   `tokenTransitionWait`, `tokenFontSize*`. No violan "no hardcodear colores".

3. Documenta que el shell **asume** que existe el catálogo `:root` de tokens (lo aporta un
   `RootCSS()`: el de `tinywasm/css` por defecto, o el del proyecto raíz de la app que lo
   **reemplaza** — el pipeline `assetmin` elige uno solo; ver el plan de css). platformd solo
   **referencia** tokens vía `RenderCSS()`; no declara su propio `:root` de colores.

4. Verificación: `grep -rn "\-\-pd-color" .` → **vacío**. Sin `Hex("#...")` de color en
   `platformd/css.go` (solo quedan valores de geometría).

---

## Cambio 5 — Distribución de rightpanel = módulo pa100t `#device`  (Frente C)

### Objetivo visual

Que un módulo montado en rightpanel se vea con la **misma distribución** que
`http://192.168.122.10:1100/#device` ("Computadores"): banda de **título** arriba + área de
contenido que, para un CRUD, es **lista a la izquierda + detalle (aside) a la derecha**.

### Referencia capturada del shell pa100t (target)

```
section.article-contend-full-page      grid: "title" var(--title-height) / "article" 89vh / 96vw
  div.title-container > div.title > h1  banda de título
  article.all-space-centered            área de contenido (CRUD: lista + aside dentro)
```
```css
.article-contend-full-page { display:grid; grid-template:
    "title"   var(--title-height)
    "article" 89vh / 96vw; }
.module-content { border-radius: 0 .4em .4em 0; border: .1vw solid rgb(194,193,193);
    grid-area: module-content; display:flex; }
```

### Problema con rightpanel actual

`rightpanel/css.go` es un `flex row` con `main` (66vw) + `aside` (30vw) al nivel superior, y
consume tokens de `tinywasm/css` (bien tipado). Pero la **distribución no coincide**: pa100t es
"banda de título + área", con el par lista/aside **dentro** del área, no dos columnas que se
reparten todo el panel.

### Tarea

1. Reproduce el patrón pa100t en `rightpanel/css.go`:
   - Wrapper = **grid** de dos filas: `"title"` (altura `tokenTitleHeight`) + `"content"` (resto,
     `1fr`/`100%`, no `89vh` fijo — debe encajar dentro del `pd-stage`).
   - El área de contenido aloja **main (lista) + aside (detalle)** cuando ambos existen; si solo
     hay `Article`, ocupa todo el ancho (single-column, como `#info` de pa100t).
   - Mantén los tokens de `tinywasm/css` (`ColorSurface`, `ColorMuted`, `RadiusMd`, `Space*`,
     `TextXl`). **Nada** de hex ni tokens nuevos de color.
   - Proporciones main/aside parametrizadas con `tokenMainWidth`/`tokenAsideWidth` (ya existen);
     ajústalas al look pa100t (lista dominante + aside de detalle).
2. `border-radius: 0 .4em .4em 0` del panel lo pone `platformd` (`clsPanelActive`); rightpanel
   **no** debe duplicarlo (evita doble borde).
3. Responsive: conserva el `@media (max-width:640px)` que colapsa a columna.

### Verificación visual (con el MCP de tinywasm)

El daemon del MCP de `tinywasm/app` corre en `:3030` (sin auth). Compara distribución vía
JSON-RPC en `POST http://localhost:3030/mcp` (`tools/call`):

- Actual: `browser_get_content` sobre `http://localhost:6060`.
- Target: navega el devbrowser a `http://192.168.122.10:1100/#device` (si pide login, compara
  contra `#info`, que comparte la distribución `article-contend-full-page`) y usa
  `browser_get_styles` / `browser_get_source`.
- Criterio: banda de título + área con lista/aside, mismos grid-areas y proporciones.

---

## Cambio 6 — Paridad de geometría del shell con pa100t  (Frente D)

Fuente: [IE_LAYOUT.md](IE_LAYOUT.md) §4. Reconcilia los **tokens de geometría** de platformd (los
`--pd-*` que se **conservan** tras el Frente A — NO los de color) con los valores reales de
pa100t. **Verifica el valor actual de cada uno** (IE_LAYOUT está fechado 2026-06-30; algunos ya
pueden estar corregidos) y aplica los pendientes:

| # | Token / regla | Objetivo pa100t | Acción |
|---|---|---|---|
| D1 | `--pd-header-height` | `3vh` | ajustar fallback y el `MediaDesktop` (hoy `5vh`) |
| D2 | `--pd-content-height` | `97vh` (desktop) | ajustar (hoy `94vh`/`95vh`) |
| D3 | `--pd-menu-size` (ancho del rail) | `4vw` | ajustar (hoy `5vw`) |
| D4 | `--pd-content-width` | `96vw` | ajustar (hoy `95vw`) |
| D5 | `border-radius` del panel (`clsPanelActive`) | `0 .4em .4em 0` | verificar; si estuviera `0 .4em 0 0`, corregir a ambos lados derechos |
| D6 | modificadores `.btn-url-up` / `.btn-url-down` (primer/último ítem del rail) | primer/último | agregar como clases opcionales del nav (posicionan el primero arriba y el último abajo) |

Notas:
- **No** re-introducir `--pd-color-*`: los colores ya se resuelven por `tinywasm/css` (Frente A).
  Solo se tocan tokens de **geometría**.
- Las filas ✅ de IE_LAYOUT §4 (D7–D14: grid root con wrapper, activación por clase, `Arvo`,
  `#area`/`#mjs`, etc.) ya coinciden — no requieren acción.
- Verifica con el MCP (`browser_get_styles` sobre `http://localhost:6060`) que el grid del shell
  (`header` + `module-content` + `menu-container`) queda en `96vw / 4vw` y header `3vh`, como el
  reference `http://192.168.122.10:1100`.

---

## Pasos de implementación

1. (B) `go.mod`: dependencia hacia la pieza de identidad del ecosistema.
2. (B) rightpanel: borrar la interfaz `Module` local; usar el contrato de identidad externo.
3. (B) platformd: **eliminar** el struct descriptor público; definir `UIModule` (`ModelName()`
   embebido, `Label()`, `Icon() svg.Icon`, `View() Component`).
4. (B) `Render` de platformd: componer barra y escenario por métodos; ícono con su clase
   (`m.Icon().Render(ClsNavIcon)`, valor cero = sin ícono); `m.View()` en el escenario.
5. (B) Mover el default a `Platform.DefaultID string`; quitar `Default` del contrato.
6. (A) `tokens.go`/`css.go`: borrar tokens de color `--pd-*`; usar `tinywasm/css`.
7. (C) `rightpanel/css.go`: distribución título + contenido lista/aside, solo con tokens.
8. (D) Reconciliar tokens de geometría de platformd con [IE_LAYOUT.md](IE_LAYOUT.md) §4 (D1–D6).
9. Ajustar tests de ambos chasis a los nuevos tipos.

---

## Estrategia de pruebas y criterios de aceptación (`gotest`, no `go test`)

- **Doble objetivo:** compila nativo y `GOOS=js GOARCH=wasm`; assets SSR bajo `//go:build !wasm`.
- **Sin duplicación de identidad:** buscar `interface { … ModelName() string }` dentro de la
  librería no debe encontrar declaración propia; solo uso del tipo externo embebido.
- **Sin descriptor público:** no existe struct exportado con `Label`/`Icon`/`View` públicos; la
  UI se entrega solo por los métodos de `UIModule`.
- **Contrato UI en compilación:** `var _ UIModule = (*fakeModule)(nil)`; un módulo sin `Icon()`
  (o cualquier método) **no compila**.
- **Sin id repetido:** el único origen del id es `ModelName()`; no hay campo `ID`.
- **Default en host:** `Platform.DefaultID` decide; sin `DefaultID`, el primer módulo.
- **Tema:** `grep --pd-color` → vacío; el shell resuelve sus colores desde tokens de
  `tinywasm/css` (verificable aplicando `css.Theme(...)` y comprobando que cambia).
- **Distribución:** con `Article` solo → single-column; con `Article`+`Aside` → dos columnas
  dentro del área; paridad visual con `#device` vía MCP. El toggle de menú sigue parcheando solo
  la clase del nav (sin re-render global).

---

## Documentación (antes de cerrar — obligatoria)

- `platformd/docs/ARCHITECTURE.md`: contrato `UIModule`/`ModelName`/`DefaultID`; platformd
  theme-agnostic consumiendo `tinywasm/css`; `Module` eliminado. Actualiza el diagrama de ciclo
  de vida (`flowchart TD`, sin `subgraph`, `<br/>` para saltos).
- `rightpanel` docs: nueva distribución (título + contenido lista/aside); solo tokens de css.
- `AGENTS.md`: reflejar el contrato `UIModule` como la ÚNICA forma de aportar vista.
- Re-indexa cada `README.md` para enlazar todo `docs/`.

## Etapas

| Etapa | Frente | Entregable | Criterio |
|-------|--------|-----------|----------|
| 1 | B | Identidad dedup + `UIModule` + `DefaultID`; `Module` eliminado | Compila; contrato por interfaz |
| 2 | B | `Render` compone por métodos; ícono con su clase; diagnóstico id vacío/dup | Tests de nav verdes |
| 3 | A | Sin tokens de color `--pd-*`; usa `tinywasm/css` | `grep --pd-color` vacío |
| 4 | C | rightpanel: distribución título+lista/aside vía tokens | Paridad visual MCP |
| 5 | D | Geometría platformd = pa100t ([IE_LAYOUT.md](IE_LAYOUT.md) §4 D1–D6) | grid `96vw`/`4vw`, header `3vh`; paridad MCP |
| 6 | — | Docs + tests dual | `gotest` verde stdlib+WASM |
