---
message: "feat: SVG icon harness migration (platformd + crudview)"
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Master plan: https://github.com/tinywasm/tinywasm/blob/main/docs/SVG_ICON_HARNESS_MASTER_PLAN.md
> Repo rules: `AGENTS.md` at this repo's root.
>
> **GATE:** requires `tinywasm/svg` already split into `svg` (Icon reference)
> + `svg/sprite` (definition) — published. Update `go.mod` first; if
> `github.com/tinywasm/svg/sprite` does not exist as an import path, STOP and
> report.

## Context (zero-context summary)

Isomorphic Go ecosystem: untagged files compile to backend AND browser
(TinyGo → WASM); `//go:build !wasm` files are backend/SSR only. **Every byte
of the WASM binary counts.**

`tinywasm/svg` is split by PACKAGE, not by build tag inside the library (a
tag-inside-library design was rejected because `tinywasm/ssr`'s extractor and
`assetmin` need sprite construction unconditionally, at all times — a build
tag can't express "always needed by this one backend consumer"):

- `github.com/tinywasm/svg` — shared reference: `const iconX = svg.Icon("x")`;
  render with `iconX.Render(class)` → `<svg aria-hidden="true" focusable="false" class="..."><use href="#x"/></svg>`.
  Safe to import from ANY file, tagged or not.
- `github.com/tinywasm/svg/sprite` — backend-only definition:
  `sprite.Define(iconX, viewBox, sprite.Path(...))` in a `//go:build !wasm`
  file, returned by `IconSvg() *sprite.Sprite`. The SSR pipeline extracts and
  injects the sprite inline into `<body>`.

**Important:** `svg/sprite` has no build tag of its own — it compiles fine for
wasm too (pure Go). The `!wasm` tag on YOUR `svg.go` is what keeps it out of
the browser bundle; forgetting it does not fail the build, it silently grows
the binary. This plan's Stage 5 leak-check is mandatory, not optional.

Current violations in this repo:

1. **`platformd/svg.go` has NO build tag** → three multi-hundred-byte SVG path
   strings (`home`, `products`, `info`) plus the old sprite machinery ship
   inside the WASM binary today. This is the single biggest measurable win.
2. **`platformd` carries icon ids as plain `string`** through the `UIModule`
   interface (`IconID() string`, `platformd.go:45`; `factory.go`;
   `web/client.go:24`) and hand-builds the reference at `platformd.go:204-210`
   with `Svg().Child(Use().Attr("href", "#"+iconID))`.
3. **`crudview/svg.go` is correctly tagged**, but because the old API couldn't
   cross the tag boundary, `crudview.go:293-297` hand-builds
   `svg.Use().Attr("href", "#"+id)` from raw strings.

## Stages

### Stage 1 — `platformd`: shared references, tagged definitions via `sprite`

1. In an **untagged** file of package `platformd` (the file that today holds
   exported symbols, or `platformd.go` next to its other constants), declare:

```go
const (
	IconHome     = svg.Icon("home")
	IconProducts = svg.Icon("products")
	IconInfo     = svg.Icon("info")
)
```

   Import `github.com/tinywasm/svg` here (safe from untagged code). They stay
   **exported** — consumers pass them to `NewUIModule`.

2. Rewrite `platformd/svg.go` to:

```go
//go:build !wasm

package platformd

import "github.com/tinywasm/svg/sprite"

// IconSvg registers the default platform navigation icons.
func (p *Platform) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		sprite.Define(IconHome, "0 0 576 512", sprite.Path(/* existing home d-string, verbatim */)),
		sprite.Define(IconProducts, "0 0 448 512", sprite.Path(/* existing products d-string, verbatim */)),
		sprite.Define(IconInfo, "0 0 16 16", sprite.Path(/* existing info d-string, verbatim */)),
	)
}
```

   **Copy the three existing `d` path strings verbatim** from the current
   `svg.go` (`M280.37 148.26...`, `M350.85 129...`, `m7 11h2v2h-2z...`). Do not
   re-draw, trim, or reformat them. Note the file KEEPS its `//go:build !wasm`
   tag — only the import (`svg` → `svg/sprite`) and function bodies
   (`svg.Define`/`svg.NewSprite` → `sprite.Define`/`sprite.NewSprite`) change.

### Stage 2 — type the icon through the `UIModule` chain

Replace `string` icon ids with `svg.Icon` end to end:

- `platformd.go:45` interface: `IconID() string` → `Icon() svg.Icon`
  (keep the comment: chassis renders via the sprite).
- `factory.go`: `NewUIModule(id, label, iconID string, view Component)` →
  `NewUIModule(id, label string, icon svg.Icon, view Component)`; field
  `iconID string` → `icon svg.Icon`; method `IconID()` → `Icon() svg.Icon`.
- `web/client.go:24` (wasm stub `mod`): same rename, `Icon() svg.Icon`.
- `platformd.go:204-210`: replace the hand-built block:

```go
// BEFORE
iconID := m.IconID()
if iconID != "" {
	link.Child(Svg().
		Attr("aria-hidden", "true").
		Attr("focusable", "false").
		Set(ClsNavIcon.AsAttr()).
		Child(Use().Attr("href", "#"+iconID)))
}
// AFTER
if icon := m.Icon(); icon != "" {
	link.Child(icon.Render(string(ClsNavIcon)))
}
```

Acceptance: `grep -rn "IconID" --include='*.go' .` → empty.

### Stage 3 — `crudview`: typed references via `sprite`

1. In **untagged** `crudview.go`, next to its other constants, declare
   (unexported — only this package uses them), importing
   `github.com/tinywasm/svg`:

```go
const (
	iconCrudNew    = svg.Icon("icon-crud-new")
	iconCrudDel    = svg.Icon("icon-crud-del")
	iconCrudCancel = svg.Icon("icon-crud-cancel")
	iconCrudSave   = svg.Icon("icon-crud-save")
	iconCrudSearch = svg.Icon("icon-crud-search")
)
```

2. In `crudview/svg.go` (already tagged `//go:build !wasm`), change the
   import from `github.com/tinywasm/svg` to `github.com/tinywasm/svg/sprite`,
   and change each `svg.Define("icon-crud-new", "0 0 16 16", ...)` to
   `sprite.Define(iconCrudNew, "0 0 16 16", ...)` etc. Keep every `d` string
   and the Pa100T provenance comments verbatim.

3. In `crudview.go:293-297`, replace the hand-built
   `svg.Svg()...Child(svg.Use().Attr("href", "#"+id))` helper with
   `icon.Render(...)` taking an `svg.Icon` parameter instead of a raw `id
   string`. Update every caller of that helper to pass the new constants.
   If callers receive the id from data rather than code, STOP and report —
   that is a design decision, not yours to make.

Acceptance: `grep -rn '"#' --include='*.go' . | grep href` → empty;
`grep -rn 'svg.Svg()\|svg.Use()' --include='*.go' .` → empty.

### Stage 4 — AGENTS.md

Add to `AGENTS.md` (mirroring `tinywasm/components/AGENTS.md`) two sections:
**"Build tags belong to the consumer"** (table: untagged = shared/ships to
WASM; `wasm` = browser only; `!wasm` = backend/SSR only — includes the note
that `tinywasm/svg` itself never uses build tags, the split is by package) and
**"SVG icons — name is shared, drawing is backend-only"** (const reference in
untagged code from `svg`, `sprite.Define` in tagged `svg.go`, `Icon.Render` as
the only render path, and the explicit warning that `svg/sprite` compiles for
wasm too so forgetting the tag does NOT fail the build — only the mandatory
`go list -deps | grep tinywasm/svg/sprite` check catches it).

### Stage 5 — tests and mandatory leak-check verification

- Update tests asserting old markup/ids; keep coverage equivalent.
- `gotest` (never `go test`).

```bash
GOOS=js GOARCH=wasm go build ./...
GOOS=js GOARCH=wasm go list -deps ./platformd | grep tinywasm/svg/sprite   # MUST be empty
GOOS=js GOARCH=wasm go list -deps ./crudview | grep tinywasm/svg/sprite    # MUST be empty
grep -rn "IconID" --include='*.go' .                                        # empty
grep -rn 'svg.Svg()\|svg.Use()' --include='*.go' .                          # empty
```

The `go list -deps | grep` commands are the substitute for a compile-time
guarantee — `svg/sprite` has no tag of its own, so only this dependency-graph
check proves the wasm build never reaches it. Do not skip.

## Anti-footguns (do NOT do)

- Do NOT modify the SVG path `d` strings — copy verbatim.
- Do NOT touch `crudview/svg.go`'s build tag — it is already correct; only
  its import and function calls change (`svg` → `svg/sprite`).
- Do NOT add stdlib imports to untagged files (`tinywasm/fmt` only).
- Do NOT rename the symbol ids (`home`, `icon-crud-new`, ...) — downstream
  CSS and consumers reference them.
- This is a breaking change to `NewUIModule`/`UIModule` — do NOT add
  compatibility overloads; downstream (`veltylabs/mjosefa-cms`) migrates via
  its own plan.
- Never run `gopush` or `codejob`.

## Stages table

| # | Stage | Files | Done |
|---|---|---|---|
| 1 | platformd refs + tagged defs via sprite | `platformd/platformd.go`, `platformd/svg.go` | ☐ |
| 2 | Typed UIModule chain | `platformd/platformd.go`, `platformd/factory.go`, `platformd/web/client.go` | ☐ |
| 3 | crudview typed refs via sprite | `crudview/crudview.go`, `crudview/svg.go` | ☐ |
| 4 | AGENTS.md sections | `AGENTS.md` | ☐ |
| 5 | Tests + mandatory leak checks | `*_test.go` | ☐ |
