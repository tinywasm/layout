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
- [x] Mobile master-detail: horizontal scroll-snap; `--cv-detail-width: 90vw`,
      list 100vw, snap + ~10vw peek, slide left→right. Desktop keeps two columns.
      Below `(max-width: 640px)`, `.cv-module-content` becomes a
      `scroll-snap-type:x mandatory` flex row (gap/padding zeroed so the
      100vw+90vw math is exact); the form panel gets `flex:0 0
      var(--cv-detail-width, 90vw)` + `scroll-snap-align:end`, the list panel
      `flex:0 0 100vw` + `scroll-snap-align:start`. DOM order stays
      article-then-aside (desktop semantics unchanged) — `order` reverses the
      VISUAL sequence to list-first via CSS only. Forward navigation (row
      select → snap to form) needed a real DOM nudge, so `tinywasm/dom` (local
      `replace`) gained `Reference.ScrollIntoView()` (mirrors the existing
      `Focus()` — real `scrollIntoView({behavior:"smooth"})` in WASM, no-op in
      the backend/SSR stub); `crudview.selectAction` calls it on the form
      panel (`v.detailPanelID()`) whenever an existing row is selected. Swipe
      back to the list is native browser scroll-snap, no code. Verified live
      via MCP: mobile emulation (375×812) shows the list first; selecting a
      row snaps to the form with the list edge peeking; scrolling back to 0
      shows the list again with selection state intact; confirmed in both
      light and dark. (The synthetic `browser_swipe_element` tool didn't
      trigger a real scroll in this environment — verified the scroll
      mechanics directly via `element.scrollTo` instead; real touch swipe is
      standard browser behavior for `overflow-x:auto` + `scroll-snap-type`,
      not custom code.)
- [x] Delete confirmation via modaldialog. `⋮ → Eliminar` (`crudview.deleteRequest`)
      no longer deletes immediately — it looks up the row's label (from
      `targetlist.Items()`), stores it + the id in two internal signals, and
      opens a `modaldialog.ModalDialog` ("¿Eliminar «label»? Esta acción no se
      puede deshacer.", Cancelar / Eliminar). Only the modal's "Eliminar"
      button (`confirmDeleteAction`) actually calls `deleteAction`/closes the
      modal; Cancelar (or the modal's own backdrop/× close) just dismisses it,
      deleting nothing. `modaldialog.ModalDialog` (components) gained a
      `Close()` method (mirrors the existing `Open()`) since nothing let a
      host dismiss it programmatically before. The confirm content is built
      once in `Init` and reused across every ⋮ → Eliminar via two signals
      (`deleteID`, `deleteLabel`), not rebuilt per row. Verified with a new Go
      test (`TestConsumer_DeleteRequiresConfirmation`: 0 delete calls after
      opening/cancelling, exactly 1 after confirming) and live via MCP in both
      light and dark — confirmed the modal opens with the correct label,
      Cancelar leaves the row in the list, and confirming fires exactly one
      `device_delete` call. (The demo harness's `demoCaller.Call` only
      implements `device_list` — a static 3-item fixture — so the row
      visually reappears after a real reload regardless of the delete call;
      that's the fixture's limitation, not a crudview bug — the automated
      test is the source of truth for the actual delete call.)
- [x] Verify desktop + emulated mobile, light + dark. Done incrementally with
      each item above (MCP screenshots + `browser_evaluate_js` scroll checks)
      rather than as one final pass — every feature (read-only gating, mobile
      scroll-snap, delete confirmation) was confirmed in both breakpoints and
      both themes as it landed. Final idle-state sanity check: desktop light,
      982×690 — two columns, blank/editable "+" form, 3-row list, no console
      errors (`browser_get_errors` clean throughout the session).

## Post-roadmap fixes and additions
Found/requested after the checklist above was already complete; same
components, same conventions.

- **⋮ menu didn't close.** Native `<details>` only closes on a summary click or
  an explicit attribute removal — never on an outside click, and picking
  Editar/Eliminar didn't close it either. Fixed in `targetlist`
  (`components`): every row's `<details>` now shares one `name="tl-menu-group"`
  (native HTML "exclusive accordion" — opening one auto-closes any other, no
  JS); a new `closeAllMenus()` removes the `open` attribute from every row and
  is wired to the Editar/Eliminar click handlers AND to a full-page
  `.tl-menu-backdrop` click-catcher shown only while a menu is open (CSS
  `:has()`: `.tl-wrap:has(.tl-menu[open]) .tl-menu-backdrop { display:block }`).
  The backdrop had to become a plain sibling of the `<ul>` (a new `.tl-wrap`
  wrapper), never a static child mixed into it — `BindChildren`'s keyed
  reconcile assumes it owns every child of the element it's bound to, and
  fights a statically-added one.
- **Last row's dropdown got clipped.** The list container must clip overflow
  to scroll, so the last row's ⋮ menu — opening downward with no room below —
  got cut off. Fixed with `.tl-row:last-child .tl-menu-list { top:auto;
  bottom:calc(100% + 2px) }`, flipping just that row's menu to open upward.
- **Locked-field color too dark.** The read-only tint used `ColorMuted`
  (`#6E6E73` fallback) — a "disabled button" gray, too dark to read/select
  text through comfortably. Changed to `ColorSurface` (`#F2F2F7` fallback, a
  hair off white) and dropped the `Opacity(0.7)` dimming entirely (it faded
  the text too, which must stay fully legible/selectable — "frosted glass",
  not "grayed out"). `fieldset/css.go`.
- **"+"/Editar now focus the first field** — standard behavior, added to the
  `view/conformance` suite so every renderer must implement it identically,
  not just crudview. `tinywasm/form` gained `Form.Focus()` (moves focus to
  `Inputs[0]`'s DOM id via `dom.Get(id).Focus()` — imperative, not reactive:
  the form's DOM already exists by the time a host unlocks it) and
  `FocusedFieldID()` (returns the id last targeted — makes the INTENT
  observable in a backend/non-WASM test, since real focus movement is a
  WASM-only DOM side effect, a no-op in the SSR stub). `crudview.newAction`,
  `.undoAction`, and `.editAction` all call it after unlocking.
  `view/conformance.Driver` gained `New`, `Edit`, `FocusedFieldID` fields and
  two clauses (`new_focuses_first_field`, `edit_focuses_first_field`);
  `view/mock.Renderer` (the headless reference renderer) and
  `crudview`'s own conformance test both implement them. Required a new TEMP
  `replace github.com/tinywasm/view` in `layout/go.mod` (previously view had
  no local checkout wired in). Verified live via MCP + `browser_evaluate_js`
  (`document.activeElement`) in both the "+" and Editar paths.

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
- `replace` points dom, css, form, components, view at local checkouts.
- SRP for css: a component's look lives IN that component; crudview holds only the
  layout + slide. Adjacent `RawRule`s need their own `;`.
