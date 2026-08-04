← [Etapa 3](PLAN_STAGE_3_MOBILE_TITLE.md) | [Plan maestro](PLAN.md) | Siguiente → [Etapa 5](PLAN_STAGE_5_SECTION_SLIDE.md)

# Etapa 4 — El botón de menú se guarda mientras bajas

> Lee primero [PLAN.md](PLAN.md): reglas del ecosistema, hot reload y cómo
> verificar en el navegador.
>
> **Requiere la [Etapa 3](PLAN_STAGE_3_MOBILE_TITLE.md) terminada.**

## Síntoma

En móvil el botón de menú está fijo en la esquina superior derecha y tapa el
contenido de forma permanente. Tras la Etapa 3, que quitó la reserva de 3rem
superior, se solapa directamente con la barra de búsqueda del listado.

## Decisión

El botón **se guarda al desplazarse hacia abajo y vuelve al desplazarse hacia
arriba**. En reposo — arriba del todo, o en una pantalla que no llega a
desplazarse — está **visible**.

> La petición original era la polaridad contraria (aparecer al bajar). Se
> descartó tras comprobarlo contra el demo: el listado de "Computadores" cabe
> entero en 812px y nunca se desplaza, así que con esa polaridad el menú quedaba
> **inalcanzable** y no había forma de salir del módulo. Con esta polaridad el
> objetivo se cumple igual — el botón desaparece en cuanto empiezas a leer — sin
> que exista un estado sin salida.

La aparición y desaparición son **discretas** (`display`), no animadas. Es como
funciona `RevealedBy` en todo este chasis, incluido el propio cajón de navegación.
Animarlo requeriría una primitiva de transformación que el DSL de estilos no tiene;
queda fuera de alcance.

## Precondición — plan aguas arriba

Esta etapa **no se puede ejecutar** hasta que esté publicado
<https://github.com/tinywasm/dom/blob/main/docs/PLAN.md>, que añade:

```go
// dom
func OnScrollCapture(handler func(scrollTop float64))
```

Hace falta porque **el evento `scroll` no burbujea**: se dispara solo en el
elemento que se desplazó. En este demo, con el móvil emulado, los scrollers
verticales son `.rp__article`, `.crudview__list` y `.targetlist__list` — todos
dentro de `rightpanel`/`crudview`, no de `platformd`. El chasis no puede conocer
esos ids, así que la única forma correcta de observarlos a todos es un listener en
**fase de captura** sobre el documento, que es exactamente lo que aporta esa
función.

Verificar antes de empezar:

```bash
grep -n "OnScrollCapture" ../dom/dom.go ../dom/interface.dom.go   # deben existir
```

Si no está, **detente**.

### `go.mod`

`dom` todavía no tiene `replace` local. Añadirlo junto a los que ya hay:

```
replace github.com/tinywasm/dom => ../dom
```

## Cambios

Todo ocurre en `platformd/platformd.go` y `platformd/css.go`.

### `platformd/platformd.go`

#### Estado

En el bloque `// internal state` del struct `Platform`:

```go
	navStowed *SignalBool
```

Y junto a `rawNotifications`, el testigo de la última posición leída:

```go
	lastScrollTop float64
```

`lastScrollTop` **no va bajo `p.mu`**. Ese mutex existe por `time.AfterFunc`, que
descarta notificaciones desde otra goroutine. Los eventos de scroll llegan por el
hilo de JS, igual que los `click`, y esos ya mutan señales sin candado.

#### Constante

Junto a las demás constantes del paquete:

```go
// scrollStowThreshold son los píxeles de desplazamiento que hacen falta para que
// el cromo reaccione. Sin umbral, un píxel de ruido lo haría entrar y salir; y
// como en la página conviven varios scrollers y sus posiciones se intercalan en un
// mismo handler, un salto pequeño puede venir de otro elemento y no de un gesto.
const scrollStowThreshold = 8
```

#### `Init`

Crear la señal junto a las demás:

```go
	p.navStowed = NewBool(false)
```

Y registrar el listener, **antes** del bloque que resuelve la ruta inicial:

```go
	// El scroll no burbujea: en captura sobre el documento es la única forma de que
	// el chasis vea el desplazamiento de un contenedor que pertenece a otro paquete.
	OnScrollCapture(func(top float64) {
		p.onScroll(top)
	})
```

En SSR esto es un no-op — el stub de backend de `dom` no hace nada — así que el
componente se sigue serializando igual.

#### El manejador

Junto a `isViewable`:

