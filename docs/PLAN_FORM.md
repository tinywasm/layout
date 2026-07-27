---
PLAN: "`tinywasm/form`: adoptar el contrato visual — sin `css` viejo, estado sin clases disfrazadas"
---

> Depende de: `github.com/tinywasm/widget v0.1.0` y `github.com/tinywasm/css v0.2.0`
> (**ambos ya publicados**).
> Bloquea a: [`PLAN_COMPONENTS`](PLAN_COMPONENTS.md) (`components/fieldset` skinea las clases
> que `form` emite) y a [`PLAN.md §8`](PLAN.md) (`crudview` compone `form` directamente).
>
> **Ejecución: un solo cambio, no por etapas.** Igual que `ssr`/`components`/`layout`: con
> `widget`/`css` ya publicados no queda ninguna dependencia externa pendiente.

---

## Por qué esto necesita su propio plan, no dos líneas dentro de `PLAN_COMPONENTS.md`

La primera versión de `PLAN_COMPONENTS.md` (§ "Etapa 2 — `fieldset`") describía el cambio en
`form` como pequeño: *"en vez de emitir clases propias para validez y bloqueo, emite
`widget.Invalid` y `widget.Locked`"*. Al revisar el repo real (`github.com/tinywasm/form`,
4023 líneas, HEAD `77d1b60`) resultó ser más de lo que esa frase cubre — hay tres problemas
concretos, no uno:

1. `form/css.go` **ya no compila** contra `css` v0.2.0.
2. El estado "inválido" es una clase disfrazada, exactamente el patrón que el arnés prohíbe.
3. Falta el cableado ARIA que debería haber acompañado esa clase desde el principio.

Cada uno se detalla abajo, con evidencia.

---

## 1. `form/css.go` está roto contra `css` v0.2.0 — no es teórico, es un `go build` que falla

```go
//go:build !wasm
package form

import "github.com/tinywasm/css"

func RenderCSS() *css.Stylesheet {
	return css.NewStylesheet(
		css.Rule(".tw-field",
			css.Display(css.Flex_),
			css.FlexDirection(css.Column),
			css.Gap(css.Px(4)),
			css.Margin(css.Zero, css.Zero, css.Rem(1), css.Zero),
		),
		css.Rule(".tw-field-error",
			css.Display(css.Block),
			css.FontSize(css.TextSm),
			css.Color(css.ColorError),
			css.MinHeight(css.Em(1.2)),
		),
		css.Rule(".tw-field-error--visible",
			css.FontWeight(css.FontWeightMedium),
		),
	)
}
```

`css.Rule`, `css.Display`, `css.FlexDirection`, `css.Gap`, `css.Px`, `css.Margin`, `css.Zero`,
`css.Rem`, `css.Color`, `css.FontSize`, `css.MinHeight`, `css.FontWeight`, `css.Block`,
`css.Column`, `css.Flex_` — **todos** quedaron unexported en el overhaul de `css` v0.2.0
([PLAN_CSS](PLAN_CSS.md) Etapa 5, ya publicado). `go.mod` de `form` todavía pina
`github.com/tinywasm/css v0.1.4`; el día que se bumpee a v0.2.0, este archivo deja de compilar
—no como advertencia, como error de compilación real en todo el árbol que dependa de `form`—.
Es exactamente el tipo de ruptura que ya se documentó al revisar la PR de `css` (`css.Raw`
removido rompiendo `widget/style`): un consumidor real que nadie migró antes de bumpear la
versión.

**El fix es la migración completa, no un parche de compatibilidad**: reescribir `RenderCSS`
como `Style() *style.Sheet` sobre `widget.Name("field")`, usando las primitivas de
`widget/style` en vez del DSL retirado.

---

## 2. El estado "inválido" es una clase disfrazada — el mismo patrón que `PLAN.md` ya diagnosticó en `crudview`

`render_input.go`:

```go
errSpan := dom.NewElement("span").
	ID(fc.Input.ErrorID()).
	Class("tw-field-error").
	Attr("aria-live", "polite").
	BindText(fc.err).
	BindClassFunc("tw-field-error--visible", func() bool {
		return fc.err.Get() != ""
	})
```

`tw-field-error--visible` es un modificador BEM alternado por Go — `PLAN.md` §2.2 ya señaló
exactamente este patrón en `crudview` (`clsBtnCrudIconHidden`): *"un estado disfrazado de
clase"*. `widget.State` existe precisamente para esto:

```go
errSpan.BindAttrFunc(widget.Invalid.Attr().Key, func() string {
	if fc.err.Get() != "" {
		return widget.Invalid.Attr().Value // "true"
	}
	return ""
})
```

y el CSS pasa de `.tw-field-error--visible { font-weight: ...; }` a
`Sheet.When(widget.Invalid, partError, style.FontWeight(style.WeightMedium))` en `@layer
states` — la misma capa donde vive cualquier otro estado del ecosistema, en vez de una
convención BEM local a este archivo.

**`Locked` es un caso distinto y merece su propia decisión, no la misma receta.** Hoy `Locked`
ya se expresa con semántica HTML real, no con una clase:

