# PLAN — crudview fixes: real `router.Caller` (async), SSR file convention, css swap, docs

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Base: the code delivered in PR #10 (branch
> `feature/crudview-and-gating-10777383567164381037`). If that PR is already
> merged, work on top of the default branch — it contains the same code. Do
> NOT restart the feature from scratch; everything listed as "keep" is
> verified working and must not be reworked.

## Review verdict of the previous delivery (context)

**Keep (verified green, do not touch):** `platformd.NewUIModule`
(`factory.go`), `Platform.CanView` with all three semantics (nav skip, panel
not built, `Activate` guard + `Init` fallback) and their tests; the crudview
structure (form left, list cards with description chip, CRUD bar with
per-callback buttons, local search, full-page variant on zero `Source`);
`gotest ./...` green.

**Broken / missing (this plan):**

1. **BLOCKING — crudview invented a private synchronous `Caller`:**

   ```go
   // crudview/crudview.go (WRONG — remove)
   type Caller interface {
       Call(op string, args model.Encodable) ([]byte, error)
   }
   ```

   The real, published contract (`tinywasm/router` ≥ v0.1.4 — the gate of this
   stage) is **asynchronous by callback**, because WASM fetch is async and the
   production adapter `mcp.NewCaller` implements exactly this:

   ```go
   // tinywasm/router (the ONLY Caller crudview may know)
   type Caller interface {
       Call(op string, args model.Encodable, callback func(result []byte, err error))
       Dispatch(op string, args model.Encodable)
   }
   ```

   With the private sync interface, crudview can never be wired to the real
   transport. Duplicating a published contract also violates the lego rule
   (reuse declared types, never re-declare them).

2. **SSR file convention violated (ssr discovery will miss the icons):**
   `IconSvg()` + `svg.Define(...)` live in `css.go`, while `svg.go` holds the
   CSS `Class` vars. The `tinywasm/ssr` pipeline discovers `IconSvg()` by
   regex **only in files named `svg.go`** (and `RenderCSS` in `css.go`) — as
   delivered, consumers' sprite generation will silently omit the CRUD icons.
3. **Icons are hand-invented placeholders**, not the captured Pa100T paths
   documented in [IE_MODULE_CONTENT.md](IE_MODULE_CONTENT.md) §2 (the four
   16×16 base64 SVGs: plus, minus, undo-arrow, save-disk). Also the
   `icon-16` class used by `renderIcon` is never defined in any stylesheet.
4. **css not swapped:** `go.mod` pins `tinywasm/css v0.1.3`; v0.1.4 (typed
   helpers) is published. All `// TODO(tinywasm/css)` `Decl{...}` literals in
   `crudview/css.go` must be swapped per
   [API_ANSWERS_CRUDVIEW.md](API_ANSWERS_CRUDVIEW.md) (that doc's exact
   mapping). Note two of those TODOs (`Position`/`Absolute`) already existed
   in v0.1.3 — swap them too.