```go
// onScroll guarda el botón de menú mientras el usuario baja y lo devuelve en
// cuanto sube. Arriba del todo siempre está a mano: una pantalla que no llega a
// desplazarse — el listado de este demo, sin ir más lejos — dejaría el menú
// inalcanzable si el botón naciera guardado.
func (p *Platform) onScroll(top float64) {
	if top <= 0 {
		p.lastScrollTop = 0
		p.navStowed.Set(false)
		return
	}
	switch {
	case top > p.lastScrollTop+scrollStowThreshold:
		p.lastScrollTop = top
		p.navStowed.Set(true)
	case top < p.lastScrollTop-scrollStowThreshold:
		p.lastScrollTop = top
		p.navStowed.Set(false)
	}
}
```

#### `Render` — el botón

Añadir el enlace de estado a la construcción que la Etapa 3 dejó:

```go
	hamburger := Button().Set(clsHamburger.AsAttr()).
		Attr("aria-label", "Menu").
		// Open aquí significa "el cromo está desplegado", que es lo mismo que
		// significa en el cajón: un control guardado no está desplegado. La señal
		// dice lo contrario de lo que se pinta, de ahí la negación.
		BindStateFunc(widget.Open, func() bool { return !p.navStowed.Get() }).
		BindChildren(p.navIcon)
	for _, n := range p.navIcon.Get() {
		hamburger.Child(n)
	}
	hamburger.On("click", func(Event) {
		p.menuOpen.Toggle()
	})
	root.Child(hamburger)
```

Como `navStowed` nace en `false`, el marcado inicial que produce SSR ya lleva
`data-open="true"` y el botón se ve desde el primer fotograma.

#### `Activate`

Junto a `p.navIcon.Set(...)` que añadió la Etapa 3:

```go
	// Cambiar de sección reinicia el cromo: el módulo nuevo empieza desde arriba y
	// con el botón a mano.
	p.lastScrollTop = 0
	p.navStowed.Set(false)
```

### `platformd/css.go`

En la regla del botón, añadir `RevealedBy`:

```go
		OnlyOn(css.Mobile, widget.Part("hamburger"),
			style.Row(style.Space1),
			style.As(style.Primary),
			style.Pad(style.Space2),
			style.Round(style.RadiusSm),
			style.Raise(style.Floating),
			style.CenterContent(),
			style.Docked(style.Viewport, style.EdgeTop, style.SideEnd, style.Space4),
			// Se guarda mientras el usuario baja. Es un estado, no una clase: lo
			// escribe Go y lo lee la hoja, y el atributo sale del propio State para
			// que marcado y selector no puedan discrepar.
			style.RevealedBy(widget.Open),
		).
```

Esto emite, dentro del `@media` de móvil, `display: none` en `.pd__hamburger` y
`.pd__hamburger[data-open="true"] { display: flex }`. Es el mismo par que ya
gobierna `.pd__menu`.

## Limitación conocida — dejarla documentada, no intentar arreglarla

El handler recibe la posición de **cualquier** scroller. Si dos están en pantalla y
se desplazan alternándose, sus posiciones se intercalan en un mismo
`lastScrollTop` y el botón puede parpadear. En móvil solo hay una columna visible a
la vez (la tira `MasterDetail` enseña una), así que en la práctica no ocurre, y el
umbral de 8px absorbe el resto. Anotarlo como comentario en `onScroll` y seguir.

## Verificación

```bash
gotest
grep -n "OnScrollCapture" platformd/platformd.go    # una aparición, dentro de Init
```

En el demo, con `browser_emulate_device` en `mobile`:

1. En `/#devices`, en reposo:
   ```js
   document.querySelector('.pd__hamburger').getAttribute('data-open')  // → "true"
   getComputedStyle(document.querySelector('.pd__hamburger')).display  // → "flex"
   ```
2. Provocar un scroll hacia abajo en el listado y comprobar que se guarda:
   ```js
   (() => {
     const s = document.querySelector('.targetlist__list') || document.querySelector('.rp__article');
     s.scrollTop = 200;
     return new Promise(r => setTimeout(() => r(
       document.querySelector('.pd__hamburger').getAttribute('data-open')), 100));
   })()
   ```
   → `null` (el atributo se quitó) y `display` pasa a `none`.

   > Si el listado no desborda porque solo tiene una fila, crea registros con el
   > botón `+` hasta que haya scroll, o usa `browser_swipe_element`.
3. Volver arriba (`s.scrollTop = 0`) → `data-open` vuelve a `"true"`.
4. Navegar a otra sección con el cajón abierto y comprobar que el botón reaparece.
5. En escritorio no cambia nada: el botón sigue con `display: none` fuera del
   `@media` de móvil.

## Criterios de aceptación

- `gotest` en verde.
- En móvil el botón nace visible y sigue visible en una pantalla que no se
  desplaza.
- Bajar lo guarda; subir lo devuelve; volver al tope lo devuelve siempre.
- Cambiar de sección lo devuelve.
- En escritorio no cambia ni un píxel.
