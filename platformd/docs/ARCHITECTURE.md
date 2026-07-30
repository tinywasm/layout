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

---

## Anatomy

The chassis is built using a structured hierarchy of semantic HTML elements. Each element maps directly to an anatomical part defined in the stylesheet.

### Part Tree

```
root .pd                          Cover  HideOverflow
├── header .pd__header            KeepSize
│   ├── div  .pd__user-block
│   ├── div  .pd__msg-slot        Fill  (BindChildren(p.notifications))
│   └── div  .pd__header-right    KeepSize
│       ├── h2 .pd__area
│       └── HeaderActions slot
├── button .pd__hamburger         mobile only
├── div    .pd__nav-overlay       mobile only, data-open
├── div    .pd__body              Sidebar(SideEnd)  Fill
│   ├── main .pd__stage           Fill  HideOverflow          ← content, first child
│   │   └── section .pd__panel    Scroll  data-current          (one per module)
│   └── nav  .pd__menu            Fill                        ← rail, LAST child
│       └── ul .pd__navbar
│           └── li .pd__nav-item
│               └── a .pd__nav-link  data-current
│                   ├── svg  .pd__nav-icon  IconBox(IconLg)
│                   └── span .pd__link-text
```

### Where the scroll lives

`Cover()` gives the root a **definite** `height: 100dvh`, not a floor, so the frame
never grows with its content: `HideOverflow()` on the root and on the stage clips,
and the header and rail stay put no matter how tall a module is. The only element
that scrolls is `.pd__panel`, through `Scroll()`. A module that renders its own
scroll region nests it inside that panel.

Every part that renders a bare `<svg>` must declare `IconBox`. An `<svg>` with no
width or height falls back to the replaced-element default of 300×150 and blows the
layout apart — this is what made the rail 150px per item before `IconBox` existed.

### State-Revealed Parts

States are carried by boolean attributes on the DOM, ensuring high efficiency and zero-class toggling logic:
*   **`menu`** (`.pd__menu`) and **`nav-overlay`** (`.pd__nav-overlay`): Revealed by the state `widget.Open` using the attribute `data-open` (driven by the mobile navigation drawer state).
*   **`panel`** (`.pd__panel`) and **`nav-link`** (`.pd__nav-link`): Highlighted or revealed by the state `widget.Current` using the attribute `data-current` (driven by the active route signal).

### Mobile-Only Elements

*   **`hamburger`** (`.pd__hamburger`): Standard button for triggering the navigation drawer. Styled exclusively for `css.Mobile` viewports.
*   **`nav-overlay`** (`.pd__nav-overlay`): Dimmed backdrop appearing behind the drawer navigation panel on mobile viewports.

---

## Architecture Flowchart

```mermaid
flowchart TD
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
