# layout
<img src="docs/img/badges.svg">

Predefined UI layouts for tinywasm modules

## Docs

- [Architecture](docs/ARCHITECTURE.md) — package layout, platform lifecycle, signal fields, activation and notification flows
- [Reverse Engineering](docs/IE_LAYOUT.md) — Reference from pa100t shell
- [Reverse Engineering Module](docs/IE_MODULE_CONTENT.md) — Reference from pa100t crud module
- [Implementation Plan](docs/PLAN.md) — Strategy for UIModule and theme parity

## Packages

### platformd

The platform shell, providing the header, navigation rail, and module hosting.

- `NewUIModule(id, label, iconID, view)`: helper to create modules.
- `CanView`: function field to gate module access.

### crudview

A standard two-column CRUD layout (form left, list right) that replicates the Pa100T experience.

- **Preconfigure, don't assemble**: The composition root should wrap `CrudView` once (e.g., in a `config/layouts` package) and feature modules should use that preconfigured version.
- **Async Source**: Uses `router.Caller` for asynchronous data fetching.

```go
view := &crudview.CrudView{
    Title: "Devices",
    Source: crudview.Source{
        Caller: routerCaller,
        ListOp: "list_devices",
        Decode: decodeDevices,
    },
}
```
