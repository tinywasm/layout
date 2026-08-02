# Design Roadmap — tinywasm/layout CRUD (mobile + desktop)

Goal: ONE crudview usable on both mobile and desktop, split into small
single-responsibility components — each owns its OWN css; `crudview` owns only
the grid + the mobile slide. Verify every change with the MCP screenshot, on
desktop AND an emulated mobile viewport, light + dark.

## Target layout
- **Desktop**: two columns always visible — edit/form (left) · list column (right).
- **List column** (top → bottom): search (plain input, NOT selectsearch, by
  explicit decision) · targetlist (rows) · ONE crud button.
- **Mobile**: horizontal scroll-snap of two panels, `direction:rtl`-anchored so
  the physical order matches desktop — form (`--cv-detail-width`, 90vw) on the
  left, list (100vw) on the right and visible by default with zero mount-time
  JS. Selecting a row (or swiping) snaps to the form, which always enters
  **from the left**; the list edge peeks (~10vw). A "‹ back" button (mobile
  only) also returns to the list without clearing the selection.

## Components (SRP — each owns its own CSS)
- **filter slot**: `rightpanel.AsideControls` takes any `widget.Filterable`.
  `crudview.New` installs a `components/searchbar.SearchBar` by default; a
  calendar or a select replaces it without touching either package.
  (Supersedes the v0.1 decision to hardcode a plain input.)
- **targetlist** (`components`): list + row cards + a ⋮ options menu
  (Editar/Eliminar), selected state, one shared native `<details
  name="tl-menu-group">` accordion group so only one row's menu is ever open.
- **crud button**: single stateful toggle ("+" ↔ "↺"), owned by crudview
  (NOT the shared `actionbutton` component — incompatible legacy markup).
- **fieldset** (`components`): the form-field skin, including the
  locked/read-only look.
- **modaldialog** (`components`): delete confirmation.
- **crudview** (`layout`): orchestrates the master-detail grid + mobile
  scroll-snap + the toggle button's active/composing state. Holds NO
  component-internal css.

## Behavior
- Search: plain input filters the targetlist live.
- Crud button: "+" by default; **active** (shows "↺") whenever a row is
  selected OR the user is mid-composing a new record (`CrudView.active()`).
  Undo resets everything back to "+".
- Save: auto-save on blur/change (`Form.OnFieldChange`) — no save button.
  Works the same for a brand-new (unselected) record as for an existing one.
- Row ⋮ menu: Editar (unlocks for editing, focuses first field), Eliminar
  (confirms via modaldialog).
- Edit gating: selecting a row shows it read-only (whole-form gate via
  `Form.SetLocked`); ⋮ → Editar unlocks it and focuses the first field.
