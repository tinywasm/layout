# PLAN: tinywasm/layout — Migración a html/svg/image

## Repositorio
`github.com/tinywasm/layout` — path local: `tinywasm/layout/`

## Dependencias de ejecución
```bash
go install github.com/tinywasm/devflow/cmd/gotest@latest
```

## Prerequisitos (ejecutar ANTES de este plan)
1. `tinywasm/dom` publicado con: `String()` en lugar de `RenderHTML()`, builders eliminados
2. `tinywasm/html` publicado
3. `tinywasm/svg` publicado con `*Sprite`
4. `tinywasm/image` publicado

---

## Objetivo

Migrar todos los layouts de `tinywasm/layout` para:
1. Usar `tinywasm/html` para builders HTML
2. Usar `tinywasm/svg` para íconos (especialmente `platformd`)
3. Renombrar `.RenderHTML()` → `.String()` en tests
4. Agregar `svg.go` en `platformd` con los íconos de la demo registrados correctamente

---

## Actualizar go.mod

```
module github.com/tinywasm/layout

go 1.25

require (
    github.com/tinywasm/css    v<nueva-version>
    github.com/tinywasm/dom    v<nueva-version>
    github.com/tinywasm/html   v<nueva-version>  // NUEVO
    github.com/tinywasm/svg    v<nueva-version>  // NUEVO
    github.com/tinywasm/fmt    v<version-actual>
    github.com/tinywasm/time   v<version-actual>
)
```

---

## Layout 1: `rightpanel/rightpanel.go`

### Migración de imports

**Buscar en `rightpanel/rightpanel.go`:**
```go
import (
    . "github.com/tinywasm/dom"
```

**Reemplazar con:**
```go
import (
    . "github.com/tinywasm/html"
    . "github.com/tinywasm/dom"
```

Verificar qué builders usa el archivo (Div, Section, H2, Aside, etc.) — todos pasan a venir de `html`.

### Tests `rightpanel/rightpanel_test.go`

**Buscar y reemplazar:**
```go
// ANTES:
func (s *stubComponent) RenderHTML() string { return s.html }
html := el.RenderHTML()
html := panel.Render().RenderHTML()

// DESPUÉS:
func (s *stubComponent) String() string { return s.html }
html := el.String()
html := panel.Render().String()
```

Todas las funciones `TestRightPanel_RenderHTML_*` pueden mantener su nombre (es solo el nombre del test, no afecta la firma).

**Verificar:**
```bash
cd tinywasm/layout
gotest -run TestRightPanel
```

---

## Layout 2: `platformd/platformd.go`

### Migración de imports

**Buscar en `platformd/platformd.go`:**
```go
import (
    . "github.com/tinywasm/css"
    . "github.com/tinywasm/dom"
    . "github.com/tinywasm/fmt"
```

**Reemplazar con:**
```go
import (
    . "github.com/tinywasm/css"
    . "github.com/tinywasm/html"  // NUEVO: Div, Span, Nav, Ul, Li, A, Section, Main, etc.
    . "github.com/tinywasm/dom"   // Event, Component, Reference, OnHashChange, GetHash, SetHash
    . "github.com/tinywasm/fmt"
```

Todos los builders en `Render()` (Div, Header, H2, Nav, Ul, Li, A, Span, Section, Main, Div) pasan a venir de `html`.

### Nuevo archivo: `platformd/svg.go`

Crear archivo `platformd/svg.go` (`//go:build !wasm`) con los íconos que usa la demo.

Estos íconos están en el HTML de referencia `pa100tv5/v3-OLD/frontend/built/level_3.html`. Registrar los 3 íconos principales de la demo:

```go
//go:build !wasm

package platformd

import "github.com/tinywasm/svg"

// IconSvg registers the default platform navigation icons.
// Consumers can register their own icons via Module.Icon using svg.Icon().
func (p *Platform) IconSvg() *svg.Sprite {
    return svg.New().
        Add(
            "icon-home",
            `<path fill="currentColor" d="M280.37 148.26L96 300.11V464a16 16 0 0 0 16 16l112.06-.29a16 16 0 0 0 15.92-16V368a16 16 0 0 1 16-16h64a16 16 0 0 1 16 16v95.64a16 16 0 0 0 16 16.05L464 480a16 16 0 0 0 16-16V300L295.67 148.26a12.19 12.19 0 0 0-15.3 0zM571.6 251.47L488 182.56V44.05a12 12 0 0 0-12-12h-56a12 12 0 0 0-12 12v72.61L318.47 43a48 48 0 0 0-61 0L4.34 251.47a12 12 0 0 0-1.6 16.9l25.5 31A12 12 0 0 0 45.15 301l235.22-193.74a12.19 12.19 0 0 1 15.3 0L530.9 301a12 12 0 0 0 16.9-1.6l25.5-31a12 12 0 0 0-1.7-16.93z"/>`,
            "0 0 576 512",
        ).
        Add(
            "icon-products",
            `<path fill="currentColor" d="M350.85 129c25.97 4.67 50.86 13.31 74.16 26.07 7.39 4.07 14.62 8.65 21.76 13.43l-74.53 129.04-21.7-21.7 40.46-40.46-40.97-40.97-40.46 40.46-20.96-20.96 21.21-21.21-40.96-40.96-21.22 21.2-20.64-20.64L269 152.9c-36.49 19.12-59.25 57.79-59.25 99.1v5H160v-5c0-42.63 12.49-84.52 35.46-120.45l40.17 40.17 21.22-21.21 40.96 40.96-21.22 21.22 20.64 20.64 21.22-21.22 40.97 40.97-21.22 21.22 21.7 21.7L296.5 386.4l-12.36-12.36 40.46-40.46-40.97-40.97-40.46 40.46-20.96-20.97 21.21-21.21-40.96-40.96-21.22 21.21-33.67-33.67c-1.71 10.94-2.57 22.07-2.57 33.53v5H96v-5C96 159.99 218.27 37.72 350.85 129z"/>`,
            "0 0 448 512",
        ).
        Add(
            "icon-info",
            `<path fill="currentColor" d="m7 11h2v2h-2zm4-7a4 4 0 1 1 -8 0 4 4 0 0 1 8 0zm-1 0a3 3 0 1 0 -6 0 3 3 0 0 0 6 0zm-3 1h-1v4h1zm0-2h-1v1h1z"/>`,
        )
}
```

### Actualizar `platformd/web/client.go`

El archivo `platformd/web/client.go` (`//go:build wasm`) usa `navIcon()` para construir íconos. Actualizar para usar `svg.Icon()`:

**Imports actuales:**
```go
import (
    . "github.com/tinywasm/dom"
    . "github.com/tinywasm/fmt"
    "github.com/tinywasm/layout/platformd"
    "github.com/tinywasm/layout/rightpanel"
)
```

**Imports nuevos:**
```go
import (
    . "github.com/tinywasm/html"
    . "github.com/tinywasm/svg"
    . "github.com/tinywasm/dom"
    . "github.com/tinywasm/fmt"
    "github.com/tinywasm/layout/platformd"
    "github.com/tinywasm/layout/rightpanel"
)
```

**Reemplazar la función `navIcon`:**
```go
// ELIMINAR:
func navIcon(id string) *Element {
    return Svg(
        Use().Attr("href", "#"+id),
    ).Class("pd-nav-icon").
        Attr("aria-hidden", "true").
        Attr("focusable", "false")
}

// Los módulos usan directamente svg.Icon():
// Icon("icon-home", "pd-nav-icon")
```

**En la definición de módulos, reemplazar:**
```go
// ANTES:
{ID: "mod1", Label: "Módulo 1", Default: true, Icon: navIcon("icon-home"), ...}

// DESPUÉS:
{ID: "mod1", Label: "Módulo 1", Default: true, Icon: Icon("icon-home", "pd-nav-icon"), ...}
```

