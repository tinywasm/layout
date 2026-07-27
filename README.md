# layout
<img src="docs/img/badges.svg">

Predefined UI layouts for tinywasm modules

## Docs

- [Architecture](docs/ARCHITECTURE.md) — package layout, platform lifecycle, signal fields, activation and notification flows
- [Reverse Engineering](docs/IE_LAYOUT.md) — Reference from pa100t shell
- [Reverse Engineering Module](docs/IE_MODULE_CONTENT.md) — Reference from pa100t crud module
- [Roadmap](docs/ROADMAP.md) — CRUD mobile + desktop design roadmap

### Plans (proposals, not yet implemented)

- [PLAN](docs/PLAN.md) — **master plan**: simplified css/widget API, the single widget contract, and why `tinywasm/widget` warrants its own repo
- [PLAN_WIDGET](docs/PLAN_WIDGET.md) — `tinywasm/widget` (new): anatomy, states and layout primitives
- [PLAN_CSS](docs/PLAN_CSS.md) — `tinywasm/css`: complete the token catalog, retire the CSS-mirror DSL
- [PLAN_SSR](docs/PLAN_SSR.md) — `tinywasm/ssr`: typed `Styler` instead of regex name matching
- [PLAN_COMPONENTS](docs/PLAN_COMPONENTS.md) — `tinywasm/components` + `tinywasm/form`: widget migration

## Packages

### platformd

The platform shell, providing the header, navigation rail, and module hosting.

- `NewUIModule(id, label, iconID, view)`: helper to create modules.
- `CanView`: function field to gate module access.

### crudview

A standard two-column CRUD layout (form left, list right) that replicates the Pa100T experience.

- **Preconfigure, don't assemble**: The composition root should use the high-level constructor `crudview.New(crudview.Config)` to wire the entire form↔list↔transport cycle once per module.
- **Presenter-Based**: Takes a `view.Presenter` that handles list, selection, saving, and deleting.

```go
view, err := crudview.New(crudview.Config{
    ParentID:  "my-module",
    Presenter: myPresenter,
})
```
