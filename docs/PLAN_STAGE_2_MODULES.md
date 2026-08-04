← [Etapa 1](PLAN_STAGE_1_HOVER_RAIL.md) | [Plan maestro](PLAN.md) | Siguiente → [Etapa 3](PLAN_STAGE_3_MOBILE_TITLE.md)

# Etapa 2 — Cada módulo demo en su propio paquete

> Lee primero [PLAN.md](PLAN.md): reglas del ecosistema, hot reload y cómo
> verificar en el navegador.

## Objetivo

`platformd/web/client.go` tiene hoy 272 líneas donde conviven el modelo de datos,
un backend CRUD en memoria, el adaptador de router, la vista del módulo, la
identidad de la marca y el cableado de la aplicación. Es un único tipo `mod` con
un `if m.name == "crud"` decidiendo qué renderizar, así que no hay forma de añadir
un cuarto módulo sin ensanchar ese `if`.

Se extraen los **tres módulos actuales** a paquetes bajo
`platformd/modules/<nombre>/`, cada uno dueño de su vista, sus datos **y su
icono**. `client.go` queda solo con el cableado.

No se añade ningún módulo nuevo en esta etapa. No cambia nada visible en la
aplicación: es una refactorización pura y ese es el criterio de aceptación
principal — la captura de pantalla antes y después debe ser idéntica.

## Estructura destino

```
platformd/
├── platformd.go
├── css.go
├── svg.go
├── factory.go
├── modules/
│   ├── devices/
│   │   ├── devices.go     ← Module, New, ModelName/Label/Icon/View, const Icon
│   │   ├── model.go       ← Device, deviceList, deviceDef
│   │   ├── store.go       ← deviceDB, newSeededDeviceDB, memCaller
│   │   └── svg.go         ← //go:build !wasm — IconSvg()
│   ├── home/
│   │   ├── home.go
│   │   └── svg.go
│   └── about/
│       ├── about.go
│       └── svg.go
└── web/
    └── client.go          ← solo cableado
```

Correspondencia con lo que hay hoy:

| Hoy en `client.go` | Va a |
|---|---|
| `mod{"crud", "Computadores", IconProducts, p}` + rama `if m.name == "crud"` | `modules/devices` |
| `mod{"mod1", "Inicio", IconHome, p}` | `modules/home` |
| `mod{"mod2", "Acerca de", IconInfo, p}` | `modules/about` |
| `mod{"hidden", "Oculto", IconInfo, p}` | se queda en `client.go` (ver abajo) |
| `deviceDef`, `Device`, `DeviceList` | `modules/devices/model.go` |
| `deviceDB`, `newSeededDeviceDB`, `memCaller` | `modules/devices/store.go` |
| `demoBrand`, `demoIdentity`, `main()` | se quedan en `client.go` |
| el tipo `mod` | **se borra** |

## Los iconos se mudan con su módulo

Hoy `platformd` exporta un catálogo fijo de iconos de navegación
(`IconHome`, `IconProducts`, `IconInfo`) y los dibuja en su propio `svg.go`. Eso
es una decisión de contenido dentro del chasis: la plataforma no tiene por qué
saber que existe algo llamado "home".

**Se eliminan de `platformd`** y cada módulo declara y dibuja el suyo.

`platformd` **conserva** `IconUser`, `IconBrand` e `iconMenu`: son cromo del
chasis (avatar por defecto, marca por defecto, botón de menú), no contenido de
ningún módulo.

## Archivos nuevos

Todos los fragmentos de abajo son el contenido íntegro del archivo salvo donde se
indique. El código movido se copia **tal cual**, incluidos sus comentarios: no
reescribas explicaciones que ya son correctas.

El bloque `import` de cada archivo nuevo se arma con lo que **ese archivo** usa, no
copiando el de `client.go` entero: `model.go` necesita `model`, `view` y
`form/input`; `store.go` necesita `orm`, `storage`, `storage/mem`, `model` y el
punto-import de `fmt` (por `Errf`); `svg.go` solo necesita `svg/sprite`.

### `platformd/modules/devices/devices.go`

Sin build tag. Declara el icono (referencia compartida) y el módulo.

