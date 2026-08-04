← [Etapa 4](PLAN_STAGE_4_HAMBURGER_SCROLL.md) | [Plan maestro](PLAN.md)

# Etapa 5 — Cambiar de sección desliza; deslizar no cambia de sección

> Lee primero [PLAN.md](PLAN.md): reglas del ecosistema, hot reload y cómo
> verificar en el navegador.

## Síntoma

Deslizando horizontalmente dentro de un artículo, la aplicación termina saltando a
**otra sección**. Estando en el contenido de un artículo, un gesto lleva a edición
o al listado, y otro gesto salta a un módulo distinto: se pierde por completo la
noción de dónde estás.

## Causa

Hay **dos contenedores de scroll-snap horizontal anidados**. Medido en el demo a
375px:

```
.pd__stage   overflow-x:auto  scroll-snap-type:x mandatory  scrollWidth 1200 / clientWidth 400
  └ .rp      overflow-x:auto  scroll-snap-type:x mandatory  scrollWidth  760 / clientWidth 400   ← anidado
```

- `.pd__stage` es el escenario de módulos: `Part("stage", style.Deck(style.SpaceNone), …)`
  en `platformd/css.go`. Tres módulos, tres páginas.
- `.rp` es la tira maestro-detalle de **cada** módulo:
  `On(css.Mobile, "", style.MasterDetail(style.Most), …)` en `rightpanel/css.go`.
  Dos paneles.

Cuando el scroller interior llega a su extremo, el navegador **encadena** el
desplazamiento al exterior y arrastra la aplicación entera al módulo siguiente. Es
comportamiento nativo del navegador, no un fallo de la aplicación: dos scrollers
en el mismo eje siempre se encadenan.

## Decisión

El escenario de módulos **deja de ser un scroller**. Cada sección pasa a ser una
capa que ocupa el contenedor entero y espera aparcada en el borde izquierdo
(`translateX(-100%)`); la que lleva el estado `Current` entra deslizándose de
izquierda a derecha. Es el mismo gesto que el `slider-panel.css` de referencia
resolvía con `margin-left: -104vw` → `0`, hecho con `transform` para que se anime
en la GPU.

Así queda **un solo** contenedor de snap horizontal en la página — el
`MasterDetail` de cada módulo — y deslizar dentro de un artículo ya no puede
cambiar de sección. Para cambiar de sección está el menú, que es lo que se
pretende.

Aplica a móvil **y a escritorio**: el escenario nunca fue un sitio donde deslizar.

## Precondición — plan aguas arriba

Esta etapa **no se puede ejecutar** hasta que esté publicado
<https://github.com/tinywasm/widget/blob/main/docs/PLAN.md> (Etapa W2), que
elimina `style.Deck` y añade:

```go
// widget/style
func SlideDeck(m Motion) Option
```

Verificar antes de empezar:

```bash
grep -n "func SlideDeck" ../widget/style/flow.go   # debe existir
grep -n "func Deck(" ../widget/style/flow.go       # debe estar VACÍO
```

Si `Deck` sigue ahí, **detente**.

## Cambios

### `platformd/css.go`

Sustituir la regla del escenario (líneas 82-90) por:

```go
		// Capas, no tira: los paneles se apilan y el activo entra deslizándose desde
		// el borde inicial. RevealedBy conmuta `display`, que es discreta y no puede
		// transicionar, de ahí que el movimiento venga de un transform.
		//
		// Esto NO es un scroller, y esa es la razón de ser del cambio: cuando lo era,
		// el scroll-snap horizontal de cada módulo (rightpanel MasterDetail, en móvil)
		// encadenaba con este al llegar a su extremo y un gesto dentro del contenido
		// arrastraba la aplicación al módulo siguiente. Un solo eje horizontal
		// desplazable por página.
		//
		// Todos los módulos siguen montados; el estado Current decide cuál se ve.
		Part(widget.Part("stage"),
			style.SlideDeck(style.MotionBase),
			style.Fill(),
		).
```

Y en la regla del panel (líneas 97-101) **quitar `style.Anchor()`**:

