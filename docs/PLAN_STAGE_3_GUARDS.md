← [Stage 2](PLAN_STAGE_2_CRUDVIEW.md) | Master → [PLAN.md](PLAN.md)

# Stage 3 — make the duplication impossible to reintroduce

Read [PLAN.md](PLAN.md) first.

## Why this stage exists

`rightpanel` and `crudview` diverged into two skeletons over three months and
nothing failed. The demo kept working, the tests kept passing, and the
duplication was only visible to someone reading both `css.go` files side by side.

The construction harness names the mechanism:

> *"An API gap always surfaces at the **leaf**, where the agent has no authority
> to publish upstream — so it patches locally. Technical debt is then not an
> accident: the workflow guarantees it."*

Prose in a `docs/` file does not break that loop; a failing test does. This stage
adds three, modelled on `components/conformance_test.go`, which already parses
each `css.go` AST to reject `css.Raw`, colour literals and viewport units.

**This stage adds only tests.** If you change a non-test file, you have left the
plan — unless a guard legitimately fails, in which case see "When a guard fails"
below.

---

## 1. New file `conformance_test.go` at the module root

Package `layout_test` is wrong here — the guards read source files, not symbols.
Use package `layout` (the root package, where `layout.go` lives) with
`//go:build !wasm`, matching `components/conformance_test.go`.

Shared helper for all three guards:

```go
// eachStyleFile walks every css.go in the repository and hands its parsed AST to
// fn. Skips docs/, web/ and .git.
func eachStyleFile(t *testing.T, fn func(path string, node *ast.File)) {
	t.Helper()
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			switch info.Name() {
			case "docs", "web", ".git", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Base(path) != "css.go" {
			return nil
		}
		fset := token.NewFileSet()
		node, perr := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if perr != nil {
			return fmt.Errorf("failed to parse %s: %w", path, perr)
		}
		fn(path, node)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
```

---

## 2. Guard A — one owner for the frame

```go
// TestOnlyOneOwnerOfTheGrid is the test that would have made the
// rightpanel/crudview duplication impossible. A layout primitive that arranges
// a module's top-level panels belongs to exactly one package: the skeleton.
// A second package emitting one is a second skeleton, which is how the first
// duplication started and how the next one would.
func TestOnlyOneOwnerOfTheGrid(t *testing.T) {
	// Split and MasterDetail arrange a module's panels. Grid and Stack are NOT
	// listed: they are ordinary rhythm inside any widget and banning them would
	// be noise.
	owners := map[string][]string{} // primitive → packages that emit it

	eachStyleFile(t, func(path string, node *ast.File) {
		pkg := filepath.Dir(path)
		ast.Inspect(node, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			switch sel.Sel.Name {
			case "Split", "MasterDetail":
				owners[sel.Sel.Name] = appendOnce(owners[sel.Sel.Name], pkg)
			}
			return true
		})
	})

	for primitive, pkgs := range owners {
		if len(pkgs) > 1 {
			t.Errorf("%s is emitted by %d packages (%v); the module frame has ONE owner.\n"+
				"If a package needs this layout, it must compose the skeleton that owns it "+
				"(rightpanel), not re-declare the primitive. See docs/ARQ_REFACTOR.md.",
				primitive, len(pkgs), pkgs)
		}
	}
}
```

`appendOnce` is a two-line helper appending a string if absent — write it beside
the test. `map` is fine: this is a `_test.go` file.

⚠️ Match on `sel.Sel.Name`, not on the source text — an aliased import
(`st "github.com/tinywasm/widget/style"`) breaks text matching and would make the
guard silently pass.

---

## 3. Guard B — no locally-named seams

```go
// TestNoLocallyDeclaredSeams rejects the leaf-patch: a one-method interface
// declared here whose name matches a capability the ecosystem already names in
// widget/capability.go. Declaring it locally forks that contract, and the copy
// can never be reused — the failure mode described in docs/ARQ_REFACTOR.md §2.
func TestNoLocallyDeclaredSeams(t *testing.T) {
	// Method names that widget/capability.go already owns.
	upstream := map[string]string{
		"OnFilterChange": "widget.Filterable",
		"Select":         "widget.Selectable",
		"Dismiss":        "widget.Dismissible",
		"Expand":         "widget.Expandable",
	}
	// ... walk every .go file (not only css.go), find *ast.InterfaceType with
	// exactly one method, and fail when that method's name is a key above.
}
```