```go
el.BindAttrBoolFunc("disabled", fc.isDisabledOrLocked)   // radio/select
el.Attr("readonly", "")                                  // texto: legible, no gris
```

Eso ya es correcto y no debe tocarse — es exactamente el motivo por el que `crudview`
documentó que el candado usa `ColorSurface` y no `ColorMuted`: el texto debe seguir siendo
legible bajo "locked". CSS ya tiene `:read-only`/`:disabled` como pseudo-clases nativas para
esto, sin necesitar ningún atributo adicional. La decisión a tomar explícitamente (no a
asumir): ¿`widget/style` debería poder engancharse a `:read-only`/`:disabled` directamente
(un `Cue` más, junto a `Hover`/`Focus`/`Press`/`Target`), o vale la pena además escribir
`data-locked` para que un widget lo pueda leer sin depender de qué atributo HTML nativo usó
`form` por debajo? Recomendación: añadir `Cue.ReadOnly`/`Cue.NativeDisabled` a `widget.Cue`
mapeando a `:read-only`/`:disabled` — reutiliza la semántica nativa en vez de duplicarla con
un `data-*`, y sigue exactamente el principio "reusar lo que ya existe" que motivó cerrar el
DSL viejo de `css`.

---

## 3. Falta el cableado ARIA que la migración debe cerrar de una vez

`ErrorID()` existe (`b.id + ".error"`) y se usa como `id` del `<span>` de error, pero **nunca**
se referencia desde el input vía `aria-describedby`, y el input nunca lleva `aria-invalid`.
Un lector de pantalla hoy no tiene forma de saber que un campo está en error ni de asociarlo a
su mensaje. Esto no es parte de lo que pidió el plan original, pero al tocar exactamente este
archivo para el punto 2, cerrar esto cuesta dos atributos más:

```go
el.Attr("aria-describedby", fc.Input.ErrorID())
el.BindAttrFunc("aria-invalid", func() string {
	if fc.err.Get() != "" {
		return "true"
	}
	return "false"
})
```

Se agrupa aquí y no en un plan de accesibilidad aparte porque es el mismo `BindAttrFunc` que ya
hay que escribir para `widget.Invalid` — el costo marginal es cero si se hace en el mismo
cambio, y no cero si se pospone (hay que volver a tocar el mismo bloque).

---

## El cambio, completo

1. **`go.mod`**: `github.com/tinywasm/css v0.2.0`, añadir `github.com/tinywasm/widget v0.1.0`.
2. **`css.go` → `style.go`**: `RenderCSS() *css.Stylesheet` se reemplaza por
   `Style() *style.Sheet` sobre `widget.Name("field")`, con partes `wrapper`/`label`/`error`,
   usando `Stack`, `Text`, `On(Danger)` donde hoy hay `css.Color(css.ColorError)`, etc.
   `fieldComponent` implementa `widget.Widget` (`WidgetName`/`WidgetKind` →
   `widget.Region` — es un campo, no un widget ARIA con rol propio) y `style.Styler`.
3. **`render_input.go`**: `tw-field-error--visible` → `widget.Invalid.Attr()` vía
   `BindAttrFunc`; añadir `aria-describedby`/`aria-invalid` en el mismo cambio (§3). Las clases
   `tw-field`/`tw-field-error` en sí se conservan como las clases derivadas de
   `widget.Name("field").Class(...)` (anatomía, no estado) — solo el modificador de estado
   desaparece.
4. **`Locked`**: decidir entre `Cue.ReadOnly`/`Cue.NativeDisabled` en `widget` (recomendado,
   §2) o mantenerlo fuera del contrato visual y dejar que cada skin (`components/fieldset`)
   siga estilizando `:read-only`/`:disabled` directamente sin pasar por `widget` en absoluto —
   cualquiera de las dos es aceptable, lo que no es aceptable es introducir un `data-locked`
   que duplique lo que el HTML ya expresa sin que ninguna pieza lo consuma.
5. **No tocar**: `input/*.go` (checkbox, ip, rut, hour, date) — no tienen dependencia de `css`
   ni clases propias; `form.go`'s `SetLocked`/`Focus`/`Validate` — su contrato público no
   cambia, solo cómo se pinta.

---

## Criterios de aceptación

1. `go build ./...` en `form` contra `css v0.2.0` + `widget v0.1.0`, sin `replace` locales.
2. `grep -rn "css\.\(Rule\|Display\|Color\|Margin\|Padding\|Gap\)" --include=*.go .` vacío en
   `form`.
3. Ninguna clase CSS alternada por Go en `render_input.go`; el único mecanismo de estado es
   `widget.State` vía `BindAttrFunc`.
4. El input de un campo en error lleva `aria-invalid="true"` y `aria-describedby` apuntando a
   `ErrorID()`; en cualquier otro estado, `aria-invalid="false"`.
5. `GOOS=js GOARCH=wasm go list -deps ./...` no contiene `tinywasm/css` ni
   `tinywasm/widget/style`.
6. `gotest` en verde en `form` y en `crudview` (consumidor real, vía `components/fieldset`).
