# layout
<img src="docs/img/badges.svg">

Predefined UI layouts for tinywasm modules

## Docs

Permanent documentation only — plans and stages live outside the README and are ephemeral.

- [Architecture](docs/ARCHITECTURE.md) — package layout, platform lifecycle, signal fields, activation and notification flows
- [Translation Dictionary](docs/DICTIONARY.md) — how consumers provide a translation dictionary for layout's UI chrome
- [Refactor arquitectónico](docs/ARQ_REFACTOR.md) — por qué `crudview` y `rightpanel` deben fusionarse, evaluado contra el construction harness
- [Reverse Engineering](docs/IE_LAYOUT.md) — Reference from pa100t shell
- [Reverse Engineering Module](docs/IE_MODULE_CONTENT.md) — Reference from pa100t crud module
- [Bug Analysis](docs/BUG_DOM.md) — resolved panic: dom `Show` on second modal open (fixed in dom v0.13.0 / components v0.4.1)
- [Construction Harness](docs/CONSTRUCTION_HARNESS.md) — ecosystem principles: typed, explicit APIs that fail at compile time

## Packages

### platformd

The platform shell, providing the header, navigation rail, and module hosting.
The stage is a slide-deck of layers (not a scroller): the active module slides
in left→right, and a swipe inside a module can never chain onto the stage. On a
phone the hamburger carries the active module's icon and stows while you scroll
down.

- `NewUIModule(id, label, iconID, view)`: helper to create modules.
- `CanView`: function field to gate module access.
- `Platform.Brand` (`BrandName()`/`BrandMark()`): optional leading header slot; empty mark falls back to the shell's default glyph.

Demo modules live one per package under `platformd/modules/<name>/`, each
owning its view, its data and its icon; the chassis ships only its own chrome
glyphs. Adding a module to an application is a package plus one line in
`p.Modules` — no `if` in the shell.

### crudview

A CRUD controller for `rightpanel` (form left, list right) that replicates the Pa100T experience. Renders no frame of its own — it builds a `rightpanel.RightPanel`, fills its slots, and owns only the state machine.

- **Preconfigure, don't assemble**: The composition root should use the high-level constructor `crudview.New(crudview.Config)` to wire the entire form↔list↔transport cycle once per module.
- **Presenter-Based**: Takes a `view.Presenter` that handles list, selection, saving, and deleting.

```go
view, err := crudview.New(crudview.Config{
    ParentID:  "my-module",
    Presenter: myPresenter,
})
```
