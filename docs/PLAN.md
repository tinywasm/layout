---
PLAN: "fix(platformd): restore the application chassis lost in the widget v0.4 migration"
TAG: v0.1.0
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — `layout`: restore the `platformd` application chassis

**Blocked by two upstream releases. Do not start until both are published:**

| Repo | Version | Provides | Plan |
|---|---|---|---|
| `tinywasm/css` | **v0.4.0** | `css.Device` (`Mobile`/`Tablet`/`Desktop`), `--rail-narrow`, `--rail-wide` | https://github.com/tinywasm/css/blob/main/docs/PLAN.md |
| `tinywasm/widget` | **v0.4.3** | `style.Cover`, `style.Sidebar`, `style.Drawer`, `Sheet.On`, `Sheet.OnlyOn`, `Sheet.StateAttrs` | https://github.com/tinywasm/widget/blob/main/docs/PLAN.md |

First action of stage 1 is bumping both in `go.mod`. If either version is not yet
available, stop and report — **do not** work around a missing primitive with hand-written
CSS, and do not add a `replace` directive.

---

## 0. What is broken, and why (read fully before touching code)

`platformd` is the application chassis: a header, a vertical nav rail, and a stage that
swaps module panels by hash route. It renders **a blank white page** today. Two commits
did it — `f0fc2ea`/`eed374a` (migrate to the widget style DSL) and `7495089` (widget
v0.4.1) — and the two defects are independent. Both must be fixed; fixing one alone still
leaves a broken screen.

### Defect A — the markup never writes the state attributes its own stylesheet selects on

`platformd/css.go` declares `style.RevealedBy(widget.Open)` on the parts `menu`, `panel`,
`orientation-warn` and `nav-overlay`. That emits, in the shipped bundle:

```css
@layer widgets { .pd__menu   { display: none } .pd__panel { display: none } }
@layer states  { .pd__menu[data-open="true"] { display: flex }
                 .pd__panel[data-open="true"] { display: block } }
```

`platformd/platformd.go` still toggles **classes** from before the migration —
`BindClass(string(clsPanelActive), …)` and `BindClass(string(clsMenuOpen), …)` — and never
writes `data-open` anywhere. Verified in the running app:

```
.pd__menu   | display=none | w=0 h=0
.pd__panel  | class="pd__panel pd__panel-active" | display=none | w=0 h=0
```

The nav rail and **every module panel** are therefore hidden permanently. `crudview` was
migrated correctly in the same pass (`crudview/crudview.go:451` uses
`BindAttrBool("data-open", …)`); `platformd` was not, and nothing caught it: `@layer
widgets` outranks `@layer primitives` regardless of specificity, so the `display: flex`
that `Stack()` puts on `.pd__panel-active` never had a chance.

### Defect B — the application frame was deleted and never replaced

The pre-migration `platformd/css.go` (512 lines, see `git show a2e4780:platformd/css.go`)
built the shell with a CSS grid: `"header header" / "module-content menu-container"`, a
fixed header height, a fixed em rail, `100vh` sizing, and a `MediaDesktop(...)` block that
switched between a mobile hamburger drawer and a desktop rail.

The migration replaced all of it with `Stack(SpaceNone)` + `Fill()`. `Fill()` emits
`height: 100%`, which resolves to `auto` because nothing above it has a height — the root
is **73px tall** in the running app. The rail, the header's three-column split, the
non-shrinking header and the mobile/desktop switch are simply gone.

This was foreseen and then dropped: `docs/VISUAL_CONTRACT_MASTER_PLAN.md` §8 states
platformd's viewport math *"desaparece con `Cover()` + `Fill()`"*, and §2.4 lists *Sidebar*
among the primitives `style.Flow` would provide. Neither `Cover()` nor `Sidebar()` was
ever implemented. Acceptance criterion §9.6 of that plan — *"verificación visual en vivo
(MCP screenshot), escritorio y móvil emulado, claro y oscuro"* — was not performed.
Upstream `widget` v0.5.0 now supplies both primitives; this plan consumes them.

### Reference screenshots

- `platformd/docs/last-acceptable-view.png` — the target: header with user block left and
  work-area name + theme toggle right, blue module panel filling the stage, nav rail down
  the right edge with the active item highlighted.
