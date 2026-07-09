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
	tinywasm/layout → tinywasm/router (Caller interface)

---

## Platform Lifecycle (`platformd`)

`Platform` implements `Init(ctx dom.Ctx)` + `Render() *dom.Element`. The framework calls `Init` exactly once when the component is first mounted — no `mounted` guard is needed.

```flowchart TD
A[New Platform] --> B[Render: build DOM tree with signal bindings]
B --> C[Init: create signals, register OnHashChange]
C --> D{hash present?}
D -- yes --> E[Activate hash module]
D -- no --> F[Activate DefaultID / first module]
E & F --> G[Runtime — signals drive all UI updates]
G --> H[user clicks hamburger] --> I[menuOpen.Toggle <br/> BindClass patches nav]
G --> J[Notify called] --> K[notifications.Set <br/> BindChildren inserts one row]
G --> L[hash changes] --> M[active.Set <br/> DeriveBool patches panel classes]
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
2. Check `p.CanView(moduleID)` if provided; fallback to default if denied.
3. `p.active.Set(moduleID)` — signals patch nav/panel classes reactively.
4. `p.menuOpen.Set(false)` — closes the mobile overlay.
5. `SetHash("#" + moduleID)` — keeps the URL in sync.

### Notification flow (`Notify` / `dismiss`)

- `Notify`: appends to `rawNotifications`, calls `notifications.Set(buildToasts())` → `BindChildren` inserts one keyed row.
- `dismiss`: removes from `rawNotifications`, calls `notifications.Set(buildToasts())` → `BindChildren` removes one keyed row; untouched rows keep DOM identity.
- `p.mu` protects `rawNotifications` (concurrent goroutines from `time.AfterFunc`).

No `Update()` is called anywhere — all UI changes go through signal `Set`.

---

## CRUD Layout (`crudview`)

Standard two-column layout: Form (left, 66vw) and List (right, 29vw).

### Structure

```flowchart TD
    CV[CrudView] --> L[Left Column: 66vw]
    CV --> R[Right Column: 29vw]
    L --> T[Title: h1]
    L --> F[Form: dom.Component]
    L --> B[CRUD Bar: Buttons]
    R --> LIST[List: SignalNodes]
    R --> S[Search: local filter]
```

### Data Flow (Async Source)

The `Source` seam uses `router.Caller` to fetch data asynchronously.

1. `Init` or `Reload` calls `Caller.Call(op, args, callback)`.
2. Callback receives `[]byte`, decodes it via `Source.Decode`.
3. `allItems` is updated, `filter()` is called.
4. `items` (SignalNodes) is updated, triggering a DOM patch.

### Signal Fields

| Field | Type | Role |
|---|---|---|
| `items` | `*SignalNodes` | Holds the rendered card elements for the list. |
| `selected` | `*SignalString` | Holds the ID of the currently selected item. |
| `search` | `*SignalString` | Holds the current search term; triggers `filter()`. |
| `canSave` | `*SignalBool` | Controls the enabled state of the Save button. |
| `canDelete` | `*SignalBool` | Controls the enabled state of the Delete button. |