Write the walk yourself following `eachStyleFile`'s shape but without the
`css.go` filter, and skipping `_test.go` files — a test may legitimately declare
a stub interface.

The failure message must name the replacement:

```
crudview/crudview.go declares a local interface with the single method
OnFilterChange; use widget.Filterable instead. A consumer never re-creates an
upstream symbol locally (construction harness, "lego rules").
```

⚠️ `Select(id string)` also appears on `view.Presenter` — but that is a
**method**, not a one-method **interface declaration**, so this guard does not
see it. Confirm with `gotest` that the guard is green on the current tree before
believing it works; a guard that has never failed and never passed is untested.

---

## 4. Guard C — every skeleton slot is reachable

```go
// TestEverySlotIsRendered guards the third failure mode: a slot added to the
// struct but never wired into Render(), which fails silently — the consumer
// sets it and nothing appears. Every exported Component field on RightPanel
// must show up in the rendered markup when populated.
func TestEverySlotIsRendered(t *testing.T) {
	// Reflect over rightpanel.RightPanel's exported fields of type
	// dom.Component, populate each with a uniquely-marked stub, render once,
	// and assert every marker is present in the output.
}
```

Use `reflect` (this is a `!wasm` test file — `reflect` is allowed here and
**must not** be "fixed" into something else). One stub per field, marker
`data-slot='<FieldName>'`. Any field whose marker is missing fails with:

```
RightPanel.AsideFooter is declared but never rendered: a slot that cannot be
seen is a silent failure (construction harness, principle 6).
```

This is the guard that would have caught `AsideControls` sitting unused for three
months while `crudview` hardcoded a search bar beside it.

---

## 5. Documentation

Update [ARCHITECTURE.md](ARCHITECTURE.md) — its "Package Layout" tree still lists
only `platformd` and `rightpanel`, which is how the duplication stayed invisible.
Replace that tree with the real one and one line each:

```
    tinywasm/layout/
    ├── platformd/      # Shell: header, nav rail, hash routing, notifications
    ├── rightpanel/     # THE module skeleton: frame, two columns, aside bands,
    │                   # mobile master-detail strip. Owns every layout primitive.
    └── crudview/       # CRUD controller: state machine + orchestration.
                        # Renders NO frame — composes rightpanel.
```

Add a short section "Who owns what" restating the two-line rule, and link
[ARQ_REFACTOR.md](ARQ_REFACTOR.md).

Update [ROADMAP.md](ROADMAP.md): the line *"**search**: plain `<input
type=search>` inside crudview (explicit decision, do not swap in
`selectsearch`)"* is now false. Replace it with:

```markdown
- **filter slot**: `rightpanel.AsideControls` takes any `widget.Filterable`.
  `crudview.New` installs a `components/searchbar.SearchBar` by default; a
  calendar or a select replaces it without touching either package.
  (Supersedes the v0.1 decision to hardcode a plain input.)
```

---

## When a guard fails

A guard failing on the current tree means either the guard is wrong or the tree
is. **Do not weaken a guard to make it pass.** Report which, with the failing
output, and stop. Adding an exception list is the same defect the guards exist to
catch, one level up.

---

## Stages table

| # | File | What lands |
|---|---|---|
| 1 | `conformance_test.go` (new) | `eachStyleFile` helper |
| 2 | same | Guard A — one owner for `Split`/`MasterDetail` |
| 3 | same | Guard B — no locally-declared seams |
| 4 | same | Guard C — every slot rendered |
| 5 | `docs/ARCHITECTURE.md`, `docs/ROADMAP.md` | the map matches the code |

---

## Definition of done

1. `gotest` green at the module root.
2. Each guard is proven to bite. For each, temporarily introduce the violation it
   targets, confirm it fails with the intended message, then revert. State in the
   PR description that this was done — a guard nobody has seen fail is a guard
   nobody knows works.
3. `grep -n "crudview" docs/ARCHITECTURE.md` → non-empty.

## Out of scope

- **Any change to non-test Go files.** Stages 1 and 2 did that work.
- **Guards over `platformd`.** The shell is a third thing (routing, chrome) and
  is not part of the skeleton/controller split.
- **A CI workflow.** `gotest` at the module root already runs these; wiring a
  pipeline is separate.
