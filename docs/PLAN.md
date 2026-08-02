---
PLAN: "refactor(layout): one skeleton — rightpanel absorbs the frame, crudview becomes a controller"
TAG: v0.2.0
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Stage M3 of a 3-repo change.** Master plan:
> `app-releases/docs/LAYOUT_UNIFICATION_MASTER_PLAN.md`.

# Plan — `layout`: master checklist

This is an **orchestrator**. The work is in three stage files, executed **in
order, one dispatch each**. Do not start a stage before the previous one is
green.

| Stage | File | What lands |
|---|---|---|
| 1 | [PLAN_STAGE_1_RIGHTPANEL.md](PLAN_STAGE_1_RIGHTPANEL.md) | `rightpanel` becomes the one correct skeleton: flow defect fixed, frame unified, `AsideFooter` slot, mobile master-detail |
| 2 | [PLAN_STAGE_2_CRUDVIEW.md](PLAN_STAGE_2_CRUDVIEW.md) | `crudview` stops rendering a grid and composes `rightpanel`; the filter becomes a slot; **the demo fills it** |
| 3 | [PLAN_STAGE_3_GUARDS.md](PLAN_STAGE_3_GUARDS.md) | tests that make the duplication impossible to reintroduce |

Rationale, evidence and the decision record: [ARQ_REFACTOR.md](ARQ_REFACTOR.md).

---

## The problem in one table

`rightpanel` (2026-04-15, 133 lines, zero logic) and `crudview` (2026-07-08, 500
lines Go + 240 CSS) are two implementations of the same skeleton:

| Concept | `rightpanel` | `crudview` |
|---|---|---|
| Two-column grid | `main`: `Split(SplitTwoThirds, Space2)` | `Root`: `Split(SplitTwoThirds, Space2)` |
| Main content | `article` | `article` / `fields` |
| Side column | `aside` | `aside` |
| Side content | `aside-content` | `aside-content` / `list` |
| Side controls | **`AsideControls`** — *"e.g. search + filter"* | `search` ← hardcoded |

After stage 2 there is **one** skeleton (`rightpanel`) and **one** controller
(`crudview`), with no overlap.

## A live defect this fixes

`rightpanel`'s flow is inverted today. Measured on the running demo at
`#mod1`:

```
.rp        Stack(SpaceNone)  → children main + aside stack VERTICALLY
.rp__main  Split(TwoThirds)  → children header + article sit SIDE BY SIDE

rp__header  top=42 left=764  w=498   ← the title
rp__article top=42 left=1270 w=256   ← the content, in a column beside it
```

The title and the body render as two columns. It is invisible in the demo only
because no module passes an `Aside`. **Stage 1 fixes it**; this is not a
cosmetic preference.

---

## Rules that apply to every stage

Restated here so no stage file has to be read together with another.

- **No Go stdlib in shared (untagged) files** — use `github.com/tinywasm/fmt`.
  `_test.go` and `//go:build !wasm` files are exempt; do **not** "fix" their
  imports.
- **`github.com/tinywasm/css` and `github.com/tinywasm/widget/style` may only be
  imported from `css.go`** (`//go:build !wasm`).
- **No `map`** in shared code — use a `switch` or a `[]fmt.KeyValue` scan.
  `_test.go` may use maps.
- **Value embed `dom.Element`**, never `*dom.Element`. TinyGo's GC pays twice
  for a pointer embed.
- **No `OnMount`.** State lives in signals; `Init(dom.Ctx)` runs once.
- **A consumer never re-creates an upstream symbol locally.** If something is
  missing in `widget`, `components`, `view`, `form` or `dom` — **stop and report
  it**, do not declare a local interface or copy the code. This rule is the
  reason this whole plan exists.
- **`layout` legitimately imports `components/*`.** `crudview` already depends on
  `targetlist` and `modaldialog`; that is assembly, not coupling.

## Upstream dependencies

Both must be **published before stage 2**. Stage 1 needs neither.

```sh
go doc github.com/tinywasm/widget Filterable                 # M1
go doc github.com/tinywasm/components/searchbar SearchBar    # M2
```

`go.mod` already carries `replace` directives pointing `components` and `widget`
at local checkouts, so local execution resolves them without a release.

## Visual contract for the whole plan

The demo's rendering at `http://localhost:8080/#crud` is the baseline. **The
crud module must look identical after stage 2** — same frame, same columns, same
control heights. Rules move between files; values do not change.

One deliberate exception, decided by the author: the **`rightpanel` root frame
becomes `As(Primary)`**, so `#mod1` and `#mod2` gain the blue frame `#crud`
already has. That is the only intended visual change in this plan, and it makes
the three modules agree.

## The demo is part of the work, not a follow-up

`platformd/web/` is this repository's only real consumer, and the harness treats
that as the publication gate:

> *"An API is not published until a consumer-shaped test, inside the library
> itself, proves it… if that test is awkward to write, the API is awkward to use,
> and you have found the defect before shipping it."*

So stage 2 ends with `client.go` filling the filter slot **explicitly** — not
leaning on `crudview.New`'s default. A slot that only its own default ever fills
has never been exercised by anyone.

And per [ROADMAP.md](ROADMAP.md), verification is done in the running app on
desktop **and** an emulated phone, in **both** themes. A stage that compiles and
passes unit tests but was never opened in a browser is not done.

## Definition of done (all three stages)

1. `gotest` green at the module root after each stage.
2. `grep -rn "Split(\|MasterDetail(" --include=css.go .` → hits in **exactly
   one** package (`rightpanel`).
3. `#crud` in the running demo is pixel-equivalent to the pre-refactor baseline;
   `#mod1`/`#mod2` show the title above the content (not beside it) inside a blue
   frame.
4. Mobile at 375×812 unregressed: one visible panel, swipe to the detail, the FAB
   and the floating title present.
5. `grep -n "searchbar.SearchBar" platformd/web/client.go` → one hit.
6. Screenshots (desktop + mobile, both themes) attached to each stage's PR.
