← [Etapa 2](PLAN_STAGE_2_MODULES.md) | [Plan maestro](PLAN.md) | Siguiente → [Etapa 4](PLAN_STAGE_4_HAMBURGER_SCROLL.md)

# Etapa 3 — En móvil el título se va y el hamburguesa dice dónde estás

> Lee primero [PLAN.md](PLAN.md): reglas del ecosistema, hot reload y cómo
> verificar en el navegador.
>
> **Requiere la [Etapa 2](PLAN_STAGE_2_MODULES.md) terminada**: aquí se lee el
> icono que cada paquete de módulo declara.

## Síntoma

En móvil el título del módulo flota como una chapa azul sobre el contenido y tapa
la barra de búsqueda. Medido en el demo a 375px: `.rp__header` sale
`position: absolute; top: 16px; left: 16px` y se pinta encima de
`.rp__aside-header`.

Además, el único punto que nombra la sección en móvil es esa chapa; el botón de
menú muestra siempre el mismo glifo de hamburguesa y no dice nada de dónde estás.

## Decisión

Dos movimientos que se sostienen mutuamente:

1. **El título desaparece en móvil.** El espacio queda limpio para lo que
   realmente se usa: el formulario y el listado.
2. **El botón de menú pasa a llevar el estado de la navegación**: renderiza el
   icono del módulo activo en vez de un glifo fijo. Deja de ser "abrir menú" a
   secas y pasa a ser "estás aquí, y desde aquí se cambia".

El glifo de hamburguesa (`iconMenu`) **no se borra**: queda como respaldo para el
caso en que el módulo activo no declare icono, que el contrato `UIModule` permite
(`Icon()` puede devolver la cadena vacía).

> **Estado intermedio esperado:** al quitar la reserva de espacio superior, el
> botón flotante queda solapando la esquina superior derecha del contenido. Eso lo
> resuelve la [Etapa 4](PLAN_STAGE_4_HAMBURGER_SCROLL.md), que lo esconde salvo
> cuando hace falta. No añadas relleno para compensarlo: el objetivo de esta etapa
> es despejar ese espacio.

## Cambios

### `rightpanel/css.go`

Hoy el bloque móvil (líneas 95-120) hace tres cosas: encoge el título, convierte la
cabecera en chapa flotante y reserva 3rem arriba en las dos columnas. **Las tres
se van.**

Borrar íntegras estas reglas:

```go
		On(css.Mobile, widget.Part("title"),
			style.FontSize(style.TextBase),
			style.FontWeight(style.WeightBold),
		).
		On(css.Mobile, widget.Part("header"),
			style.Docked(style.Parent, style.EdgeTop, style.SideStart, style.Space4),
			style.Row(style.Space1),
			style.As(style.Primary),
			style.Round(style.RadiusMd),
			style.Pad(style.Space2),
			style.Raise(style.Floating),
			style.Width(style.Content),
		).
		On(css.Mobile, widget.Part("main"),
			style.PadEdge(style.EdgeTop, style.Space12),
		).
		On(css.Mobile, widget.Part("aside"),
			style.PadEdge(style.EdgeTop, style.Space12),
		)
```

Y poner en su lugar, como última regla de la hoja:

```go
		// En móvil no hay cabecera: el chasis lleva el nombre de la sección en su
		// botón de menú, que es el único cromo que sobrevive ahí. Una chapa flotante
		// con el título repetía ese dato encima del contenido y tapaba la barra de
		// búsqueda del aside.
		//
		// Se apaga la cabecera entera y no solo el <h1>: ControlBox le fija una
		// altura mínima de --control-height, así que ocultar únicamente el título
		// dejaría una banda vacía de 50px en lo alto de cada módulo.
		//
		// HeadControls cae con ella. Es aceptable porque hoy no lo usa nadie en el
		// repositorio salvo los tests; un control que deba sobrevivir en móvil va en
		// AsideControls, que es la banda que sí se ve ahí.
		On(css.Mobile, "",
			style.MasterDetail(style.Most),
			style.Pad(style.SpaceNone),
		).
		On(css.Mobile, widget.Part("header"),
			style.Hide(),
		)
```

> **Cuidado:** la regla `On(css.Mobile, "", MasterDetail…)` que ya existe (líneas
> 91-94) **no se toca ni se duplica**. Arriba se muestra para dar contexto de dónde
> encaja la nueva. Si `On(css.Mobile, "")` aparece dos veces las opciones se
> acumulan sobre la misma regla, así que no rompe, pero deja el archivo confuso:
> añade solo el bloque `Part("header")`.