```go
		// Scroll(), no Fill(): un módulo más alto que el viewport tiene que
		// desplazarse dentro de su propia capa. Scroll() es Fill() más overflow-y.
		//
		// Sin Anchor(): SlideDeck ya posiciona cada capa en absoluto, lo que la
		// convierte en el bloque contenedor de su contenido — que es lo que el cromo
		// flotante de un módulo necesita para resolverse contra SU panel. Añadir
		// Anchor() aquí lo ROMPE: emite position:relative en @layer widgets, que gana
		// sobre el position:absolute que el flujo emite en @layer primitives, y las
		// capas volverían al flujo apiladas una debajo de otra.
		Part(widget.Part("panel"),
			style.Stack(style.SpaceNone),
			style.Scroll(),
		).
```

### `platformd/platformd.go`

En `Activate`, **borrar** el bloque que empuja el scroller (líneas 472-478):

```go
	// ── BORRAR ──
	// The stage is a Deck: the panels are all mounted side by side and this is
	// what slides between them. `display` is discrete and cannot transition, so
	// the movement has to come from the scroller.
	if el, ok := Get(moduleID); ok {
		el.ScrollIntoView()
	}
```

Ya no hay nada que desplazar: la transición la hace el CSS a partir del estado
`Current` que `p.active` escribe sobre cada `<section>`, con el
`BindStateFunc(widget.Current, …)` que ya existe en `Render`.

Si tras el borrado `Get` deja de usarse en el archivo, quitarlo del dot-import no
procede — `Get` viene del paquete `dom` importado con punto y hay más usos; no
toques los imports salvo que el compilador se queje.

## Lo que NO se toca

- **`rightpanel/css.go` no cambia.** Su `MasterDetail` es justo el scroller que
  debe sobrevivir: es la navegación horizontal *dentro* de un artículo, con su
  snap, y es correcta.
- **`rightpanel.showPanel` no cambia.** Su guardia `strip.ScrollsX()` sigue siendo
  necesaria; lo que cambia es que ahora, además, ya no hay ningún scroller ancestro
  al que un `ScrollIntoView` pudiera escaparse.

## Verificación

```bash
gotest
grep -n "ScrollIntoView" platformd/platformd.go   # → vacío
grep -n "style.Anchor()" platformd/css.go         # → solo en la regla de "menu"
```

En el demo, en **móvil** (`browser_emulate_device` en `mobile`):

1. Ya no hay dos scrollers anidados:
   ```js
   [...document.querySelectorAll('*')]
     .filter(el => getComputedStyle(el).overflowX.match(/auto|scroll/))
     .map(el => el.className)
   ```
   → solo debe aparecer `rp` (y los verticales), **nunca** `pd__stage`.
2. `.pd__stage` es el bloque contenedor y recorta:
   ```js
   (() => { const cs = getComputedStyle(document.querySelector('.pd__stage'));
            return cs.position + " / " + cs.overflow; })()
   ```
   → `"relative / hidden"`.
3. Las capas están aparcadas y solo entra la activa:
   ```js
   [...document.querySelectorAll('.pd__panel')].map(el => ({
     id: el.id,
     cur: el.getAttribute('data-current'),
     t: getComputedStyle(el).transform,
   }))
   ```
   → la sección activa con `data-current="true"` y `transform: matrix(1,0,0,1,0,0)`
   (o `none`); las demás desplazadas en X por su ancho completo.
4. Entrar en `/#devices`, seleccionar una fila para pasar al detalle y **deslizar
   hasta el extremo**: debe quedarse en el detalle. Repetir el gesto: **no** debe
   aparecer otro módulo. Usar `browser_swipe_element` sobre `.rp`.
5. Cambiar de sección desde el menú: la sección nueva debe **entrar deslizándose
   desde la izquierda**, no aparecer de golpe ni llegar desde la derecha.

En **escritorio**: repetir los pasos 1, 2, 3 y 5. El comportamiento es el mismo; lo
que no existe ahí es el `MasterDetail`, porque `rightpanel` vuelve a ser dos
columnas.

## Criterios de aceptación

- `gotest` en verde.
- `.pd__stage` no tiene `overflow-x` ni `scroll-snap-type` en ningún viewport.
- Deslizar dentro de un artículo hasta el extremo **nunca** cambia de sección, ni
  en móvil ni en escritorio.
- Cambiar de sección por el menú produce un deslizamiento de izquierda a derecha.
- El botón flotante de crudview sigue anclado dentro de su propio módulo (comprobar
  que `.crudview__action` sigue en la esquina inferior derecha del panel y no se
  desplaza a otra sección).
