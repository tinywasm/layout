---
PLAN: "feat(crudview): always-visible row-count chip on the list card"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 1460330547499299982
PR: https://github.com/tinywasm/layout/pull/32
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> **DRAFT** — the maintainer will refine placement/style. Implement the wiring
> exactly as specified.

# Plan — a row-count chip on the crud list card

A `crudview` list gives no total count — the user has to eyeball / scroll to
know "how many". Add a small count that rides the **top-end corner of the list
card**, always visible (normal mode too, not only selection mode), reflecting
`v.list.Count()` after every reload / filter.

It lives on the **list card**, not on a footer button and not on the filter
control, because the count is a property of the list and must read the same
whether the filter is a search bar, a calendar (`reservation` module), or
absent. Reuse `components/countbadge.CountBadge` — the same out-of-flow bubble
the footer's `🗑 N` / `✏ N` already use — so there is one count idiom.

## Ecosystem rules (this repo)

- `github.com/tinywasm/layout` is a public library — code/comments/errors in
  **English**.
- **No Go stdlib** in WASM-reachable files: `crudview.go` uses
  `github.com/tinywasm/fmt` — `fmt.Sprintf("%d", n)`. Do NOT add
  `strconv`/`strings`/`errors`.
- Embed `dom.Element` by value. `switch`, not `map`.
- Tests: `gotest`, never `go test`.

---

## Stage 1 — `crudview/crudview.go`

`countbadge` and `fmt` are already imported.

### 1a. New signal field

In the `CrudView` struct `// internal` block, right after `hasMultiRows`
(around line 158):

```go
	// rowCount is the list's current visible-row count as text, for the
	// count chip on the list card (see Render). Set in filter() on every
	// reload/search — always current, shown in every mode.
	rowCount *SignalString
```

### 1b. Init

In `Init`, next to `v.hasRows = NewBool(false)` (around line 181):

```go
	v.rowCount = NewString("0")
```

### 1c. filter()

In `filter()`, next to the existing `hasRows` / `hasMultiRows` sets (around
line 563):

```go
	if v.rowCount != nil {
		v.rowCount.Set(fmt.Sprintf("%d", v.list.Count()))
	}
```

Keep the `hasRows` and `hasMultiRows` guards exactly as they are.

### 1d. Render() — the chip on the list card

The list card is built at (around line 703):

```go
		v.panel.Aside = Div().Set(clsListaBox.AsAttr()).Child(v.list)
```

Change it to render the count chip as the card's first child, sibling of the
list:

```go
		v.panel.Aside = Div().Set(clsListaBox.AsAttr()).
			Child((&countbadge.CountBadge{Count: v.rowCount, Visible: v.hasRows}).Render()).
			Child(v.list)
```

`Visible: v.hasRows` — the chip hides on an empty list (a "0" chip is noise,
same rule the footer badges follow). Do NOT gate it on selection mode — it is
a browsing aid.

---

## Stage 2 — `crudview/css.go`

The `countbadge` rides its host's top-end corner via `position: absolute`, so
the host must be a positioning context. Add `style.Anchor()` to the list part
(around line 43):

```go
		Part(widget.Part("list"),
			style.As(style.Inset),
			style.Pad(cardInset),
			style.Scroll(),
			style.Round(style.RadiusMd),
			style.Fill(),
			style.Anchor(), // positioning context for the row-count chip (countbadge, OnEdge)
		).
```

`Anchor()` is `position: relative` — zero visual change on its own; it is the
documented host contract for `countbadge` (see
`https://github.com/tinywasm/components/blob/main/countbadge/README.md`).

Nothing else in `css.go` changes. Do NOT touch `Part("action")` /
`Part("action-delete")` / `Part("footer")`.

---

## Stage 3 — tests

### `crudview/bulk_stylesheet_test.go`

Add:

```go
// The list card is a positioning context so the row-count chip (countbadge)
// resolves against it, not the page.
func TestListCardAnchorsTheRowCountChip(t *testing.T) {
	caller := &conformance.FakeCaller{}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} })
	v := &CrudView{Title: "CRUD", Presenter: p, Form: html.Div()}
	v.Init(&fakeCtx{})

	b := ruleBlock(v.RenderCSS().String(), ".crudview__list {")
	if !fmt.Contains(b, "position: relative") {
		t.Errorf(".crudview__list must be position: relative (Anchor), block:\n%s", b)
	}
}
```

### `crudview/bulk_test.go`

Add (uses the existing `setupBulkTest`, which seeds 3 devices):

```go
// The row-count chip reflects the visible count and updates on filter.
func TestRowCountChipTracksTheList(t *testing.T) {
	v, _ := setupBulkTest(t, true)

	if v.rowCount.Get() != "3" {
		t.Errorf("rowCount = %q, want \"3\" (3 seeded devices)", v.rowCount.Get())
	}

	v.search.Set("Device One")
	v.filter()
	if v.rowCount.Get() != "1" {
		t.Errorf("after filtering to one row, rowCount = %q, want \"1\"", v.rowCount.Get())
	}

	v.search.Set("no-such-device-xyz")
	v.filter()
	if v.rowCount.Get() != "0" {
		t.Errorf("empty list rowCount = %q, want \"0\"", v.rowCount.Get())
	}
	// the chip's Visible (v.hasRows) is false here, so "0" never renders
	if v.hasRows.Get() {
		t.Errorf("hasRows must be false for an empty list (chip hidden)")
	}
}
```

Run `gotest` — vet + tests + race + wasm all green. `bulk_wasm_test.go` is
unaffected (DOM-attribute assertions), re-run it.

---

## Acceptance criteria

```bash
go build ./...
gotest                                                 # all green
grep -n "rowCount" crudview/crudview.go                 # field + Init + filter + Render = 4 hits
grep -n "style.Anchor()" crudview/css.go                # present on the list part
```

Manual (app-demo `#devices`, dev server): the list card shows a small "15"
chip top-right; type in the search → the chip follows the filtered count;
clear the search → back to the full count; an empty result → no chip.

## Stages table

| # | File | Action |
|---|---|---|
| 1 | `crudview/crudview.go` | `rowCount *SignalString` (field + `Init` + `filter()`); render `countbadge` on the list card |
| 2 | `crudview/css.go` | `style.Anchor()` on `Part("list")` |
| 3 | `crudview/bulk_stylesheet_test.go`, `crudview/bulk_test.go` | anchor assertion + count-tracking test |
