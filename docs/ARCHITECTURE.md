# tinywasm/layout Architecture

## Package Layout

    tinywasm/layout/
    ├── platformd/      # Shell: header, nav rail, hash routing, notifications
    ├── rightpanel/     # THE module skeleton: frame, two columns, aside bands,
    │                   # mobile master-detail strip. Owns every layout primitive.
    └── crudview/       # CRUD controller: state machine + orchestration.
                        # Renders NO frame — composes rightpanel.

## Who owns what

The module frame has exactly one owner: `rightpanel`. It owns every layout
primitive — the split, the two columns, the aside bands, the mobile
master-detail strip — and every module in this repository composes it.
`crudview` renders NO frame: it is a CRUD controller that builds a
`rightpanel.RightPanel`, fills its slots (Title, Article, Aside, AsideControls,
AsideFooter) and keeps only the state machine. `platformd` is the shell
(routing, chrome) and is a third thing.

Three tests in `conformance_test.go` make the split un-reintroducible:

- `TestOnlyOneOwnerOfTheGrid` fails if a second package emits
  `Split`/`MasterDetail` — a second skeleton is how the duplication started.
- `TestNoLocallyDeclaredSeams` rejects a locally-declared `Filterable`-style
  interface that forks an upstream contract.
- `TestEverySlotIsRendered` fails when a slot exists on `RightPanel` but is
  never wired into `Render()`.

See [ARQ_REFACTOR.md](ARQ_REFACTOR.md) for why the two skeletons diverged and
what the construction harness has to say about it.

## Dependencies

    tinywasm/layout → tinywasm/html (element builders)
    tinywasm/layout → tinywasm/svg  (Icon helper, *Sprite)
    tinywasm/layout → tinywasm/dom  (Component, Event, lifecycle)
    tinywasm/layout → tinywasm/css  (Stylesheet, Token)
	tinywasm/layout → tinywasm/router (Caller interface)

---

## Platform Lifecycle (`platformd`)

`Platform` implements `Init(ctx dom.Ctx)` + `Render() *dom.Element`. The framework calls `Init` exactly once when the component is first mounted — no `mounted` guard is needed.

The application chassis is built exclusively from the `Cover` and `Sidebar` layouts using the style DSL. Sizing, grid flow, mobile reflow, and drawer navigation are defined by widget-level tokens and layout models. Route selection and visibility are carried exclusively by reactive state attributes (`data-current`, `data-open`) tied to `widget.Current` and `widget.Open` states, completely eliminating legacy CSS class toggles.

`Cover` fixes the shell to the viewport height, so the frame itself never scrolls; the active module panel carries `Scroll` and is the only element that does. Any part rendering a bare `<svg>` declares `IconBox` — without a box an svg falls back to 300×150 and breaks the layout. Both require `widget` v0.4.4 or later.

```flowchart TD
A[New Platform] --> B[Render: build DOM tree with signal bindings]
B --> C[Init: create signals, register OnHashChange]
C --> D{hash present?}
D -- yes --> E[Activate hash module]
D -- no --> F[Activate DefaultID / first module]
E & F --> G[Runtime — signals drive all UI updates]
G --> H[user clicks hamburger] --> I[menuOpen.Toggle <br/> BindAttrBool data-open patches UI]
G --> J[Notify called] --> K[notifications.Set <br/> BindChildren inserts toast row]
G --> L[hash changes] --> M[active.Set <br/> DeriveBool patches panel and link data-current]
```

### Signal fields on `Platform`

| Field | Type | Bound to |
|---|---|---|
| `active` | `*SignalString` | nav link `BindAttrBool("data-current", DeriveBool(...))`, panel `BindAttrBool("data-current", DeriveBool(...))`, header `BindText(DeriveString(...))` |
| `menuOpen` | `*SignalBool` | nav overlay `BindAttrBool("data-open", ...)`, navigation rail `BindAttrBool("data-open", ...)` |
| `notifications` | `*SignalNodes` | `msgSlot` container via `BindChildren` |

### Activation flow (`Activate`)

`Activate(moduleID)` is the single write point for routing:
1. Early-return if `active == moduleID && !menuOpen` (no-op).
2. Check `p.CanView(moduleID)` if provided; fallback to default if denied.
3. `p.active.Set(moduleID)` — signals patch nav/panel state attributes reactively.
4. `p.menuOpen.Set(false)` — closes the mobile overlay.
5. `SetHash("#" + moduleID)` — keeps the URL in sync.