### Tests `platformd/platformd_test.go`

Buscar y reemplazar:
```go
// ANTES:
html := comp.Render().RenderHTML()

// DESPUÉS:
html := comp.Render().String()
```

**Verificar:**
```bash
cd tinywasm/layout
gotest -run TestPlatform
```

---

## Tests: `platformd/svg_test.go` (nuevo)

```go
//go:build !wasm

package platformd_test

import (
    "strings"
    "testing"
)

func TestPlatform_IconSvg_HasRequiredIcons(t *testing.T) {
    p := &Platform{}
    sprite := p.IconSvg()
    if sprite == nil { t.Fatal("IconSvg() returned nil") }

    m := sprite.Map()
    required := []string{"icon-home", "icon-products", "icon-info"}
    for _, id := range required {
        if _, ok := m[id]; !ok {
            t.Errorf("missing icon: %s", id)
        }
    }
}

func TestPlatform_IconSvg_HasCurrentColor(t *testing.T) {
    p := &Platform{}
    sprite := p.IconSvg()
    s := sprite.String()
    if !strings.Contains(s, "currentColor") {
        t.Error("icons must use fill=currentColor or stroke=currentColor")
    }
}
```

---

## Verificación completa

```bash
cd tinywasm/layout
go build ./...
gotest
```

Verificar en browser que los íconos son visibles:
1. Iniciar el servidor de desarrollo: `cd tinywasm/layout/platformd/web && tinywasm start` (o el comando equivalente del devflow)
2. Abrir el browser en la URL indicada
3. Verificar que los 3 íconos de navegación son visibles en el nav rail
4. Verificar que el panel activo slide correctamente al hacer clic

## Documentación a Actualizar

### `layout/platformd/README.md` (si existe) o crear

Crear/actualizar `tinywasm/layout/platformd/README.md`:
```markdown
# tinywasm/layout/platformd

Shell layout with hash-based routing, nav rail, header, and notifications.

## Usage

    p := &platformd.Platform{
        AppName: "My App",
        Modules: []platformd.Module{
            {
                ID:      "home",
                Label:   "Home",
                Default: true,
                Icon:    svg.Icon("icon-home", "pd-nav-icon"),
                View:    &MyView{},
            },
        },
    }
    dom.Append("body", p)

## Icons

Icons are SVG sprite references. Register them in your consumer's `svg.go`:

    //go:build !wasm
    package main

    import "github.com/tinywasm/svg"
    // or consume Platform's built-in icons (icon-home, icon-products, icon-info)

The platform registers `icon-home`, `icon-products`, `icon-info` by default via `Platform.IconSvg()`.

## CSS Tokens

See `platformd/tokens.go` for the full list of CSS custom properties.
All visual customization should go through these tokens.
```

### `layout/docs/` — crear ARCHITECTURE.md

Crear `tinywasm/layout/docs/ARCHITECTURE.md`:
```markdown
# tinywasm/layout Architecture

## Package Layout

    tinywasm/layout/
    ├── platformd/      # Shell: header, nav rail, hash routing, notifications
    │   ├── platformd.go    # Main struct and Render()
    │   ├── css.go          # !wasm: RenderCSS() *css.Stylesheet
    │   ├── svg.go          # !wasm: IconSvg() *svg.Sprite (built-in nav icons)
    │   ├── tokens.go       # !wasm: CSS token definitions
    │   └── web/            # Demo app (wasm binary entry point)
    └── rightpanel/     # Content panel with header/body/aside slots

## Dependencies

    tinywasm/layout → tinywasm/html (element builders)
    tinywasm/layout → tinywasm/svg  (Icon helper, *Sprite)
    tinywasm/layout → tinywasm/dom  (Component, Event, lifecycle)
    tinywasm/layout → tinywasm/css  (Stylesheet, Token)
```

Ver `tinywasm/docs/MASTER_PLAN.md` para el orden global de ejecución.