- `platformd/docs/current-view-error.png` — today: a squashed header, a stray full-width
  strip (that is `.pd__hamburger`, which has no desktop rule any more), and white below.

Colour differences between the two are **not** in scope: the acceptable capture predates
the `css` v0.3.x 12-token palette. Match the **structure**, not the hues.

---

## 1. Repo rules the executor must follow

Restated from `AGENTS.md` because a literal-minded agent will otherwise break them:

- **Hot reload — never compile manually.** The TinyWasm dev server rebuilds on every save.
  Do not run `go build` or `GOOS=js GOARCH=wasm go build` to "pick up" a change. A one-off
  compile check is the only reason to build by hand.
- **`gotest`, never `go test`.** Stdlib assertions only (`testing`/`strings`), no testify.
- **No Go stdlib in WASM paths.** Use `github.com/tinywasm/fmt`, never `fmt`/`strconv`/
  `strings`/`errors`. DOM only through `github.com/tinywasm/dom`, never `syscall/js`. No
  `encoding/json`. `switch`, not `map`. No `defer`/`recover`. Embed `dom.Element` **by
  value**.
- **No generics.** Concrete typed signals only: `SignalString`, `SignalBool`,
  `SignalNodes`, `DeriveString`, `DeriveBool`, `Bind*`. Never `Signal[T]`.
- **One component contract.** `Render() *dom.Element` plus optional `Init(ctx dom.Ctx)`.
  There is no `OnMount`/`OnUpdate`/`OnUnmount` and no manual `Update()`. Handlers only
  mutate signals; the bound DOM patches surgically. Never re-render the whole platform.
- **Build tags.** `css.go` is `//go:build !wasm` and must never reach the WASM binary.
  Untagged files ship to WASM.
- **SSR provider names are matched by regex and must be exact.** The CSS entry point must
  be the method `RenderCSS`; anything else is silently never emitted and the component
  renders unstyled. All providers in a package share one receiver.

### 1.1 Anti-footguns specific to this plan

- **`Platform.WidgetKind()` returns `widget.Menu`. Leave it.** `widget.Kind.Allows`
  (`kind.go:112`) permits `Open` **and** `Current` for `Menu` only. Changing it to
  `Region` makes both states invalid and `Stylesheet()` panics at SSR time with
  `state Open is not meaningful for kind Region`. `Menu` also maps to `LayerDropdown`
  (`--z-dropdown: 100`), which is the layer the drawer and its backdrop need.
- **Do not "fix" the widget style DSL from this repo.** If a visual need cannot be
  expressed with the options `widget` v0.5.0 exports, that is an upstream defect: stop and
  report it. Do not reach for `css.Raw()`, do not hand-write a `@media` block, do not add
  `!important`.
- **Do not touch `crudview` or `rightpanel` sheets** except where §5 explicitly says so.
  They are not the cause of the blank page.
- **`gopush` and `codejob` are developer tools managed outside this plan.** Never invoke
  them.

---

## 2. Stage 1 — dependency bump

In `go.mod`:

```
github.com/tinywasm/css    v0.4.0
github.com/tinywasm/widget v0.5.0
```

Leave every other dependency at its current version. The `replace` directives at the
bottom of `go.mod` are all commented out — leave them commented.

---

## 3. Stage 2 — rewrite `platformd/css.go`

Replace the whole file. Target below; every part name here must exist in the markup
produced by stage 3, and vice versa — stage 4 asserts it.