### Notification flow (`Notify` / `dismiss`)

- `Notify`: appends to `rawNotifications`, calls `notifications.Set(buildToasts())` → `BindChildren` inserts one keyed row.
- `dismiss`: removes from `rawNotifications`, calls `notifications.Set(buildToasts())` → `BindChildren` removes one keyed row; untouched rows keep DOM identity.
- `p.mu` protects `rawNotifications` (concurrent goroutines from `time.AfterFunc`).

No `Update()` is called anywhere — all UI changes go through signal `Set`.

---

## Header chrome (`platformd`, v0.2.0 pass)

The header is a three-slot frame: **brand** at the leading edge, **messages**
centred (`CenterContent` on `msg-slot`), **user menu** at the trailing edge.
Parts touching the application frame — header, nav, menu, msg, and the
`rightpanel` root/header/title, the `crudview` root — are squared with
`EdgeToEdge()`; interior elements keep their radius.

- **`Platform.Brand`** (`Brand` interface: `BrandName()`, `BrandMark()`) fills the
  leading slot. An empty mark falls back to the shell's `IconBrand` glyph at the
  same box as the avatar (`IconBox(IconLg)` + `Round(RadiusFull)` +
  `HideOverflow`); a `nil` Brand renders no slot at all. `AppName` is NOT a
  fallback: it titles the phone drawer, Brand lives in the desktop header, and
  only one of the two surfaces exists at a time.
- **Notifications** are plain text in severity colour: `Glyph(Subtle|Success|
  Accent|Danger)` emits `color` + `fill: currentColor` with no background box.
- **Navigation semantics**: the current route renders `As(Accent)` (amber —
  "where I am"), hover renders `As(Inset)` (tonal). The light selection tint
  (`Highlight`) is deliberately not used in the nav; selected rows in consumer
  lists re-point at `Accent` for the same reason.

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

### Data Flow (`view.Presenter`)

`crudview` is a pure renderer: it never talks to a `router.Caller` and holds no data of its own.
A domain module builds a `view.Presenter` (via `view.New(caller, record, listOp, newList, project,
opts...)`, importing `view`+`model`+`router`, never `layout`) and hands it to `crudview.New(Config{
Presenter: p})`.

1. `Init` (or a save/delete callback) calls `Presenter.Reload()`, which invokes the list op and
   decodes into `Presenter.Items()` synchronously.
2. `filter()` reads straight from `Presenter.Items()` (no local `allItems` copy).
3. `items` (SignalNodes) is updated, triggering a DOM patch.

Save/Delete follow the same shape: `crudview` syncs form values into `Presenter.Record()`, then
calls `Presenter.Save(record)` / `Presenter.Delete(id)` — both synchronous, both returning `error`
directly.

### Signal Fields

| Field | Type | Role |
|---|---|---|
| `items` | `*SignalNodes` | Holds the rendered card elements for the list. |
| `selected` | `*SignalString` | Holds the ID of the currently selected item. |
| `search` | `*SignalString` | Holds the current search term; triggers `filter()`. |
| `canSave` | `*SignalBool` | Controls the enabled state of the Save button. |
| `canDelete` | `*SignalBool` | Controls the enabled state of the Delete button. |

### El pegamento vive aquí

The high-level pattern for constructing a CRUD view is `crudview.New(Config)`. This constructor is the single place where the standard CRUD view↔form↔`Presenter` loop is wired:

- A module builds a `view.Presenter` (owns list/select/save/delete against a `router.Caller`) and passes it once via `Config{ParentID, Presenter}`.
- All callbacks (`OnSelect`, `OnNew`, `OnSave`, `OnDelete`, `OnCancel`) are automatically wired.
- Form inputs are automatically populated on selection using `form.LoadValues` with records returned by `Presenter.Select(id)`.
- Saves are validated and synced via `form.SyncValues` before shipping to `Presenter.Save`.
- `OnSave`/`OnDelete` are only wired when `Presenter.CanSave()`/`CanDelete()` are true.
- Empty search string placeholders default to `"Search…"`, but can be customized via `Presenter.SearchPlaceholder()`.

#### Principle: Standard-shaped tests

As a policy established in this layer (C4), no public high-level API should be published without a consumer-shaped test (like `crudview/consumer_test.go`) validating the entire integration of forms, models, and transport logic within this library.
