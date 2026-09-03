---
PLAN: "feat(crudview): bulk-edit (✏) button appears only with 2+ rows"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Independent polish — not a gate for any other plan. Noticed while testing
> the demo: a one-row list showed the ✏ bulk-edit button, which never makes
> sense.

# Plan — the `✏` bulk-edit button requires at least 2 rows

## Context

`layout@v0.2.6` already gates the three `crudview` footer affordances by
capability and collapses them to what is actionable (empty list → only `+`;
composing a new record → only `↺`; a loaded existing row → `🗑` but not `✏`).

One case slipped: **with exactly one row in the list, `✏` still shows.** It
should not. `✏` is *always* a **bulk** operation — `view.Updater.Update(ids,
rec, fields)` patches a *set* of rows by the same delta; with a single row it
is just a worse way to do what tapping the row + editing the form already
does (single-record edit through `Saver.Save`, see
`docs/BULK_ACTIONS_MASTER_PLAN.md` §4). `🗑` is different and correctly stays:
`view.Deleter.Delete(ids...)` is variadic, N=1 is a real, common case
(delete this one row), and the normal-mode `🗑` on a loaded row opens that
row's own confirm dialog.

So: `🗑` visible with **≥1** row; `✏` visible with **≥2** rows.

## The change — `layout/crudview/crudview.go`

Two edits, both in `crudview.go`. No other file.

### 1. Add a `hasMultiRows` signal beside `hasRows`

In the `CrudView` struct (the `// internal` block, right after the `hasRows`
field and its comment, around line 152):

```go
	// hasMultiRows is hasRows' sibling for the ONE affordance that is
	// meaningless below N=2: the bulk-edit ✏. Delete (🗑) reads hasRows —
	// it is variadic and a single-row delete is a real case; ✏ patches a
	// SET by one delta, so a lone row would just be a worse single edit
	// (which the form already does — master plan §4).
	hasMultiRows *SignalBool
```

In `Init(ctx Ctx)`, next to `v.hasRows = NewBool(false)` (around line 165):

```go
	v.hasMultiRows = NewBool(false)
```

In `filter()` (around line 555), next to the existing
`v.hasRows.Set(v.list.Count() > 0)`:

```go
	if v.hasMultiRows != nil {
		v.hasMultiRows.Set(v.list.Count() > 1)
	}
```

Keep the existing `if v.hasRows != nil { v.hasRows.Set(v.list.Count() > 0) }`
exactly as it is — do not merge the two into one guard, they are independent
signals.

### 2. `✏` visibility reads `hasMultiRows` instead of `hasRows`

In `Render()`, the `btnEdit` construction (inside `if _, ok :=
v.Presenter.(view.Updater); ok {`), its `BindStateFunc(widget.Open, ...)`
currently reads (around line 787):

```go
			BindStateFunc(widget.Open, func() bool {
				if v.mode.Get() == string(modeEditing) {
					return true
				}
				return v.mode.Get() == string(modeNormal) && !v.active() && v.hasRows.Get()
			}).
```

Change the last line's `v.hasRows.Get()` to `v.hasMultiRows.Get()`:

```go
			BindStateFunc(widget.Open, func() bool {
				if v.mode.Get() == string(modeEditing) {
					return true
				}
				// ✏ is always a bulk patch — hidden below 2 rows (a single
				// row is edited by tapping it + the form, master plan §4).
				return v.mode.Get() == string(modeNormal) && !v.active() && v.hasMultiRows.Get()
			}).
```

**Do NOT touch** the `btnDelete` (`cv-cruddelete`) visibility — it must keep
reading `v.hasRows.Get()` (≥1). **Do NOT touch** the `btnEdit` `disabled`
derive (`!v.hasChecked.Get() || !v.hasEdits.Get()` in editing mode) — that is
about a marked-and-dirty precondition, unrelated to row count.

## Tests — `layout/crudview/bulk_test.go`

The test `TestFooterHidesBulkActionsWhenNothingToActOn` already covers
empty-list / composing / loaded-row. Add a one-row case to it, right after the
"Rows exist, nothing loaded: both bulk actions offered" block:

```go
	// Exactly ONE row: 🗑 stays (single delete is real), ✏ hides (bulk edit
	// needs a set).
	v.search.Set("Device One") // the seeded label — filters to a single row
	v.filter()
	if v.list.Count() != 1 {
		t.Fatalf("precondition: expected exactly 1 row, got %d", v.list.Count())
	}
	if !shown("cv-cruddelete") {
		t.Errorf("with one row 🗑 must still show (single delete is valid)")
	}
	if shown("cv-crudedit") {
		t.Errorf("with one row ✏ must hide — bulk edit needs 2+ rows")
	}
	v.search.Set("")
	v.filter()
```

(`shown(name)` and the `setupBulkTest` helper already exist in that file.)

Run `gotest` — vet + tests + race + wasm all green. `bulk_wasm_test.go` asserts
DOM attributes and is unaffected, but re-run it.

## Ecosystem rules (this repo)

- `github.com/tinywasm/layout` is a public library — code/comments/errors in
  **English**.
- **No Go stdlib** in WASM-reachable files: use `github.com/tinywasm/fmt`.
  `crudview.go` already follows this — do not add `strings`/`strconv`/`errors`.
- Embed `dom.Element` by value. `switch`, not `map`.
- Tests: `gotest`, never `go test`.

## Acceptance criteria

```bash
go build ./...                                    # clean
gotest                                            # all green (incl. wasm)
grep -n "hasMultiRows" crudview/crudview.go       # 3 hits: field, Init, filter; + 1 in Render's btnEdit bind = 4
grep -n "v.hasRows.Get()" crudview/crudview.go    # still present for cv-cruddelete, NOT for cv-crudedit
```

Manual (any crudview host, e.g. app-demo `#devices`): filter to a single row →
`🗑` visible, `✏` gone. Clear the filter (3+ rows) → both back.

## Stages table

| # | File | Action |
|---|---|---|
| 1 | `crudview/crudview.go` | add `hasMultiRows *SignalBool` (field + `Init` + `filter()`); `btnEdit` visibility reads it instead of `hasRows` |
| 2 | `crudview/bulk_test.go` | extend `TestFooterHidesBulkActionsWhenNothingToActOn` with the one-row case |
