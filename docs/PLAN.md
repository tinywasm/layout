---
PLAN: "fix: el login usa Page como fondo, no Primary"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 17724461717170660930
---

> Este plan se despacha con el flujo CodeJob. Ver skill: agents-workflow.

# Plan — `tinywasm/layout`: el backdrop del login no debe ser el color de marca

## El problema

`login.Login` pinta **todo el viewport** de la pantalla previa al login con
`--color-primary` sólido: en [`login/css.go`](../login/css.go), su `Root(...)`
declara `style.As(style.Primary)`.

Esto funciona si el color primario de la app es lo bastante apagado como para
cubrir una pantalla entera sin cansar la vista — el color de referencia con el
que se diseñó este patrón es un morado medio (`#654FF0`). No funciona con un
color primario muy saturado (p. ej. un naranja de marca vivo): a pantalla
completa se lee como alarma, no como identidad — el ojo es más sensible a
campos grandes de colores cálidos saturados, que es justo por qué esa familia
de colores se usa para señalización de peligro en dosis pequeñas.

## Por qué el arreglo es aquí y no en la app que lo consume

El propio sistema de tokens de este ecosistema ya resuelve esta distinción en
todos los demás lugares — ver
[`widget/style/surface.go`](https://github.com/tinywasm/widget/blob/main/style/surface.go),
método `Surface.resolve()`:

- `Page`, `Panel`, `Inset` — las tres pintan con `ColorBackground`/
  `ColorSurface` (neutro, adaptado a claro/oscuro), nunca con el color de
  marca.
- `Primary` es la ÚNICA superficie que pinta con `ColorPrimary`, y está
  pensada para un control acotado: `Interactive(Primary)` es exactamente el
  tratamiento de un botón.

`login.Login` es el único lugar de todo el sistema que usa `Primary` como
fondo de una superficie completa en vez de como relleno de un control. No es
una preferencia de una app en particular — es una inconsistencia del propio
widget frente a su propio sistema de tokens, que sólo se nota cuando el color
de marca de quien lo consume es lo bastante intenso como para exponerla.

## El comentario viejo advierte algo que ya no aplica

El comentario que este plan reemplaza dice, en parte: *"Page-on-Page left the
form looking like loose markup rather than a front door"* — una advertencia
real, pero de una configuración que ya no existe. La tarjeta (`PartCard`) usa
`style.As(style.Inset)` — un token distinto de `ColorBackground` (`Page`
resuelve a `ColorSurfaceSunken`, no al mismo `--color-background` del fondo —
ver `widget/style/surface.go`) — más `style.Raise(style.Floating)`, que pinta
una sombra real (`ShadowMd`) independiente del color. Cambiar sólo el `Root`
de `Primary` a `Page` dejaría la tarjeta contra un fondo neutro **distinto**
del suyo propio, con sombra propia: no es la combinación "Page-on-Page" que
el comentario viejo describe, esa habría sido Root Y Card ambos en `Page`.

## El cambio

### 1 — [`login/css.go`](../login/css.go)

Cambia `style.As(style.Primary)` por `style.As(style.Page)` en `Root(...)`:

```go
		Root(
			style.Cover(),
			style.CenterContent(),
			style.Anchor(),
			style.As(style.Page),
			style.Pad(style.Space4),
		).
```

Reemplaza el comentario que hoy explica `Primary` (el bloque que empieza en
`// Primary is the whole point of the screen...`, líneas 26-29) por:

```go
		// Page, not Primary: a full-bleed wash of the brand color only reads
		// well when that color happens to be muted enough to cover a whole
		// viewport — which is not guaranteed, and the rest of this token
		// system already treats Primary as the paint for a bounded control (a
		// button), never a surface (see widget/style's own Surface.resolve()).
		// A card only reads as elevated against something it is clearly ON;
		// Page's own neutral tone against Panel/Inset already does that job
		// without betting the whole screen on one brand color's saturation.
```

No toques nada más de la función: `Cover`, `CenterContent`, `Anchor`, `Pad` y
las demás `Part(...)` (`PartCard`, `PartHeader`, `PartTitle`, `PartSubtitle`,
`PartMark`) siguen exactamente igual.

### 2 — [`login/login.go`](../login/login.go)

El comentario de tipo sobre `Login` (líneas 31-42) describe el backdrop viejo.
Reemplázalo por:

```go
// Login is the pre-authentication screen: an elevated card (title, subtitle,
// form) centered on the app's own page background, with an optional corner
// mark pinned independently of that card — the reference this productionizes
// (a legacy pa100t deployment) keeps its own crest bottom-left regardless of
// how tall the form above it grows, which a mark living inside the card would
// not survive.
//
// The backdrop is Page — the same neutral surface every other screen sits
// on — not Primary: a brand color saturated enough to read well on a button
// rarely survives being stretched across an entire viewport, and the token
// system already reserves Primary for bounded controls (see widget/style's
// own Surface.resolve()). The card itself, plus an optional LogoMark, carry
// the brand instead of the backdrop.
```

No cambies nada del resto del struct (`Title`, `Subtitle`, `Form`, `LogoMark`
y sus propios comentarios de campo quedan igual).

## Test

En [`login/login_test.go`](../login/login_test.go), agrega (mismo paquete
`login`, no `login_test` — ya es lo que usa este archivo):

```go
func TestLogin_RootUsesPageNotPrimary(t *testing.T) {
	css := (&Login{Title: "App"}).RenderCSS().String()

	if strings.Contains(css, "--color-primary") {
		t.Errorf("login root must not paint the brand color as its backdrop, got:\n%s", css)
	}
	if !strings.Contains(css, "--color-background") {
		t.Errorf("login root must use the neutral page background, got:\n%s", css)
	}
}
```

Este archivo hoy sólo importa `strings`, `testing` y `. "github.com/tinywasm/html"`
— `RenderCSS()` es un método de `*Login` ya visible en el propio paquete, no
hace falta importar nada nuevo.

## Criterios de aceptación

- [ ] `login/css.go`: `Root(...)` usa `style.As(style.Page)`.
- [ ] Los dos comentarios (el de `Root` en `css.go` y el del tipo `Login` en
      `login.go`) están reemplazados exactamente como arriba — no queda texto
      que siga describiendo un backdrop de marca.
- [ ] `TestLogin_RootUsesPageNotPrimary` pasa.
- [ ] `go test ./...` en verde, incluida toda la suite existente sin
      modificar — en particular los cinco tests ya existentes en
      `login/login_test.go`, que sólo comprueban marcado HTML y no deberían
      verse afectados.
- [ ] `grep -n "style.Primary" login/css.go` → vacío.

## Fuera de alcance

No se toca ningún otro widget de este repo (`crudview`, `rightpanel`,
`platformd`, `landing`, …) ni la definición de `Surface`/`style.Page`/
`style.Primary` en `tinywasm/widget` — esos ya están bien; el defecto era sólo
la elección que hace `login.Login`. No se agrega una opción para que una app
elija entre Page y Primary: el patrón correcto es uno solo, no una
configuración.
