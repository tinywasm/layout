# tinywasm/layout Architecture

## Package Layout

    tinywasm/layout/
    ├── platformd/      # Shell: header, nav rail, hash routing, notifications
    │   ├── platformd.go    # Main struct, Init(ctx), Render()
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

---

## Platform Lifecycle (`platformd`)

`Platform` implements `Init(ctx dom.Ctx)` + `Render() *dom.Element`. The framework calls `Init` exactly once when the component is first mounted — no `mounted` guard is needed.

```flowchart TD
A[New Platform] --> B[Render: build DOM tree with signal bindings]
B --> C[Init: create signals, register OnHashChange]
C --> D{hash present?}
D -- yes --> E[Activate hash module]
D -- no --> F[Activate default/first module]
E & F --> G[Runtime — signals drive all UI updates]
G --> H[user clicks hamburger] --> I[menuOpen.Toggle → BindClass patches nav]
G --> J[Notify called] --> K[notifications.Set → BindChildren inserts one row]
G --> L[hash changes] --> M[active.Set → DeriveBool patches panel classes]
```

### Signal fields on `Platform`

| Field | Type | Bound to |
|---|---|---|
| `active` | `*SignalString` | nav link `BindClass(clsNavActive, DeriveBool(...))`, panel `BindClass(clsPanelActive, DeriveBool(...))`, header `BindText(DeriveString(...))` |
| `menuOpen` | `*SignalBool` | root `BindClass(clsMenuOpen, ...)`, hamburger `On("click")`, overlay `On("click")` |
| `notifications` | `*SignalNodes` | `msgDesktop` and `msgMobile` containers via `BindChildren` |

### Activation flow (`Activate`)

`Activate(moduleID)` is the single write point for routing:
1. Early-return if `active == moduleID && !menuOpen` (no-op).
2. `p.active.Set(moduleID)` — signals patch nav/panel classes reactively.
3. `p.menuOpen.Set(false)` — closes the mobile overlay.
4. `SetHash("#" + moduleID)` — keeps the URL in sync.

### Notification flow (`Notify` / `dismiss`)

- `Notify`: appends to `rawNotifications`, calls `notifications.Set(buildToasts())` → `BindChildren` inserts one keyed row.
- `dismiss`: removes from `rawNotifications`, calls `notifications.Set(buildToasts())` → `BindChildren` removes one keyed row; untouched rows keep DOM identity.
- `p.mu` protects `rawNotifications` (concurrent goroutines from `time.AfterFunc`).

No `Update()` is called anywhere — all UI changes go through signal `Set`.