- "+" also focuses the first field (`view/conformance` "new/edit focuses
  first field" clause — a cross-renderer contract, not crudview-specific).
- List order: newest record first.

## Done
- [x] `targetlist` component (rows, ⋮ menu, selected state).
- [x] crudview wired to targetlist; search at top of list column.
- [x] Single toggle button (+ ↔ ↺) replacing the 4-button bar.
- [x] Auto-save on field blur/change (`Form.OnFieldChange`, `tinywasm/form`).
- [x] Read-only gating: `Form.SetLocked`, select-locks/Editar-unlocks,
  fieldset shows a "frosted glass" tint (`ColorSurface`, not `ColorMuted` —
  text must stay fully legible/selectable while locked).
- [x] Mobile master-detail scroll-snap (`--cv-detail-width`, `order`,
  `dom.Reference.ScrollIntoView()` for the forward snap).
- [x] Mobile direction fix: the form was entering from the right (physical
  order was list-left/form-right, opposite of desktop's form-left/list-right)
  — reported as unintuitive. Fixed by adding `direction:rtl` to the scroll
  strip (`order:1` then renders at the container's "start" edge, which RTL
  makes the right — where the browser's native default scroll position
  already rests, so the list still shows by default with no mount-time JS)
  and `direction:ltr` reset on each panel so its own content isn't mirrored;
  `order`/`scroll-snap-align` values did not need to change. No changes to
  `tinywasm/dom` — `ScrollIntoView()` stays exactly as-is. Added a "‹ back"
  button (mobile-only, local `iconCrudBack` in `crudview/svg.go`) alongside
  the swipe, not replacing it — its handler only moves the viewport
  (`ScrollIntoView` on the list panel), it does not call `undoAction`.
- [x] Mobile width fix: a gray strip of bare stage showed on one side and the
  view didn't use the full width. Root cause was in `platformd`, not crudview:
  `clsPanelActive` sized the active module panel at `--pd-content-width` (96vw)
  on mobile, so every module panel was 15px narrower than the screen — and any
  module sizing its content in `vw` (crudview's `100vw`/`90vw` scroll-snap
  panels) then overflowed that narrower container by the gap. Fixed at the
  root: mobile `clsPanelActive` is now `width:100%` (fills the screen, matching
  the desktop override); the now-unused `tokenContentWidth`/`--pd-content-width`
  token was removed. Also decoupled crudview's panels from the viewport
  (`100vw`/`90vw` → `100%`/`90%`) so they always equal their scroll container
  regardless of the host panel width — preventing this bug class from
  recurring.
- [x] Delete confirmation via `modaldialog` (`ModalDialog.Close()` added).
- [x] targetlist ⋮ menu: `name="tl-menu-group"` accordion + `closeAllMenus()`
  + backdrop-click-to-close (CSS `:has()`); last row's dropdown flips upward
  so it isn't clipped by the scrolling list container.
- [x] "+"/Editar focus the first field (`Form.Focus()`/`FocusedFieldID()`,
  new `view/conformance.Driver` fields `New`/`Edit`/`FocusedFieldID` +
  clauses, implemented by `view/mock.Renderer` and crudview).
- [x] Toggle button flips to "↺" on "+" too, not just on row-select
  (`CrudView.active()` = selected≠"" OR composing; `composing` set by
  `newAction`, cleared by `undoAction`/selecting a row).
- [x] Real in-memory backend for the demo (`platformd/web/client.go`):
  `tinywasm/storage/mem` + `tinywasm/orm` replace the static 3-item fixture,
  so save/list/delete are the real thing — new records persist for the life
  of the process (reset on restart, by design) and sort newest-first.
- [x] Desktop panels didn't fill the stage's height — `.cv-module-content`'s
  `grid-template` used `none` (auto) for the row track, which only grows to
  its content's intrinsic height; the container itself measured full height,
  but its two columns stopped short and left bare background below. Fixed:
  `grid-template: 1fr / 2fr 1fr` (`crudview/css.go`).
- [x] Verified live via MCP on desktop + emulated mobile, light + dark,
  throughout, not as a separate final pass.
- [x] Fieldset polish (`components/fieldset`, `tinywasm/form`):
  - Error message moved inside the box's top-right corner (same offset as the
    box's own padding — even top/right margins), transparent background,
    absolutely positioned so it never grows the box regardless of message
    length.
  - Placeholder no longer defaults to the field name (`form/input/base.go`) —
    the label chip already shows it; a placeholder repeating it added
    nothing. Widgets that need a real example value still set one explicitly
    (`input.Address`, `input.Rut`).
  - Cancelling ("↺") a new-record draft no longer leaves a field focused —
    `undoAction` stopped calling `Form.Focus()` (only `newAction`/`editAction`
    should), and `Form.Reset()` now also clears the tracked
    `FocusedFieldID()`. Standard behavior: new `view/conformance.Driver.Cancel`
    field + `cancel_clears_focus` clause, implemented by `view/mock.Renderer`
    and crudview.
- [x] Visual integration pass (against a reference screenshot of the original
  design):
  - Hover color was inconsistent (`ColorPrimary`, hardcoded `filter:
    brightness`, ad-hoc backgrounds) across components — `tinywasm/css`
    already publishes a dedicated `ColorHover` token; switched `targetlist`'s
    row hover and `fieldset`'s field hover to it so every hover indicator
    reads as the same color app-wide.
  - The list/form panels looked "cut" — partial border-radius on
    `clsAsideSearch`/`clsAsideList`/`clsAsideActions` implied they connected,
    but a `gap` between them (and a separate nested "inset box" inside each
    white panel) meant they didn't. Replaced with the reference's model: each
    column is ONE bordered card (`Border` + `BorderRadius` + `Overflow(Hidden)`
    on `clsArticleContend`/`clsAsideContend`) — the title/search band sits
    flush at the top (no radius/background of its own; the parent card
    clips it), the content fills the rest, and inner "double box" nesting
    (`clsBoxContent`, `clsListaBox`) lost its own margin/radius/background
    duplication. The toggle button moved OUT of the list card into a new
    sibling `clsAsideWrap` (flex column: card, then button) — it's its own
    floating piece, not a third row stitched onto the list, matching the
    reference. Mobile's scroll-snap sizing moved from `clsAsideContend` to
    `clsAsideWrap` (the new direct flex child).

## Notes for the next agent
- Hot reload auto-compiles — never `go build`/restart to see a change
  (AGENTS.md); go.mod `replace` changes DO need a relink.
- Use `gotest`, never bare `go test`, in any tinywasm repo.
- `replace` (TEMP, local dev) points dom, css, form, components, view at local
  checkouts — `orm`/`storage` are NOT replaced (published versions, unmodified).
- SRP for css: a component's look lives IN that component; crudview holds
  only the layout + slide. Adjacent `RawRule`s need their own `;`.
- The demo's in-memory DB (`platformd/web/client.go`'s `deviceDB`) is
  package-level and built once — never re-create it inside `View()`, or
  switching tabs would wipe the data every time.