```go
// Package devices es el módulo demo de un CRUD completo: modelo real, backend en
// memoria y crudview montado sobre él. Es el patrón a copiar para un módulo que
// administra registros.
package devices

import (
	"github.com/tinywasm/components/searchbar"
	"github.com/tinywasm/layout/crudview"
	"github.com/tinywasm/layout/platformd"
	"github.com/tinywasm/model"
	"github.com/tinywasm/svg"
	"github.com/tinywasm/view"

	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
)

// Icon es la referencia compartida al glifo del módulo. El dibujo vive en svg.go,
// que es backend puro. El prefijo mod- evita chocar con los ids del sprite del
// chasis dentro del sprite de página, que es la fusión de todos ellos.
const Icon = svg.Icon("mod-devices")

// Module es el módulo de equipos. Implementa platformd.UIModule.
type Module struct {
	p *platformd.Platform
}

// New construye el módulo. Recibe la plataforma porque este módulo notifica el
// resultado de sus mutaciones en la barra de mensajes del chasis.
func New(p *platformd.Platform) *Module { return &Module{p: p} }

var _ platformd.UIModule = (*Module)(nil)

func (m *Module) ModelName() string { return "devices" }
func (m *Module) Label() string     { return "Computadores" }
func (m *Module) Icon() svg.Icon    { return Icon }

func (m *Module) View() Component {
	pres := view.New(&memCaller{db: deviceDB}, &Device{}, "device_list",
		func() model.ModelSlice { return &deviceList{} },
		view.WithTitle("Computadores"),
		view.WithSaveOp("device_save"),
		view.WithDeleteOp("device_delete"),
	)
	cv, err := crudview.New(crudview.Config{
		ParentID:  m.ModelName(),
		Presenter: pres,
		// El filtro lo elige la aplicación, no el layout. Esta es la línea que un
		// despliegue cambia por un calendario o un select de categoría; ni crudview
		// ni rightpanel se enteran.
		Filter: &searchbar.SearchBar{Placeholder: "Buscar..."},
	})
	if err != nil {
		panic(err)
	}
	// Sin notificación al seleccionar: elegir una fila es navegar, no actuar — un
	// aviso ahí no aporta nada y en móvil, donde el módulo llena la pantalla, el
	// mensaje se superpone a la vista entera. Guardar y borrar SÍ notifican, porque
	// confirman una mutación real.
	cv.OnNew = func() { m.p.Notify(Msg.Info, "Nuevo", 2000) }
	cv.OnSaved = func(err error) {
		if err == nil {
			m.p.Notify(Msg.Success, "Guardado", 2000)
		}
	}
	cv.OnDeleted = func(id string, err error) {
		if err == nil {
			m.p.Notify(Msg.Error, "Eliminado "+id, 2000)
		}
	}
	cv.OnCancel = func() { m.p.Notify(Msg.Info, "Cancelado", 2000) }
	return cv
}
```

> **Ojo con el id:** `ModelName()` pasa de `"crud"` a `"devices"`. `crudview.Config.ParentID`
> **debe** seguir siendo el mismo valor que `ModelName()` — de ahí que se pase
> `m.ModelName()` en vez de un literal. La sección del escenario y el panel usan ese
> id para resolverse; dos literales distintos rompen la navegación en silencio.
> Consecuencia: la URL del módulo pasa de `#crud` a `#devices`.

### `platformd/modules/devices/model.go`

Sin build tag. Contenido: `deviceDef`, `Device` con todos sus métodos, y
`DeviceList` **renombrado a `deviceList`** (solo lo usa este paquete), copiados
literalmente de las líneas 40-98 del `client.go` actual, incluidos sus comentarios.
Las aserciones de compilación se conservan:

```go
var _ model.Model = (*Device)(nil)
var _ view.Itemizer = (*Device)(nil)
var _ model.ModelSlice = (*deviceList)(nil)
```

### `platformd/modules/devices/store.go`

Sin build tag. Contenido: `deviceDB`, `newSeededDeviceDB()` y `memCaller` con sus
métodos `Call` y `Dispatch`, copiados literalmente de las líneas 100-165 del
`client.go` actual, incluidos sus comentarios. Todas las referencias a
`DeviceList` pasan a `deviceList`.

### `platformd/modules/devices/svg.go`

