# layout
<img src="docs/img/badges.svg">

Predefined UI layouts for tinywasm modules

## Docs

- [Architecture](docs/ARCHITECTURE.md) — package layout, platform lifecycle, signal fields, activation and notification flows
- [Reverse Engineering](docs/IE_LAYOUT.md) — Reference from pa100t shell
- [Reverse Engineering Module](docs/IE_MODULE_CONTENT.md) — Reference from pa100t crud module
- [Roadmap](docs/ROADMAP.md) — CRUD mobile + desktop design roadmap
- [Visual Contract Master Plan](docs/VISUAL_CONTRACT_MASTER_PLAN.md) — Visual contract definition and migration guidelines
- [Restoration Plan](docs/PLAN.md) — Detailed plan to fix and restore the platformd chassis
- [Widget v0.4 Plan](docs/PLAN_WIDGET_V4.md) — The visual contract widget v0.4 specifications
- [Components Plan](docs/PLAN_COMPONENTS.md) — Layout component plans and guidelines

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