```go
//go:build !wasm

package platformd

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

// RenderCSS implements the visual contract for platformd using the style DSL.
func (p *Platform) RenderCSS() *css.Stylesheet {
	return style.For(p).
		// The outermost frame of the application: fills the viewport, stacks
		// header over body. Everything else sizes against this.
		Root(
			style.Cover(),
			style.As(style.Page),
		).
		Part(widget.Part("header"),
			style.Row(style.Space2),
			style.KeepSize(),
			style.As(style.Panel),
			style.Pad(style.Space1),
		).
		Part(widget.Part("user-block"),
			style.Row(style.Space1),
			style.FontSize(style.TextBase),
			style.FontWeight(style.WeightBold),
		).
		// Fill() here is what pushes header-right to the far edge: it grows to
		// take the free space between the two blocks.
		Part(widget.Part("msg-slot"),
			style.Row(style.Space1),
			style.Fill(),
		).
		Part(widget.Part("msg"),
			style.Pad(style.Space1),
			style.Round(style.RadiusSm),
		).
		Part(widget.Part("msg-info"), style.As(style.Subtle)).
		Part(widget.Part("msg-success"), style.As(style.Success)).
		Part(widget.Part("msg-warning"), style.As(style.Highlight)).
		Part(widget.Part("msg-error"), style.As(style.Danger)).
		Part(widget.Part("header-right"),
			style.Row(style.Space2),
			style.KeepSize(),
		).
		Part(widget.Part("area"),
			style.FontSize(style.TextBase),
			style.As(style.Subtle),
		).
		// The rail sits at the inline-end edge; the stage takes everything else.
		// Below the stage's minimum width the two reflow into one column with no
		// media query — that is Sidebar's own behaviour, not something to add.
		Part(widget.Part("body"),
			style.Sidebar(style.SideEnd, style.RailNarrow, style.SpaceNone),
			style.Fill(),
		).
		Part(widget.Part("stage"),
			style.Fill(),
			style.HideOverflow(),
		).
		Part(widget.Part("panel"),
			style.Stack(style.SpaceNone),
			style.Fill(),
			style.RevealedBy(widget.Current),
		).
		Part(widget.Part("menu"),
			style.Stack(style.Space1),
			style.As(style.Panel),
			style.Fill(),
		).
		Part(widget.Part("navbar"),
			style.Stack(style.SpaceNone),
			style.Fill(),
		).
		Part(widget.Part("nav-item"),
			style.Row(style.SpaceNone),
			style.Fill(),
		).
		Part(widget.Part("nav-link"),
			style.Row(style.Space1),
			style.Pad(style.Space2),
			style.Fill(),
			style.Animate(style.MotionFast),
		).
		Part(widget.Part("nav-icon"),
			style.Width(style.Content),
		).
		Part(widget.Part("link-text"),
			style.FontSize(style.TextBase),
		).
		// The active route reads as "current", the same vocabulary the rail and
		// crudview's list rows share. It is a STATE, never a class.
		When(widget.Current, widget.Part("nav-link"),
			style.As(style.Highlight),
		).
		Cue(widget.Hover, widget.Part("nav-link"),
			style.As(style.Panel),
		).
		// ── mobile-only chrome ────────────────────────────────────────────────
		OnlyOn(css.Mobile, widget.Part("hamburger"),
			style.Row(style.Space1),
			style.As(style.Primary),
			style.Pad(style.Space2),
			style.Round(style.RadiusSm),
		).
		OnlyOn(css.Mobile, widget.Part("nav-overlay"),
			style.Backdrop(style.Viewport),
			style.Veil(),
			style.RevealedBy(widget.Open),
		).
		// On a phone the rail stops being a column and becomes a panel that
		// slides in from the edge, gated by the same Open state as the overlay.
		On(css.Mobile, widget.Part("menu"),
			style.Drawer(style.SideEnd, style.TwoThirds),
			style.Stack(style.Space1),
			style.As(style.Panel),
			style.RevealedBy(widget.Open),
		).
		Stylesheet()
}
```

**`orientation-warn` is deleted, not ported.** It was an empty placeholder `<div>` with no
content and no binding, shown by the pre-migration sheet through
`Media("(orientation: landscape) and (min-width: 600px) and (max-width: 1024px)")`.
`css.Device` classifies by width only — orientation is deliberately not part of the closed
enum (see the `css` plan §6) — so the trigger cannot be expressed, and the element has
displayed nothing since the migration regardless. Remove the part from the sheet, the
`clsOrientationWarn` variable, and the `root.Child(...)` call that appends it. If a
landscape-tablet warning is wanted later it is a new feature with real content, planned on
its own.

### 3.1 What changed and why — do not silently revert any of these

