---
PLAN: "refactor: crudview drops its row-count overlay; the list's own header owns the count"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Do NOT run `gopush` or `codejob`.
>
> **Dispatch order:** this plan depends on `tinywasm/components` shipping the
> selection header (`listselect.Header`) that the `target*` list widgets now
> render — see `components/docs/PLAN.md` Part 2. Dispatch this only once that
> is merged and published, and once this repo's `go.mod` resolves a
> `components` version that contains it.

# PLAN — `tinywasm/layout` (`crudview`)

## Context — the bug this is half of

Testing `#reservation` in the demo: the top-right corner of the list card is a
visual pile-up in selection mode. Two absolutely-positioned overlays claim the
same corner of `.crudview__list`:

- **crudview's** row-count bubble — a `countbadge.CountBadge` bound to
  `v.rowCount`, `OnEdge(EdgeTop, SideEnd)`, showing the total (`4`). Visible in
  every mode while `hasRows`.
- **the list widget's** select-all master check — `OnEdge(EdgeTop, SideEnd, …)`,
  showing a glyph + `k / N`. Visible in selection mode.

The `components` side of the fix (its own plan) moves the select-all control
**and the count** into a normal in-flow header strip inside the `target*`
widget, shown only in selection mode: `[select-all box] [k / N]`. That strip is
the single owner of "how many rows / how many selected".

**This plan removes crudview's competing overlay.** After both plans land there
is exactly one place a count appears: the list widget's own header, in
selection mode only. That matches the original request for this feature — *"show
the count, but only while selecting"* — so removing the always-on chip is
intentional, not a regression.

crudview keeps the **other** two `countbadge`s — the ones on the 🗑 and ✏
commit buttons (`Count: v.checkedCount`). Those are a different concern (a
notification bubble on an action button) and are not touched.

## Repo rules (from `AGENTS.md` + `CONSTRUCTION_HARNESS.md`)

- English only in code, comments, identifiers, errors.
- No Go stdlib in WASM-compiled files; `github.com/tinywasm/fmt`. Embed
  `dom.Element` by value. No generics. Signals only for reactive state.
- `Render()` + optional `Init(ctx)` only — no lifecycle hooks, no `mounted`
  guard.
- Hot reload is on — do not run `go build` to "apply" a change; a one-off
  `GOOS=js GOARCH=wasm go build -o /dev/null ./crudview/` is only a compile
  check.
- `gotest`, never `go test`. Every behaviour change needs a test in this repo.
- Sprite-leak check stays clean:
  `GOOS=js GOARCH=wasm go list -deps ./crudview | grep tinywasm/svg/sprite` →
  empty.

## Changes — `crudview/crudview.go`

1. **Remove the `rowCount` field.** In the `CrudView` struct, delete:

   ```go
   // rowCount is the list's current visible-row count as text, for the
   // count chip on the list card (see Render). Set in filter() on every
   // reload/search — always current, shown in every mode.
   rowCount *SignalString
   ```

2. **Remove its initialisation** in `Init`: delete the line
   `v.rowCount = NewString("0")`.

3. **Remove its update** in `filter()`: delete the block

   ```go
   if v.rowCount != nil {
       v.rowCount.Set(fmt.Sprintf("%d", v.list.Count()))
   }
   ```

   Leave the `v.hasRows` and `v.hasMultiRows` blocks around it exactly as they
   are — the footer's 🗑 / ✏ visibility still reads them.

4. **Remove the overlay from the aside.** In `Render`, the aside is currently:

   ```go
   v.panel.Aside = Div().Set(clsListaBox.AsAttr()).
       Child((&countbadge.CountBadge{Count: v.rowCount, Visible: v.hasRows}).Render()).
       Child(v.list)
   ```

   Make it just the list:

   ```go
   v.panel.Aside = Div().Set(clsListaBox.AsAttr()).
       Child(v.list)
   ```

   Update the comment above it (`// The list — its own inset card inside the
   aside's content band.`) if it mentions the chip.

5. **Do not touch** the two footer `countbadge.CountBadge{Count: v.checkedCount,
   Visible: v.hasChecked}` children on `btnDelete` and `btnEdit`. The
   `countbadge` import stays (still used there). `fmt` stays (used elsewhere).

## Changes — `crudview/css.go`

6. In `Part(widget.Part("list"), …)`, remove `style.Anchor()` and its comment
   `// positioning context for the row-count chip (countbadge, OnEdge)`. The
   list part no longer hosts an `OnEdge` child of crudview's, so it needs no
   positioning context. (The `target*` widget inside manages its own; that is
   its concern, not this sheet's.) Leave every other option on that part
   (`As(Inset)`, `Pad`, `Scroll`, `Round`, `Fill`) unchanged.

## Tests

7. `grep -rn "rowCount\|CountBadge\|countbadge" crudview/*_test.go` — for every
   hit tied to the **row-count** chip (not the checked-count footer bubbles):
   - a test asserting a `countbadge` / bound `v.rowCount` exists in the aside →
     delete that assertion (or the test if that was its whole point).
   - a test asserting the list part carries `Anchor()` / `position: relative`
     for the chip → delete it.
   - Tests for the footer commit-button bubbles (`v.checkedCount`,
     `v.hasChecked`) must stay green untouched.
8. If a conformance / render test counts children of the aside `Div`, adjust
   the expected count down by one (the chip child is gone).
9. `gotest ./...` green.

## Docs

10. `layout/docs/ARCHITECTURE.md` — if it describes a row-count chip on the
    crudview list card, replace with: the count is owned by the list widget's
    own selection header (`components/listselect.Header`), shown in selection
    mode only; crudview renders no count overlay of its own.
11. `layout/README.md` — keep every `docs/` file indexed (no new file here, but
    verify).
12. `layout/AGENTS.md` — no change expected; the "RevealedBy + an icon child →
    the part needs a flow" note and the capability-gated footer table are
    unaffected. Only edit if one of them mentions the row-count chip.

## Acceptance

- `grep -rn "rowCount" layout/` → empty.
- `grep -rn "Anchor()" crudview/css.go` → empty (it was there only for the
  chip).
- `grep -rn "CountBadge" crudview/crudview.go` → exactly the two footer
  bubbles on `btnDelete` / `btnEdit`, both `Count: v.checkedCount`.
- `gotest ./...` green; WASM build of `./crudview` clean; sprite-leak check
  empty.
- Manual (daemon hot-reloads `layout` + `components`): demo `#reservation`,
  pick a day → list fills, **no** count bubble in the card corner in normal
  mode. Press 🗑 → the list widget's in-flow header appears with `k / N`; the
  card corner has nothing floating over row 1. Leave selection mode → header
  gone, card corner clean.

## Stages

| # | Scope | Files |
|---|---|---|
| 1 | drop the overlay | `crudview/crudview.go` (field, Init, `filter()`, `Render` aside) |
| 2 | drop the positioning context | `crudview/css.go` (`list` part `Anchor()`) |
| 3 | tests | `crudview/*_test.go` |
| 4 | docs | `docs/ARCHITECTURE.md`, `README.md`, `AGENTS.md` (verify) |
