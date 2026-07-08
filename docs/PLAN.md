# PLAN — `NewUIModule` factory, `Platform.CanView` gating, and the `crudview` layout (form + list)

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

## Context (zero-context summary)

`tinywasm/layout` publishes reusable UI layout skeletons for tinywasm modules:

- `platformd` — the SPA shell: header, nav rail, hash routing, notifications.
  Owns the `UIModule` interface (`layout.Module` identity + `Label()` +
  `IconID()` + `View()`).
- `rightpanel` — a two-column content skeleton with optional slots.

Consumers today must (a) hand-write a ~20-line ceremony struct per module just
to satisfy `UIModule`, and (b) hand-assemble every module view from scratch.
This plan removes both and adds cosmetic permission-gating to the shell.

**The target module-view design is NOT invented here** — it replicates the
proven Pa100T layout, reverse-engineered in
[IE_MODULE_CONTENT.md](IE_MODULE_CONTENT.md) (READ IT FIRST — it is the source
of truth for structure, CSS and interaction semantics): **form on the left
(protagonist), record list on the right, CRUD action bar bottom-left
(delete/cancel/new/save), search box bottom-right, module title clearly at the
top-left corner of the panel**. The shell was already replicated
([IE_LAYOUT.md](IE_LAYOUT.md) → `platformd`); this plan replicates the panel
interior as `crudview`.

**Ecosystem pillars (non-negotiable):** minimal WASM binary, avoid heap
allocations, zero consumer boilerplate, reusable architecture.

## The preconfiguration contract (applies to every layout in this repo)

Layouts are not consumed directly by feature modules. The consumer app's
composition root (e.g. `mjosefa-cms/config/layouts`) **preconfigures** each
layout once — app-wide defaults, transport glue, error/notification wiring —
and modules only inject their content through that thin wrapper. For that to
stay thin, every layout here MUST:

1. Be configured by **exported struct fields (data)** — no constructors with
   behavior, no functional options. A wrapper then reads like a struct literal
   with defaults filled in.
2. Have **useful zero values**: every optional field nil/empty = feature off,
   with no error path (nil callback ⇒ button not rendered; zero `Source` ⇒
   full-page variant).
3. Keep **all strings the consumer might brand** (placeholders, labels,
   titles) as fields — never hardcoded inside the layout.
4. Expose behavior only through **typed callbacks and `router.Caller`** — a
   layout never knows models, wire protocols, or `tinywasm/form`.

`rightpanel` already complies; `crudview` (Stage 3) is specified to comply;
any future layout follows the same contract. Document it (Stage 5).

**Gate:** the `crudview` stage requires `router.Caller`
(`tinywasm/router/docs/PLAN.md`) merged and tagged. Stages 1–2 have no gate.

## Stage 1 — `platformd.NewUIModule` (additive)

The owner of a contract provides its trivial constructor (same convention as
`mcp.Text(...)` living next to `mcp.Result`):

```go
// in platformd — kills the per-module ceremony struct
func NewUIModule(id, label, iconID string, view Component) UIModule
```

Private value-type implementation; no exported struct (one way to build).

## Stage 2 — `Platform.CanView` (cosmetic permission gating)

New optional field on `Platform`:

```go
// CanView filters which modules the shell presents. nil = show all.
// resource is the module's ModelName() — by ecosystem convention the module ID
// IS the RBAC resource name. This gating is COSMETIC (the wasm binary is
// public); the server always re-validates every call. Consumers wire it from
// the user profile's permissions after login.
CanView func(resource string) bool
```

Semantics (all three, not just the nav):

1. Nav rail: modules with `CanView(id) == false` render no link.
2. Stage: their `<section>` panel is not built at all (not merely hidden).
3. `Activate(id)` (including via hash) on a non-viewable module is a no-op that
   falls back to the first viewable module — a hand-typed hash must not reveal
   the panel skeleton.

`Init` re-evaluates the default module against `CanView`. Zero allocations when
`CanView == nil` (fast path unchanged).