| Before | After | Reason |
|---|---|---|
| `Stack(SpaceNone)` + `Fill()` on root | `Cover()` | `Fill()` is `height: 100%` against an auto-height ancestor — it resolves to nothing. `Cover()` is the viewport frame. |
| no `body` part | `Sidebar(SideEnd, RailNarrow, SpaceNone)` | The rail and the stage must be siblings in one container for `Sidebar` to place them. This is a **new element** in the markup — see §4. |
| `panel` + `panel-active` | `panel` with `RevealedBy(widget.Current)` | Which panel is showing is a route state, not a class. `panel-active` disappears entirely. |
| `nav-active` part | `When(widget.Current, "nav-link", As(Highlight))` | Same reason. `nav-active` disappears entirely. |
| `menu` with unconditional `RevealedBy(Open)` | plain part; `RevealedBy(Open)` only inside `On(css.Mobile, …)` | This single line is what hid the rail on every viewport. On desktop the rail is always visible. |
| `msg-desktop` + `msg-mobile` | one `msg-slot` | The mobile slot was a full-viewport veil that covered the app whenever a toast existed — and its type classes (`pd-msg-info`, built as a raw string in `platformd.go:274`) matched **no selector at all** after the migration, so message severity had no colour. One slot with four typed variants replaces both. |
| `hamburger` styled for all viewports | `OnlyOn(css.Mobile, …)` | The stray full-width strip under the header in `current-view-error.png` is this element with no desktop rule. |
| `Animate(MotionSlow)` on root and menu | `Animate(MotionFast)` on `nav-link` only | A transition on the shell root animates layout on every route change. Motion belongs on the thing that reacts to a pointer. |

### 3.2 Accepted cosmetic differences from `last-acceptable-view.png`

State these in the PR description; they are deliberate, not regressions to chase:

- The work-area name renders in the module's own casing (`crud`), not forced uppercase.
  `text-transform` is not in the style DSL and will not be added. A module wanting
  uppercase returns an uppercase `Label()`.
- Panel slide-in/slide-out keyframes are gone. `Keyframes` is not in the DSL; `Animate`
  covers transitions only.
- Palette follows `css` v0.3.3 tokens throughout.

---

## 4. Stage 3 — rewrite `Render()` in `platformd/platformd.go`

### 4.1 Class variables

Delete `clsPanelActive`, `clsNavActive`, `clsMenuOpen`, `clsMsgDesktop`, `clsMsgMobile`,
`clsOrientationWarn`.
Add `clsBody`, `clsMsgSlot`, and the four message-variant classes. Every remaining
variable must correspond to a `Part()` in §3.

```go
var (
	clsRoot            = NamePlatform.Root()
	clsHeader          = NamePlatform.Class("header")
	clsUserBlock       = NamePlatform.Class("user-block")
	clsMsgSlot         = NamePlatform.Class("msg-slot")
	clsMsg             = NamePlatform.Class("msg")
	clsMsgInfo         = NamePlatform.Class("msg-info")
	clsMsgSuccess      = NamePlatform.Class("msg-success")
	clsMsgWarning      = NamePlatform.Class("msg-warning")
	clsMsgError        = NamePlatform.Class("msg-error")
	clsHeaderRight     = NamePlatform.Class("header-right")
	clsArea            = NamePlatform.Class("area")
	clsBody            = NamePlatform.Class("body")
	clsStage           = NamePlatform.Class("stage")
	clsPanel           = NamePlatform.Class("panel")
	clsMenu            = NamePlatform.Class("menu")
	clsNavbar          = NamePlatform.Class("navbar")
	clsNavItem         = NamePlatform.Class("nav-item")
	clsNavLink         = NamePlatform.Class("nav-link")
	clsLinkText        = NamePlatform.Class("link-text")
	ClsNavIcon         = NamePlatform.Class("nav-icon")
	clsHamburger       = NamePlatform.Class("hamburger")
	clsNavOverlay      = NamePlatform.Class("nav-overlay")
)
```

`ClsNavIcon` stays exported — `factory.go` and the demo use it.

### 4.2 State attributes — the actual bug fix

Every element the stylesheet reveals by state must bind that attribute. The attribute keys
are not free strings: they come from `widget.State.Attr()` (`state.go:41`) —
`Open → data-open`, `Current → data-current`.

