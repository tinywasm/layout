# PLAN — `tinywasm/layout`: migrate to `widget` v0.4.0

Execution document. Steps, reference code, test strategy. **Ephemeral**.

`layout` is the final consumer: it is the only repository in the suite that
breaks **twice** — once in its own three packages, and once inside its
`components` dependency. It is therefore the last to migrate and the first to
notice anything the others got wrong.

Upstream rename table, authoritative and not restated here:
[`widget/docs/MIGRATION.md`](https://github.com/tinywasm/widget/blob/main/docs/MIGRATION.md).

Every count below was measured by running
`go get github.com/tinywasm/widget@v0.4.0 && go mod tidy && go build ./...`
against this repository at `bc32558`.

---

## 🚦 0. Blocking gate — two of them

**Do not start until both have shipped.**

| Gate | Why | Check |
|---|---|---|
| `tinywasm/components` published with its v0.4.0 migration | `go build ./...` currently fails **inside `components@v0.2.0`**, not only here. No amount of editing this repository fixes that. | `go list -m -versions github.com/tinywasm/components` shows a version newer than `v0.2.0`, and `go build ./...` reports no error whose path contains `components@` |
| `tinywasm/ssr` published with per-producer `recover()` (E-7) | This migration can produce a sheet that compiles and then panics (§3). Without E-7 that surfaces as a stack trace inside generated code, naming nothing. | `grep -rn "recover()" "$(go env GOMODCACHE)"/github.com/tinywasm/ssr@*/invoke.go` prints a match for the newest version |

The second gate is a strong preference rather than a hard blocker: the migration
is possible without it, just much harder to debug. **Do not add a `replace`
directive to work around either.** Note the two commented-out `replace` lines
already in `go.mod` — leave them commented.

---

## 1. Scope and measured starting point

| | Count |
|---|---|
| Own packages failing to compile | 3 of 3 (`crudview`, `platformd`, `rightpanel`) |
| Dependency packages failing | 1 (`components/modaldialog`, and more behind it) |
| Types implementing `widget.Widget` | **0** — see §2 |
| `Hidden()` / `Shown()` calls to collapse | 10 |
| `style.On` call sites | 22 |
| Packages that will compile and then **panic** | up to 2 — see §3 |

**FORBIDDEN:**

| Prohibition | Reason |
|---|---|
| Changing a rule's intent | This is a migration. A part that was a `Panel` stays a `Panel`. |
| Adding a compatibility shim or alias | v0.4.0 has no deprecation period; recreating old names defeats it. |
| Uncommenting the `replace` directives in `go.mod` | They are dev-loop conveniences. Both gates must be satisfied by published versions. |
| Touching `//go:build !wasm` | Every `css.go` keeps it. |
| Reproducing an old shape with new names | Use the collapsed forms (`Interactive`, `RevealedBy`). |

---

## 2. ⚠️ The largest piece — nothing here implements `widget.Widget`

**This is not a rename, and it is the bulk of the work.**

`grep -rc "WidgetName\|WidgetKind"` over this repository returns **zero**. The
three packages build their sheets from a bare `Name` constant:

```go
const NamePlatform widget.Name = "pd"

func (p *Platform) RenderCSS() *css.Stylesheet {
    return style.Of(NamePlatform).
```

v0.4.0's entry point is `style.For(w widget.Widget)`, which requires:

```go
type Widget interface {
    WidgetName() Name
    WidgetKind() Kind
}
```

So each of `Platform`, `CrudView` and `RightPanel` gains two methods, and
**someone must choose three ARIA kinds** — a decision, not a rename. The `Kind`
is load-bearing in v0.4.0: it determines the ARIA role, the stacking level of any
overlay, and which states `Validate()` accepts.

```go
// crudview/crudview.go, platformd/platformd.go, rightpanel/rightpanel.go
func (p *Platform) WidgetName() widget.Name { return NamePlatform }
func (p *Platform) WidgetKind() widget.Kind { return widget.??? }
```

Recommended kinds, and why:

| Package | Name | What it is | Recommended `Kind` | Consequence |
|---|---|---|---|---|
| `platformd` | `pd` | application shell with a menu and a nav overlay | **`Menu`** | allows `Open`; `role="menu"`; overlays resolve to `--z-dropdown` |
| `crudview` | `crudview` | record view with new/cancel actions that reveal | **`Disclosure`** | allows `Open`; `role="group"`; `--z-dropdown` |
| `rightpanel` | `rp` | split detail panel, no states | **`Region`** | no states used, so any kind works; `Region` is the honest one |

`Region` is the zero value of `Kind`, so a package that genuinely has no
interaction states will validate under it. `platformd` and `crudview` will
**not** — see §3.

Keep the `Name*` constants: `Render()` still uses them for class attributes.
They stop being the sheet's entry point, that is all.

---

## 3. ⚠️ `Open` will panic under the wrong `Kind`

`platformd` and `crudview` both style `widget.Open`:

| Package | parts revealed by `Open` |
|---|---|
| `platformd` | `"menu"`, `"nav-overlay"` |
| `crudview` | `"action-new"`, `"action-cancel"` |

`Kind.Allows()` permits `Open` only for `Menu`, `Dialog`, `Disclosure` and
`Combobox`. Choose one of those for both packages or `Stylesheet()` panics:

```
widget/style: sheet pd: part "menu": state Open is not meaningful for kind Region
```

Because `ssr` calls `RenderCSS()` at build time, that aborts asset extraction for
the whole application — which is why §0's second gate matters.

The §2 recommendations already satisfy this. **If a different kind is chosen,
re-check it against `Allows()` before committing.**

### The `Hidden`/`Shown` collapse

Ten calls across the three packages. Each pair becomes one:

```go
// before — a pair split across two rules, plus an ordering rule to remember
Part(widget.Part("menu"), style.Hidden()).
…
When(widget.Open, widget.Part("menu"), style.Shown())

// after
Part(widget.Part("menu"), style.RevealedBy(widget.Open))
```

This also fixes a live defect inherited from v0.3.x: the old `Shown()` emitted
`display: block`, overriding the flow primitive's `display: flex` from an earlier
cascade layer. Any `Row` or `Stack` revealed this way lost its layout when it
opened. **If the CSS here contains a workaround for that, delete it.**

---

## 4. Mechanical rewrites

Counts are actual occurrences across the three `css.go` files.

| Old | New | Uses |
|---|---|---|
| `style.On(…)` | `style.As(…)` | 22 |
| `style.Of(Name)` | `style.For(receiver)` | 3 |
| `style.Space0` | `style.SpaceNone` | 4 |
| `style.Text(…)` | `style.FontSize(…)` | 4 |
| `style.Scrolls()` | `style.Scroll()` | 4 |
| `style.Accent` | `style.Primary` | 4 |
| `style.Sunken` | `style.Inset` | 2 |
| `style.RatioTwoThirds` | `style.SplitTwoThirds` | 2 |
| `style.Fixed()` | `style.KeepSize()` | 2 |
| `style.Hidden()` / `style.Shown()` | `style.RevealedBy(state)` — §3 | 10 |

Unchanged: `Pad`, `Round`, `Stack`, `Row`, `Split`, `Fill`, `Width`, `Animate`,
`Space1/2`, `Radius*`, `Text*`, `Weight*`, `Page`, `Panel`, `Content`,
`MotionSlow`.

### `Split` gains behaviour it never had

`crudview` and `rightpanel` both call `Split(RatioTwoThirds, Space2)`. In v0.3.x
that primitive **never collapsed** — it set `container-type` on the same selector
its `@container` rule targeted, and an element is never its own query container,
so the rule never applied. Measured in Chromium at 320px: two columns where one
was intended.

v0.4.0 replaces the mechanism with intrinsic sizing, so these two splits will
**start stacking below ~40rem for the first time**. That is the fix working. Look
at both at a narrow width before merging, and delete any hand-written CSS that
compensated for the old behaviour.

---

## 5. Two silent appearance changes

Neither produces a compile error.

**5.1 Surfaces now carry a radius.** Every `As(Panel)` gains
`border-radius: var(--radius-md)`; `As(Primary)`, `As(Inset)` and the rest gain
`--radius-sm`, unless the rule already overrides it. This repository has 22
surface applications against 6 explicit `Round()` calls, so roughly sixteen rules
gain a corner. Add `style.Round(style.RadiusNone)` wherever square is intended.

**5.2 The palette changed.** Step 1 pulls a newer `css` transitively, whose
contrast-corrected values change `--color-primary`, `--color-success` and
`--color-error`. The previous values failed WCAG AA at the colours that actually
rendered. Expect a brand-colour diff; it is not a regression.

Note `go.mod`'s commented `replace` for a local `css` checkout "with the brand
palette set to the Pa100T reference" — whoever maintains that palette should
re-run the contrast check against the new catalog before re-enabling it.

---

## 6. Implementation order

| # | Stage | Files | Gate |
|---|---|---|---|
| 0 | **Both gates in §0** | — | published versions, no `replace` |
| 1 | Decide the three `Kind`s (§2) | — | recorded in the PR description |
| 2 | Bump | `go.mod`, `go.sum` | `go mod tidy` clean; no `components@` error remains |
| 3 | `WidgetName()` / `WidgetKind()` on the three types | `*/crudview.go`, `*/platformd.go`, `*/rightpanel.go` | compiles |
| 4 | `rightpanel` — no states, simplest | `rightpanel/css.go` | compiles and emits |
| 5 | `crudview` — `Split` plus `RevealedBy` | `crudview/css.go` | compiles and **emits without panicking** |
| 6 | `platformd` — largest, two overlays | `platformd/css.go` | compiles and **emits without panicking** |
| 7 | Web demo and any `web/` assets referencing changed classes | `web/` | dev server renders |
| 8 | Docs | `docs/`, `README.md` | §8.4 empty |

Stages 5 and 6 are the real gates: their failure mode is a panic, not a compile
error.

---

## 7. Test strategy

| Test | Asserts |
|---|---|
| `TestEveryPackageEmits` | `RenderCSS()` on a **zero value** of each of the three types does not panic. This is what catches §3; a compile-only check does not. |
| `TestKindAllowsEveryState` | table-driven: every state passed to `When()` satisfies its package's own `Kind.Allows()` |
| `TestNoRemovedSymbols` | no `css.go` contains `style.On(`, `style.Of(`, `Hidden()`, `Shown()`, `Fixed()`, `Scrolls()`, `style.Accent`, `style.Sunken`, `RatioTwoThirds` |
| `TestSplitCollapses` | the emitted sheet contains no `container-type` and no `@container` — proof the §4 replacement landed |

Existing assertions that check emitted CSS substrings **will** need updating
where §4 and §5 change output. Change the expectation only after confirming the
new output is what those sections predict.

---

## 8. Acceptance criteria — grep-verifiable

1. `go build ./...` → clean, including every dependency path.
2. `gotest` green.
3. `grep -rn "style\.On(\|style\.Of(\|style\.Accent\|style\.Sunken\|Hidden()\|Shown()\|style\.Fixed()\|Scrolls()\|RatioTwoThirds" --include='*.go' .` → **empty**.
4. `grep -rn "style\.Of\|style\.On" docs/ *.md` → **empty**.
5. `grep -rl "func .*) WidgetKind() widget.Kind" --include='*.go' .` → **exactly 3 files**.
6. `grep -c "go:build !wasm" */css.go` → **1 for each of the 3**.
7. `GOOS=js GOARCH=wasm go list -deps ./...` contains **neither** `github.com/tinywasm/css` **nor** `github.com/tinywasm/widget/style`.
8. `grep -nE '^[[:space:]]*replace' go.mod` → **empty** (the two existing ones stay commented).
9. `go.mod` requires `widget v0.4.0` and a `components` version newer than `v0.2.0`.

---

## 9. Position in the suite

`layout` is last. The order is fixed by the dependency graph, not by preference:

```
css (done, on main)
  └── widget v0.4.0 (published)
        ├── form            → no plan needed; bump only, verified clean
        ├── components      → 8 packages, 1 panics
        │     └── layout    → 3 packages, up to 2 panic  ← this plan
        └── ssr             → 3 silent-drop paths + E-7, land early
```

`ssr` sits outside the chain and should land **first** despite not blocking
compilation, because every repository below it gains a failure mode — a
build-time panic — that `ssr` currently reports without naming the package
responsible.