## Stage 3 — `layout/crudview`: form + list + CRUD bar + search

New package `crudview`, replicating IE_MODULE_CONTENT.md §1–§3. It owns the
recurring cycle (load list → pick record → edit in form → save/delete) that
every CRUD module repeats. **No dependency on `tinywasm/components` and no
dependency on `tinywasm/form`**: the list, search and CRUD bar are internal;
the form arrives as an injected `dom.Component` slot, and all record/form glue
goes through typed callbacks (the consumer's composition root wires them once).

```go
package crudview

// Item is one record in the right-hand list.
type Item struct {
	ID          string // selection key
	Label       string // main text of the card
	Description string // small chip at the card's bottom-right (e.g. an IP)
}

// Source is the data seam: a fake in tests, a router.Caller adapter in prod.
type Source struct {
	Caller router.Caller
	ListOp string                 // logical operation, e.g. "list_devices"
	Args   func() model.Encodable // list request args (nil → no args)
	Decode func(raw []byte) ([]Item, error)
}

type CrudView struct {
	dom.Element // value embed — NEVER *dom.Element

	Title  string        // h1, top-left corner of the panel (white on accent bar)
	Form   dom.Component // LEFT slot, the protagonist (typically *form.Form)
	Source Source        // feeds the right-hand list; zero Source = full-page
	                     // variant: no list/search/CRUD bar (title + form only)

	// Interaction callbacks — the composition root wires these to the form
	// and transport ONCE per app; modules never re-implement them.
	OnSelect func(it Item)                       // list card clicked → load into form
	OnNew    func()                              // (+) pressed → reset form for create
	OnSave   func(done func(err error))          // (💾) pressed; done(nil) reloads list
	OnDelete func(id string, done func(err error)) // (−) pressed on selection
	OnCancel func()                              // (↺) pressed → undo current edit
	OnError  func(err error)                     // list load/decode failures; nil = drop
}

func (v *CrudView) Init(ctx dom.Ctx)     // builds internal list/search/bar, first load
func (v *CrudView) Render() *dom.Element
func (v *CrudView) Reload()              // re-runs ListOp (e.g. after save/delete)
func (v *CrudView) Select(id string)     // programmatic selection (also visual)
```

Behavior (mirrors IE_MODULE_CONTENT.md §3):

- Clicking a card marks it selected (`target-li-on` equivalent) and fires
  `OnSelect`. `(+)` clears selection and fires `OnNew`. `(−)` is disabled with
  no selection; fires `OnDelete(selectedID, done)`. `(💾)` fires `OnSave`;
  on `done(nil)` the view calls `Reload()` itself. `(↺)` fires `OnCancel`.
- Nil callback ⇒ its button is not rendered (read-only views pay nothing).
- The search input filters the loaded items **locally** by `Label` and
  `Description` (signal-driven; no server round-trip, no allocations beyond
  the filtered node list).
- Selection, filter text and button-enabled state live in typed signals.

Implementation rules:

- Typed signals only (`SignalString`/`SignalBool`/`SignalNodes`); zero `any`,
  zero `map`, zero generics, zero stdlib (`tinywasm/fmt`/`json`).
- CSS in `crudview/css.go` under `//go:build !wasm` (`RenderCSS()`), porting
  the captured rules of IE_MODULE_CONTENT.md §2 with tokens instead of
  hardcoded colors (accent `#3f88bf`, grays `#e9e9e9`/`#c2c1c1` become the
  theme tokens platformd already maps — the app brands via `RootCSS`). Grid
  proportions (title 8vh / article / controls 9vh; left ≈ 66/96, right ≈
  29/96) become CSS custom properties with those defaults.
- The four CRUD icons (plus, minus, undo, save — captured as 16×16 paths in
  IE_MODULE_CONTENT.md) are registered via the established SVG pattern:
  `Define`+`Path` in the package, `IconSvg()` under `!wasm`.
- Errors never swallowed: every failure path goes through `OnError` when set.
- Init once-guard like the rest of the ecosystem; no `front.go`.

## Tests (`gotest ./...`, never `go test`)

Prerequisite: `go install github.com/tinywasm/devflow/cmd/gotest@latest`.
The wasm suite activates via `//go:build wasm` files (real browser).

- `platformd`: `NewUIModule` satisfies `UIModule`; `CanView` semantics 1–3.
- `crudview` native: `Render` tree shape with full Source vs zero Source
  (full-page variant); nil callbacks render no buttons; decode error reaches
  `OnError`.
- `crudview` wasm (`crudview_wasm_test.go`): fake `router.Caller` (canned
  bytes, from `router/mock`) → list paints cards with label + description
  chip; clicking a card fires `OnSelect` and marks selection; `(−)` disabled
  without selection, enabled with one; `OnSave` `done(nil)` triggers a second
  `ListOp` call; search input filters the visible cards; transport error hits
  `OnError`.

## Documentation (mandatory)

- `README.md`: add `crudview` to the index with the full consumer example
  (the "pick layout, inject content" shape) and document `NewUIModule` +
  `CanView`. Add a **"Preconfigure, don't assemble"** section showing the
  intended consumption pattern: the app's composition root wraps a layout once
  (e.g. a `config/layouts` package exposing `Crud(cfg)`), and feature modules
  call that wrapper — modules never touch this repo's structs directly.
- `docs/ARCHITECTURE.md`: add `crudview` section (structure diagram from
  IE_MODULE_CONTENT.md §1, the Source seam, signal fields table, callback
  flow) and the `CanView` activation-flow update.
- Demo `web/client.go`: add one `crudview` module wired to a canned in-process
  `Caller` and a plain `dom.Component` form stand-in, so the demo shows the
  intended usage end-to-end.

## Harness checklist (mandatory)

- Value embedding only; typed signals; no `any`/`map`/generics/stdlib in wasm
  paths (`Item` is a concrete struct — no generic rows).
- No repeated string literals in logic (class names are typed `Class` vars, as
  the existing packages already do).
- CSS/SVG only in `css.go`/`svg.go` under `!wasm`.
- One way to build per concern: `NewUIModule` for the wrapper, struct literal
  for `CrudView` — no alternative constructors.
- Errors propagate (`OnError`); no silent `return` on failure paths.
- No existing public symbol changes: `UIModule`/`Platform`/`RightPanel`
  consumers compile unmodified.

## Acceptance criteria

1. `gotest ./...` green, native + browser suites.
2. Visual parity with the reference capture: title top-left, form left with
   fieldset chips, list cards right with description chip, CRUD bar bottom
   with the four icons, search bottom-right (compare demo against
   IE_MODULE_CONTENT.md; screenshot in the PR).
3. `CanView` denies: no nav link, no panel in DOM, hash activation falls back
   (asserted in wasm test).
4. Wire-level knowledge check: `grep -rn "tools/call\|mcp\." crudview/` is
   empty — crudview knows only `router.Caller`. `grep -rn "tinywasm/form\|tinywasm/components" crudview/`
   is also empty.
5. Demo wasm binary size reported before/after in the PR; the shell without
   crudview keeps its previous size (pay-per-use).

## Stages

| Stage | File(s) | Action | Gate |
|---|---|---|---|
| 1 | `platformd/factory.go` | `NewUIModule` | — |
| 2 | `platformd/platformd.go` | `CanView` field + semantics 1–3 | — |
| 3 | `crudview/crudview.go`, `crudview/css.go`, `crudview/svg.go` | the layout per spec + IE_MODULE_CONTENT.md | router plan tagged |
| 4 | `platformd/*_test.go`, `crudview/*_test.go`, `crudview/*_wasm_test.go` | tests above | 1–3 |
| 5 | `README.md`, `docs/ARCHITECTURE.md`, `web/client.go` | docs + demo | 4 |