### `platformd/platformd.go`

#### Campo de estado

En el bloque `// internal state` del struct `Platform`, junto a `active`,
`menuOpen` y `notifications`, añadir:

```go
	navIcon *SignalNodes
```

Es `*SignalNodes` y no `*SignalString` a propósito: el icono se pinta con
`icono.Render(clase)`, que es la **única** ruta de render permitida para un
`svg.Icon` en este ecosistema. Enlazar el `href` a mano obligaría a construir el
bloque `<svg><use/></svg>` por fuera de esa ruta.

#### `Init`

Junto a la creación de las demás señales:

```go
	p.navIcon = NewNodes()
```

Debe ir **antes** de la llamada a `Activate`/`fallback` que hay al final de `Init`,
porque `Activate` ya escribe en ella.

#### Un ayudante nuevo, sin exportar

Colocarlo junto a `isViewable`:

```go
// activeIcon es el glifo del módulo en el que estamos. iconMenu es el respaldo:
// UIModule permite que Icon() devuelva la cadena vacía, y un botón sin glifo no
// se puede pulsar porque no se ve.
func (p *Platform) activeIcon() svg.Icon {
	id := p.active.Get()
	for _, m := range p.Modules {
		if m.ModelName() == id {
			if ic := m.Icon(); ic != "" {
				return ic
			}
			break
		}
	}
	return iconMenu
}
```

Sin `map`: un barrido sobre el slice, que es además el orden en el que el chasis ya
recorre los módulos en `Render`.

#### `Render` — el botón

Reemplazar la construcción actual del botón (líneas 291-297) por:

```go
	hamburger := Button().Set(clsHamburger.AsAttr()).
		Attr("aria-label", "Menu").
		BindChildren(p.navIcon)
	// SSR no procesa los bindings de "children": los nodos iniciales se añaden a
	// mano, igual que en la ranura de mensajes de más arriba.
	for _, n := range p.navIcon.Get() {
		hamburger.Child(n)
	}
	hamburger.On("click", func(Event) {
		p.menuOpen.Toggle()
	})
	root.Child(hamburger)
```

`aria-label` sigue diciendo `"Menu"`: el glifo cambió, pero lo que el botón **hace**
no. Se queda como cadena plana — "Menu" es idéntico en todos los idiomas y
envolverlo en `lang.Translate` sería ruido.

#### `Activate`

Después de `p.active.Set(moduleID)` y `p.menuOpen.Set(false)`, añadir:

```go
	// El botón de menú lleva el estado de la navegación: en móvil no hay cabecera
	// ni rail visible, así que su glifo es lo único que dice en qué sección estás.
	p.navIcon.Set([]*Element{p.activeIcon().Render(string(ClsNavIcon))})
```

## Verificación

```bash
gotest

# El título ya no se dibuja como chapa flotante
grep -n "Docked" rightpanel/css.go        # → sin resultados dentro del bloque móvil
```

En el demo, con `browser_emulate_device` en `mobile`:

1. `browser_screenshot` en `/#devices` → la barra de búsqueda debe verse **entera**,
   sin ninguna chapa azul encima. Compárala con
   `platformd/docs/last-acceptable-view.png` solo como referencia de estilo: el
   título ya no debe aparecer.
2. ```js
   // browser_evaluate_js — la cabecera está apagada en móvil
   getComputedStyle(document.querySelector('.rp__header')).display   // → "none"
   ```
3. ```js
   // browser_evaluate_js — el botón lleva el icono del módulo activo
   document.querySelector('.pd__hamburger use').getAttribute('href')
   ```
   En `/#devices` → `"#mod-devices"`. Navegar a `/#home` y repetir → `"#mod-home"`.
   Navegar a `/#about` → `"#mod-about"`.
4. Pulsar el botón sigue abriendo el cajón: `browser_click_element` sobre
   `.pd__hamburger`, luego comprobar
   `document.querySelector('.pd__menu').getAttribute('data-open')` → `"true"`.
5. En escritorio (`browser_emulate_device` en `desktop`) **nada cambia**: la
   cabecera con el título sigue ahí y el botón sigue oculto.

## Criterios de aceptación

- `gotest` en verde.
- En móvil, `.rp__header` tiene `display: none` y ninguna columna reserva 3rem
  arriba.
- El `<use>` dentro de `.pd__hamburger` sigue al módulo activo en las tres rutas.
- El cajón sigue abriéndose al pulsar el botón.
- En escritorio no cambia ni un píxel.
