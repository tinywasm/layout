← [Plan maestro](PLAN.md) | Siguiente → [Etapa 2](PLAN_STAGE_2_MODULES.md)

# Etapa 1 — El rail de navegación no debe parpadear al pasar el puntero

> Lee primero [PLAN.md](PLAN.md): reglas del ecosistema, hot reload y cómo
> verificar en el navegador.

## Síntoma

En escritorio, al pasar el puntero por los ítems del menú lateral estos **crecen
ligeramente** y todo el menú parpadea sin parar.

## Causa

`platformd/css.go` declara el hover del ítem así:

```go
Cue(widget.Hover, widget.Part("nav-link"),
	style.As(style.Inset),
).
```

`style.Inset` es una superficie **con borde**. El estado base de `nav-link` no
tiene ninguno. El CSS realmente servido lo confirma:

```css
.pd__nav-link       { padding: var(--space-2); width: 100%; /* sin border */ }
.pd__nav-link:hover { border: 1px solid var(--color-outline); }   /* ← +2px */
```

Medido en el navegador: `.pd__nav-link` mide `55.16 × 56` px con `borderWidth: 0px`
sobre un rail de `--rail-narrow: 3.5rem` = 56px. Al entrar el hover pasa a
`57.16 × 58`: **no cabe**. Además `.pd__nav-link` lleva
`transition: all var(--duration-fast)`, así que el borde se anima y la caja crece
progresivamente durante 150ms.

El bucle es este: el ítem crece → empuja a sus hermanos y ensancha el panel
flotante (`.pd__menu:hover .pd__drawer-panel { width: max-content }`) → el puntero
queda fuera del ítem → se pierde el `:hover` → el ítem encoge → el puntero vuelve
a entrar. Repetido, es el parpadeo.

## Precondición — plan aguas arriba

Esta etapa **no se puede ejecutar** hasta que esté publicado
<https://github.com/tinywasm/widget/blob/main/docs/PLAN.md> (Etapa W1), que hace
que las reglas de `When`/`Cue`/`CueWithin` emitan `outline: 1px solid …` +
`outline-offset: -1px` en lugar de `border`. `outline` se dibuja sobre el mismo
píxel pero **no ocupa espacio de layout**, así que la caja no cambia de tamaño.

Verificar antes de empezar:

```bash
grep -n "overlay" ../widget/style/sheet.go        # debe existir el campo rule.overlay
grep -n "outline-offset" ../widget/style/emit_decls.go   # debe existir
```

Si no están, **detente**: la etapa no aplica todavía.

## Cambios en este repositorio

### `platformd/css.go`

`Cue(widget.Hover, widget.Part("nav-link"), style.As(style.Inset))` **se queda tal
cual**. El arreglo es aguas arriba; aquí solo hay que actualizar el comentario que
lo acompaña, porque el que hay hoy explica una decisión de color y no dice nada de
la geometría. Reemplazar el bloque de comentario de esa regla por:

```go
		// El control entero se ilumina, no solo el glifo: el hover habla del blanco
		// al que apuntas, y el blanco es el botón. Inset, no Accent: la selección ya
		// es ámbar y un hover del mismo color sería indistinguible de ella.
		//
		// El borde de Inset NO ensancha la caja: una regla de estado lo emite como
		// outline (ver widget/style), así que el ítem mide lo mismo con y sin
		// puntero encima. Esa igualdad es obligatoria aquí — el rail mide exactamente
		// --rail-narrow y el panel flotante se dimensiona con width: max-content: dos
		// píxeles de más y el ítem se sale del puntero que lo activó, lo pierde,
		// encoge, y el menú entero entra en un bucle de parpadeo.
		Cue(widget.Hover, widget.Part("nav-link"),
			style.As(style.Inset),
		).
```

### `platformd/consumer_stylesheet_test.go`

Añadir un test con este nombre exacto:

```go
func TestHoverOnANavLinkDoesNotResizeIt(t *testing.T)
```

Genera la hoja con `(&Platform{}).RenderCSS().String()` y verifica sobre el texto:

- La regla `.pd__nav-link:hover` **no** contiene la subcadena `border:`.
- La regla `.pd__nav-link:hover` **sí** contiene `outline:` y `outline-offset: -1px;`.

Este test es la red que impide que el defecto vuelva por un cambio de superficie.

## Verificación en el demo

Con el servidor corriendo y el navegador en escritorio
(`browser_emulate_device` con `mode: "desktop"`):

```js
// browser_evaluate_js
(() => {
  const el = document.querySelector('.pd__nav-link');
  const before = el.getBoundingClientRect();
  const cs = getComputedStyle(el, null);
  return JSON.stringify({w: before.width, h: before.height, border: cs.borderWidth});
})()
```

Debe reportar `border: "0px"` en reposo. Después, comprobar en la hoja servida que
la regla de hover ya no lleva borde:

```bash
curl -s http://localhost:8080/style.css | grep -o '\.pd__nav-link:hover{[^}]*}'
```

El resultado debe contener `outline:1px solid` y **no** `border:1px solid`.

Por último, a ojo: pasar el puntero lentamente por los tres ítems del rail. El
panel debe desplegarse una sola vez y quedarse quieto; ningún ítem debe cambiar de
tamaño.

## Criterios de aceptación

- `gotest` en verde.
- `curl -s http://localhost:8080/style.css | grep -c '\.pd__nav-link:hover{[^}]*border:'` → `0`.
- El rail no parpadea al recorrerlo con el puntero.
- **No** se ha tocado ningún archivo fuera de `platformd/css.go` y
  `platformd/consumer_stylesheet_test.go`.
