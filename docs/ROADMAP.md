# Design Roadmap — tinywasm/layout CRUD (mobile + desktop)

Goal: ONE crudview usable on both mobile and desktop. The way to get there is to
split into small single-responsibility components — each owns its OWN css; the
`crudview` layout owns only the grid + the mobile slide. Verify every change with
the MCP screenshot, on desktop AND an emulated mobile viewport, light + dark.

Prior work (form population fix, fieldset skin, reference palette, header,
theme-aware selection) is merged — do NOT redo it.

## Target layout
- **Desktop**: two columns always visible — edit/form (left) · list column (right).
- **List column** (top → bottom): search (plain input, kept as-is — NOT
  selectsearch, by explicit decision) · targetlist (rows) · ONE crud button.
- **Mobile**: a horizontal scroll-snap of two panels — list (100vw) and form
  (`--cv-detail-width`, 90vw). Selecting a row snaps to the form; the list edge
  peeks (~10vw). Swipe with the thumb → snaps back to the list to pick another.
  The 90% is a single token so it is easy to tune later.

## Components (SRP — each owns its own CSS)
- **search**: plain `<input type=search>` inside crudview — kept as-is, NOT the
  `selectsearch` component (explicit decision, do not swap it in).
- **targetlist** (NEW, in components): list + row cards + a ⋮ options menu
  (Editar / Eliminar) top-right of each row; selected state. Callbacks:
  OnSelect, OnEdit, OnDelete.
- **crud button**: single stateful button, owned by crudview itself (NOT the
  shared `actionbutton` component — that package's CSS/markup is a legacy,
  incompatible `button[name=...]` background-image system; reusing it would
  have fought the icon-based look already in place). Toggles + ↔ ↺.
- **fieldset** (exists): the form-field skin.
- **modaldialog** (exists): delete confirmation.
- **crudview** (layout): orchestrates the master-detail grid + mobile scroll-snap.
  Holds NO component-internal css.

## Behavior (from the design Q&A)
- Search: the plain input filters the targetlist live (unchanged).
- CRUD button: "+" (new) by default; selecting an existing row → "↺" (undo).
  Undo undoes everything → deselects → clears the form → back to "+".
- Save: **auto-save** — no save button; a field commits on blur / Enter.
- Row ⋮ menu: **Editar**, **Eliminar** (delete confirms via modaldialog).
- Edit gating: selecting a row shows it **read-only**; ⋮ → Editar unlocks the
  fields. (Whole-form gate, not per-field.)

## Steps
- [x] `targetlist` component: rows (label + description badge), selected state, and
      a ⋮ menu with Editar/Eliminar (native <details>). Own css. OnSelect/OnEdit/
      OnDelete callbacks. Built + tested.
- [x] crudview uses targetlist for the list; search moved to the TOP of the list
      column (filters the list). Dead cv-list/row css removed (SRP).
- [x] DECIDED: keep the plain search input (do NOT swap for `selectsearch`).
- [x] Replaced the 4-button bar with ONE toggle button (+ ↔ ↺) at the bottom of
      the list column (`btn_crudtoggle`); Editar/Eliminar live in the ⋮ menu.
      Icon swap is reactive (BindClass hides one of two icon children based on
      `v.selected`). Undo fully resets (deselect + clear form + OnCancel).
- [x] Auto-save: added `Form.OnFieldChange(fn)` to `tinywasm/form` (local
      `replace`) — fires on blur (text/textarea/datalist) or change
      (select/radio), AFTER the field's value signal is updated. crudview wires
      it to `autoSaveAction` (validate → SyncValues → Save → Reload). Verified
      live (toast fires) AND with a new Go wasm test
      (`form/tests/onfieldchange.front_test.go`) that caught and fixed a real
      bug: the callback was captured by value at field-construction time inside
      `New()`, before `OnFieldChange` is even called (it's meant to be chained
      AFTER `New()`, like `HideSubmit`) — fixed by capturing a closure that
      re-reads `f.onFieldChange` at commit time instead.
- [x] Form read-only by default; ⋮ Editar unlocks the fields. Added
      `Form.SetLocked(bool)` to `tinywasm/form` (local `replace`) — a whole-form
      reactive gate (`*dom.SignalBool` shared by every `fieldComponent`) that
      disables text/textarea/datalist/select/radio inputs via
      `BindAttrBoolFunc("disabled", ...)`, replacing the old static
      `IsDisabled()`-only attribute. `crudview.selectAction` locks whenever an
      existing row is selected (`it.ID != ""`); `editAction` (⋮ → Editar)
      unlocks after selecting; `newAction`/`undoAction` unlock (new/blank record
      stays editable). `fieldset`'s `RenderCSS` gained a `:has(:disabled)` /
      `:disabled` skin (muted background, `not-allowed` cursor) so the gate is
      visible, not just functional. Verified with a new Go test
      (`crudview/consumer_test.go` `TestConsumer_ReadOnlyGating`) and live via
      MCP screenshots: select → grayed-out fields, ⋮ Editar → editable again,
      confirmed in both light and dark.
- [ ] Mobile master-detail: horizontal scroll-snap; `--cv-detail-width: 90vw`,
      list 100vw, snap + ~10vw peek, slide left→right. Desktop keeps two columns.
- [ ] Delete confirmation via modaldialog.
- [ ] Verify desktop + emulated mobile, light + dark.

## Known issue
- ~~Some module icons (crudview `+`/`↺`/search, targetlist `⋮`) render blank/inconsistently
  in the served sprite.~~ FIXED upstream in `tinywasm/app` (2026-07-22): assetmin's
  logger is now wired (`AssetsHandler.SetLog`, was silently dropping mass-scan
  errors) and the watcher now requests a browser reload after the background
  `LoadSSRModules` scan lands, so dependency-module icons/CSS are no longer
  missed on a cold start. If icons still render blank after pulling the latest
  `app`, this is a NEW regression, not the old bug — re-open with fresh
  repro details.

## Notes for the next agent
- Hot reload auto-compiles — never `go build`/restart to see a change (AGENTS.md);
  go.mod `replace` changes DO need a relink.
- `replace` points dom, css, form, components at local checkouts.
- SRP for css: a component's look lives IN that component; crudview holds only the
  layout + slide. Adjacent `RawRule`s need their own `;`.