```go
// menu: revealed by Open on mobile
nav := Nav().Set(clsMenu.AsAttr()).
	BindAttrBool("data-open", p.menuOpen)

// overlay: same state, same signal
overlay := Div().Set(clsNavOverlay.AsAttr()).
	BindAttrBool("data-open", p.menuOpen)

// each module panel: revealed by Current
panel := Section().Set(clsPanel.AsAttr()).
	ID(id).
	Attr("data-id", id).
	BindAttrBool("data-current", DeriveBool(func() bool {
		return p.active.Get() == id
	}))

// each nav link: highlighted by Current
link := A("#"+id).Set(clsNavLink.AsAttr()).
	Attr("data-id", id).
	BindAttrBool("data-current", DeriveBool(func() bool {
		return p.active.Get() == id
	}))
```

`BindAttrBool(name string, on *SignalBool) *Element` is `dom` v0.11.4 `element.go:141`.
`DeriveBool` returns a `*SignalBool`, so it composes directly. **Every `BindClass` call in
this file is removed** — grep-verifiable in §7.

### 4.3 Tree shape

The rail and the stage must become siblings inside `.pd__body`, and — because
`Sidebar(SideEnd, …)` makes the **last** child the rail — the stage is appended first:

```
root .pd                          Cover
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
│   │   └── section .pd__panel    data-current                  (one per module)
│   └── nav  .pd__menu            Fill                        ← rail, LAST child
│       └── ul .pd__navbar
│           └── li .pd__nav-item
│               └── a .pd__nav-link  data-current
│                   ├── svg  .pd__nav-icon
│                   └── span .pd__link-text
```

The hamburger and the overlay stay **outside** `.pd__body`: they are viewport-fixed
overlays, and a third element inside a `Sidebar` container would break its
`:first-child`/`:last-child` contract.

### 4.4 Notifications

`buildToasts` currently builds the variant class as a raw string
(`"pd-msg-" + Convert(n.Type.String()).ToLower().String()`), which matches nothing. Replace
the string concatenation with a `switch` over `MessageType` returning one of the four
`clsMsg*` variables — **`switch`, not a map**, per the WASM rules.

Both `#pd-msg-desktop` and `#pd-msg-mobile` containers collapse into one
`.pd__msg-slot`; `BindChildren(p.notifications)` moves onto it. Delete the now-unused
second container and its ID.

### 4.5 Unchanged

`Init`, `Activate`, `isViewable`, `fallback`, `Notify`, `dismiss`, `notificationCount`,
the `UIModule` interface and the hash-routing behaviour are correct. Do not modify them.
`p.menuOpen` keeps its role — it now drives `data-open` instead of a class.

---

## 5. Stage 4 — the test that makes this bug class impossible to reship

`platformd` has no stylesheet-conformance test; `crudview` has one
(`crudview/consumer_stylesheet_test.go`) that compares class sets — and it would **not**
have caught this bug, because the mismatch was in state *attributes*, not classes.

Create **`platformd/consumer_stylesheet_test.go`** (`//go:build !wasm`), modelled on the
crudview file, with three assertions:

1. **Class parity, both directions.** Every `pd`-prefixed class in the rendered markup
   exists as a selector in the sheet, and every `pd`-prefixed class selector in the sheet
   appears in the markup. Reuse the extraction helpers from
   `crudview/consumer_stylesheet_test.go` — attributes are single-quoted in `dom` output.

2. **State-attribute parity — the new one.** For every pair returned by
   `Sheet.StateAttrs()` (widget v0.5.0), assert the rendered markup contains that
   attribute key. Failure message, verbatim:
   `stylesheet selects on %q but no element in the markup ever writes it`

   Rendering `Platform` with at least two modules and one queued notification is required
   for the markup to contain every branch.

3. **No forbidden output.** The emitted sheet contains no `!important`, and declares
   `@layer tokens, primitives, widgets, states;` exactly once.

Then extend `crudview/consumer_stylesheet_test.go` and add the equivalent to
`rightpanel/` with assertion 2 — that is the only change those packages get. `crudview`
already writes `data-open` correctly, so it should pass immediately; if it does not, that
is a second real bug and it must be reported, not worked around.