5. **Hand-rolled `contains()` helper** duplicates `tinywasm/fmt` (`fmt.Contains`).
6. **Stage 5 docs skipped:** `README.md` and `docs/ARCHITECTURE.md` untouched
   (the demo `platformd/web/client.go` WAS updated — adapt it to the async
   Caller, don't rewrite it).

## Stage 1 — adopt the real `router.Caller` (async)

- Add `github.com/tinywasm/router` (latest tag) to `go.mod`; delete the
  private `Caller` interface; `Source.Caller` becomes `router.Caller`.
- Rework the load path to callback style:

  ```go
  func (v *CrudView) Reload() {
      if v.Source.Caller == nil { return }
      var args model.Encodable
      if v.Source.Args != nil { args = v.Source.Args() }
      v.Source.Caller.Call(v.Source.ListOp, args, func(raw []byte, err error) {
          if err != nil { v.handleError(err); return }
          if v.Source.Decode == nil { return }
          items, err := v.Source.Decode(raw)
          if err != nil { v.handleError(err); return }
          v.allItems = items
          v.filter()
      })
  }
  ```

  Save/Delete flows are unchanged (they already use the consumer's `done`
  callback); only the internal `Reload` becomes callback-driven. Signals make
  the late `filter()` safe — `BindChildren` patches when the reply lands.
- Tests: replace the local fake with the published mock in
  `tinywasm/router/mock` (canned bytes, inline callback). Assert the async
  path: reply delivered after `Init` returns still paints the list.
- Replace `contains()` with `fmt.Contains` (or `Convert(...).Contains` — use
  whatever `tinywasm/fmt` exposes; do not keep the local loop).

## Stage 2 — restore the SSR split convention

- `crudview/svg.go` (`//go:build !wasm`): `IconSvg() *svg.Sprite` with the
  **four captured Pa100T icon paths + the search magnifier** from
  IE_MODULE_CONTENT.md §2 (decode the documented base64 SVGs; keep the
  existing icon IDs `icon-crud-new/del/cancel/save/search`).
- Move the CSS `Class` vars from `svg.go` into `crudview.go` (neutral, needed
  by wasm — same placement platformd uses).
- `crudview/css.go` (`!wasm`): stylesheet only, no `svg.Define`. Define the
  icon sizing rule that `renderIcon` relies on (replace the undeclared
  `icon-16` string with a typed `Class` styled here).

## Stage 3 — css v0.1.4 swap

- Bump `go.mod` to `tinywasm/css v0.1.4` (or newer tag).
- Replace every `Decl{Prop: ..., Val: ...}` marked `// TODO(tinywasm/css)`
  with its typed helper (`PaddingTop`, `MarginTop`, `MarginBottom`,
  `FlexGrow`, `FlexWrap`, `AlignContent`, `Position(Relative)`,
  `Position(Absolute)`, `BackgroundSize/Position/Repeat`, keywords
  `SpaceBetween`, `InlineFlex`, `Wrap`, `FlexStart`, `NoRepeat`, …).
  `RawRule` stays only for the `::-webkit-scrollbar` rules.
- Acceptance: `grep -n "TODO(tinywasm/css)" crudview/` → empty.

## Stage 4 — Stage-5 docs (now against the correct API)

- `README.md`: add `crudview` to the index with the consumer example using
  `router.Caller`; document `NewUIModule` + `CanView`; add the
  **"Preconfigure, don't assemble"** section (the app's composition root wraps
  a layout once — e.g. a `config/layouts` package exposing `Crud(cfg)` — and
  feature modules call that wrapper).
- `docs/ARCHITECTURE.md`: `crudview` section — structure diagram
  (IE_MODULE_CONTENT.md §1), the `Source` seam (async), signal fields table,
  callback flow; plus the `CanView` activation-flow update.
- Adapt the demo `platformd/web/client.go` to the async Caller (its canned
  in-process fake calls the callback inline).

## Harness checklist (mandatory)

- Reuse declared contracts — `router.Caller` only; no private transport
  interfaces, no `any`/`map`/generics, no stdlib in wasm paths.
- Value embedding; typed signals; CSS/SVG strictly in `css.go`/`svg.go`
  (`!wasm`) per the ssr discovery convention.
- No string literals for classes/icon IDs in logic — typed `Class` vars and
  shared constants.
- Errors propagate via `OnError`; no silent failure paths.
- Do not touch `platformd` beyond the demo file.

## Acceptance criteria

1. `gotest ./...` green (native + browser); the wasm test proves the async
   round-trip through a `router.Caller` mock.
2. `grep -rn "type Caller" crudview/` → empty; `go.mod` has `tinywasm/router`
   and `tinywasm/css ≥ v0.1.4`; `grep -n "TODO(tinywasm/css)" crudview/` → empty.
3. `svg.go` owns `IconSvg()` (Pa100T paths); `css.go` owns only the
   stylesheet; icons render in the demo (screenshot in the PR, compared
   against IE_MODULE_CONTENT.md).
4. README + ARCHITECTURE updated as specified.
5. Existing platformd tests and public API untouched.

## Stages

| Stage | File(s) | Action |
|---|---|---|
| 1 | `go.mod`, `crudview/crudview.go`, `crudview/*_test.go` | real async `router.Caller`; mock from router/mock; drop `contains()` |
| 2 | `crudview/svg.go`, `crudview/css.go`, `crudview/crudview.go` | SSR split: IconSvg (Pa100T paths) → svg.go; Class vars → crudview.go; icon size class |
| 3 | `go.mod`, `crudview/css.go` | css v0.1.4 + swap all TODO Decls |
| 4 | `README.md`, `docs/ARCHITECTURE.md`, `platformd/web/client.go` | Stage-5 docs against the fixed API + demo adaptation |
