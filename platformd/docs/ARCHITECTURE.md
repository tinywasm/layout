# platformd Architecture

## UIModule Contract

The platform consumes modules via the `UIModule` interface:

```go
type UIModule interface {
    ModelName() string // from layout.Module, used as ID/route
    Label() string     // navigation text
    Icon() svg.Icon    // navigation icon (paints with ClsNavIcon)
    View() Component   // main content
}
```

## Theme Agnostic

`platformd` is theme-agnostic and does not define its own colors. It references semantic tokens from `github.com/tinywasm/css` (`ColorSecondary`, `ColorSurface`, etc.). The actual theme is provided by the root application.

## Routing

- Uses hash-based routing (`#slug`).
- `DefaultID` on `Platform` determines the initial module if no hash is present.
- `Activate(id)` is the single entry point for module switching.