`platformd/platformd_test.go` (232 lines) exists — update any assertion referencing the
deleted classes rather than deleting the test.

---

## 6. Stage 5 — live verification (mandatory; this is the step that was skipped last time)

The dev server runs at `http://localhost:8080` against `platformd/web/`. Hot reload is on;
do not rebuild by hand.

| # | Check | Expected |
|---|---|---|
| 1 | Desktop, default route | Header spans the top with the user block left and area name + theme toggle right; the blue CRUD panel fills the stage; the nav rail runs down the right edge. |
| 2 | `document.querySelector('.pd').offsetHeight` | Equals the viewport height, not ~73. |
| 3 | `.pd__menu` and `.pd__panel[data-current]` computed `display` | Neither is `none`. |
| 4 | Click each nav item | Panel swaps, `data-current` moves, hash updates. |
| 5 | Emulate a phone (≤639px) | Rail is hidden; the hamburger appears; tapping it slides the drawer in over a dimmed backdrop; tapping the backdrop closes it. |
| 6 | Emulate a tablet (640–1023px) | Rail behaves as on desktop; no hamburger. |
| 7 | Light and dark | Both legible; the shell frame changes with the theme. |
| 8 | Console | No errors. |

Capture desktop and mobile screenshots and attach them to the PR. Replace
`platformd/docs/last-acceptable-view.png` with the new desktop capture **only after**
checks 1–8 pass, and delete `platformd/docs/current-view-error.png` — it documents a fixed
defect.

---

## 7. Acceptance criteria

Each is grep- or test-verifiable.

1. `gotest` green in this repo.
2. `grep -rn "BindClass" platformd/` → **empty**. State is carried by attributes only.
3. `grep -rn "panel-active\|nav-active\|menu-open\|msg-desktop\|msg-mobile\|orientation-warn" .`
   → **empty**.
4. `grep -rn "css.Raw\|!important\|@media\|RawRule" --include=*.go .` → **empty**. No
   hand-written CSS survived.
5. `grep -rn "pd-msg-" --include=*.go .` → **empty** (the dead raw-string variant classes).
6. `GOOS=js GOARCH=wasm go list -deps ./platformd | grep -E "tinywasm/(css|widget/style|svg/sprite)"`
   → **empty**. Build-time-only packages must not ship to the browser.
7. `Sheet.StateAttrs()` for `platformd`, `crudview` and `rightpanel` is fully covered by
   each package's markup — asserted by the stage-4 tests, not by inspection.
8. All eight live checks in stage 5 pass, with screenshots attached.

---

## 8. Stages

| # | Stage | Files | Gate |
|---|---|---|---|
| 1 | Dependency bump | `go.mod`, `go.sum` | blocks all |
| 2 | Rewrite the sheet | `platformd/css.go` | — |
| 3 | Rewrite the markup + state bindings | `platformd/platformd.go` | needs 2 |
| 4 | Conformance tests | `platformd/consumer_stylesheet_test.go` (new), `platformd/platformd_test.go`, `crudview/consumer_stylesheet_test.go`, `rightpanel/` (new test) | needs 3 |
| 5 | Live verification + screenshots | `platformd/docs/*.png` | needs 4 |
| 6 | Docs | `platformd/docs/ARCHITECTURE.md`, `docs/ARCHITECTURE.md`, `README.md` | needs 5 |

Stages 2 and 3 are one logical change and must land together — the sheet and the markup
are two halves of the same contract, which is precisely what came apart last time.

### Stage 6 — documentation

- **`platformd/docs/ARCHITECTURE.md`** — add an "Anatomy" section: the part tree from §4.3,
  which parts are state-revealed and by which state, and which are mobile-only. Add a
  `flowchart TD` of the shell (no `subgraph`, `<br/>` for breaks).
- **`docs/ARCHITECTURE.md`** — record that the chassis is built from `Cover` + `Sidebar`
  and that route selection is `widget.Current`, never a class.
- **`README.md`** — re-index every file under `docs/`.
- **`docs/VISUAL_CONTRACT_MASTER_PLAN.md`** — leave it in place; it is the historical
  record of the decision. Do not edit it to hide the gap.
- Everything in English.