```go
//go:build !wasm

package devices

import "github.com/tinywasm/svg/sprite"

// IconSvg registra el glifo del módulo. tinywasm/ssr fusiona el resultado de cada
// IconSvg() del grafo en un único sprite inyectado en <body>.
//
// El receptor se instancia como valor cero (`&devices.Module{}`), así que este
// método no puede tocar ningún campo: solo devuelve definiciones.
func (m *Module) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		sprite.Define(Icon, "0 0 448 512", sprite.Path("M350.85 129c25.97 4.67 47.27 18.67 63.92 42 14.65 20.67 24.64 46.67 29.96 78 4.67 28.67 4.32 57.33-1 86-7.99 47.33-23.97 87-47.94 119-28.64 38.67-64.59 58-107.87 58-10.66 0-22.3-3.33-34.96-10-8.66-5.33-18.31-8-28.97-8s-20.3 2.67-28.97 8c-12.66 6.67-24.3 10-34.96 10-43.28 0-79.23-19.33-107.87-58-23.97-32-39.95-71.67-47.94-119-5.32-28.67-5.67-57.33-1-86 5.32-31.33 15.31-57.33 29.96-78 16.65-23.33 37.95-37.33 63.92-42 15.98-2.67 37.95-.33 65.92 7 23.97 6.67 44.28 14.67 60.93 24 16.65-9.33 36.96-17.33 60.93-24 27.98-7.33 49.96-9.67 65.94-7zm-54.94-41c-9.32 8.67-21.65 15-36.96 19-10.66 3.33-22.3 5-34.96 5l-14.98-1c-1.33-9.33-1.33-20 0-32 2.67-24 10.32-42.33 22.97-55 9.32-8.67 21.65-15 36.96-19 10.66-3.33 22.3-5 34.96-5l14.98 1 1 15c0 12.67-1.67 24.33-4.99 35-3.99 15.33-10.31 27.67-18.98 37z")),
	)
}
```

El `viewBox` y el `d` son exactamente los que hoy tiene `IconProducts` en
`platformd/svg.go`.

### `platformd/modules/home/home.go`

```go
// Package home es el módulo demo más simple: un rightpanel con un artículo y nada
// más. Es el patrón a copiar para una pantalla de contenido estático.
package home

import (
	"github.com/tinywasm/layout/platformd"
	"github.com/tinywasm/layout/rightpanel"
	"github.com/tinywasm/svg"

	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/html"
)

const Icon = svg.Icon("mod-home")

type Module struct{}

func New() *Module { return &Module{} }

var _ platformd.UIModule = (*Module)(nil)

func (m *Module) ModelName() string { return "home" }
func (m *Module) Label() string     { return "Inicio" }
func (m *Module) Icon() svg.Icon    { return Icon }

// Sin Module: dentro de platformd la SECCIÓN del escenario es la dueña del id de
// ruta, y RightPanel estampa Module.ModelName() en su propia raíz — dos nodos con
// un mismo id. Title lleva la etiqueta, que es todo lo que este panel necesita.
func (m *Module) View() Component {
	return &rightpanel.RightPanel{
		Title:   m.Label(),
		Article: Div().Text("Contenido de " + m.Label()),
	}
}
```

### `platformd/modules/home/svg.go`

Igual que el de `devices`, con `viewBox` `"0 0 576 512"` y el `d` que hoy tiene
`IconHome` en `platformd/svg.go`.

### `platformd/modules/about/about.go`

Copia de `home.go` con `ModelName() = "about"`, `Label() = "Acerca de"` y
`const Icon = svg.Icon("mod-about")`.

### `platformd/modules/about/svg.go`

Igual, con `viewBox` `"0 0 16 16"` y el `d` que hoy tiene `IconInfo` en
`platformd/svg.go`.

## Archivos modificados

### `platformd/svg.go`

Borrar las tres líneas `sprite.Define(IconHome, …)`, `sprite.Define(IconProducts, …)`
y `sprite.Define(IconInfo, …)`. Quedan solo `IconUser`, `IconBrand` e `iconMenu`.
Ajustar el comentario de cabecera para que no prometa "iconos de navegación".

### `platformd/platformd.go`

En el bloque `const` de iconos (líneas 48-55), **borrar** `IconHome`,
`IconProducts` e `IconInfo`. Queda:

```go
const (
	IconUser  = svg.Icon("pd-user")
	IconBrand = svg.Icon("pd-brand")
	iconMenu  = svg.Icon("pd-menu")
)
```

Ningún otro cambio en este archivo.

### `platformd/web/client.go`

Queda con: los imports, `demoBrand`, `demoIdentity`, el módulo oculto y `main()`.
El tipo `mod` **se borra**.

El módulo `hidden` no se muda a un paquete: no es un módulo, es la prueba de que
`CanView` filtra. Se queda aquí como lo que es, un fixture, y reutiliza el icono de
`about` en vez de declarar uno propio:

```go
// hiddenModule existe solo para probar CanView: se registra y el chasis no debe
// pintarlo ni en el rail ni en el escenario. No tiene paquete propio porque no es
// un módulo demo — es un caso de permisos.
type hiddenModule struct{}

func (hiddenModule) ModelName() string { return "hidden" }
func (hiddenModule) Label() string     { return "Oculto" }
func (hiddenModule) Icon() svg.Icon    { return about.Icon }
func (hiddenModule) View() Component   { return nil }
```

Y el registro en `main()`:

```go
	p.DefaultID = "devices"   // ← el campo ya existe; cambia el valor de "crud" a "devices"

	p.Modules = []platformd.UIModule{
		devices.New(p),
		home.New(),
		about.New(),
		hiddenModule{},
	}
```

Imports que **se van** de `client.go`: `crudview`, `searchbar`, `input`, `model`,
`orm`, `storage`, `storage/mem`, `rightpanel`, `view`. Imports que **entran**: los
tres paquetes de módulo. Se conservan `dom`, `fmt`, `html`, `svg`, `platformd`,
`themetoggle` y el import en blanco de `fieldset`.

> El import en blanco `_ "github.com/tinywasm/components/fieldset"` **se queda en
> `client.go`**. Es la piel global de formularios de la aplicación: pertenece a la
> raíz de composición, no a un módulo. Moverlo a `devices` haría que la piel
> apareciera y desapareciera según qué módulos registre una aplicación.

## Verificación

```bash
# Ningún icono de contenido sigue en el chasis
grep -rn "IconHome\|IconProducts\|IconInfo" platformd/     # → vacío

# El tipo mod desapareció
grep -rn "type mod struct" platformd/web/                  # → vacío

# Los paquetes de módulo NO arrastran el sprite al binario del navegador
GOOS=js GOARCH=wasm go list -deps ./platformd/modules/devices | grep tinywasm/svg/sprite  # → vacío
GOOS=js GOARCH=wasm go list -deps ./platformd/modules/home    | grep tinywasm/svg/sprite  # → vacío
GOOS=js GOARCH=wasm go list -deps ./platformd/modules/about   | grep tinywasm/svg/sprite  # → vacío

gotest
```

En el demo (recuerda: **no compiles a mano**, el hot reload ya recargó):

1. `browser_navigate` a `/#devices` → se ve el CRUD "Computadores" con la lista y
   el formulario, exactamente igual que antes.
2. `browser_navigate` a `/#home` y `/#about` → sus paneles.
3. `browser_navigate` a `/#hidden` → **no** se activa; cae al módulo por defecto.
4. `browser_evaluate_js` con
   `[...document.querySelectorAll('.pd__nav-link use')].map(u => u.getAttribute('href'))`
   → debe devolver `["#mod-devices", "#mod-home", "#mod-about"]`. Si alguno sale
   vacío o el icono se ve gigante y negro, es que su `IconSvg()` no fue recogido:
   revisa el nombre exacto del método y la regla de un solo receptor por paquete.

## Criterios de aceptación

- `gotest` en verde.
- Los tres greps de arriba devuelven vacío.
- La aplicación se ve **idéntica** a antes en escritorio y en móvil; lo único que
  cambia son los ids de ruta (`#crud` → `#devices`, `#mod1` → `#home`,
  `#mod2` → `#about`).
- `platformd/web/client.go` baja de 272 líneas a menos de 90.
- Añadir un cuarto módulo demo consiste en crear un paquete y una línea en
  `p.Modules`, sin tocar ningún `if`.
