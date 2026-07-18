---
PLAN: "feat: migrate crudview to view v0.2.0, self-wired CRUD behavior, conformance adoption, fix platformd demo"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 11691698272361580726
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Self-contained: the executing agent has zero prior context.
> Orchestrated by `tinywasm/app` → `docs/CRUD_VIEW_CONFORMANCE_MASTER_PLAN.md`
> (https://github.com/tinywasm/app/blob/main/docs/CRUD_VIEW_CONFORMANCE_MASTER_PLAN.md).
> **Phase B**: requires `tinywasm/view v0.2.0` published (Phase A). Do not start before.

# Plan — layout: migrate crudview to view v0.1.2 and fix the platformd demo

## Context (why this plan exists)

Running the `platformd` demo (`platformd/web/client.go`) panics in the browser
at startup:

```
panic: form.New: form has no renderable field — every Field.Type is a plain
model.Kind, not a form input.Input. Declare the widget in the model Definition
(input.Text(), input.Number(), …) instead of model.Text()/model.Int()
  main.mod.View  platformd/web/client.go:70
  platformd.(*Platform).Render  platformd/platformd.go:238
```

Root cause: the demo's `mockModel.Schema()` declares `{Name: "id", Type:
model.Text()}` — a plain data Kind with no form widget. `crudview.New` →
`form.New(record)` correctly fails loudly, and the demo turns that error into a
panic. The fix is to declare widgets (`input.Text()`) as the harness intends.

Additionally this repo pins `view v0.1.0` (old API: `CanSave/CanDelete/Save/
Delete` inside `Presenter`). `view v0.2.0` is a breaking refactor:

```go
// view v0.2.0 — the API this plan migrates TO
type Presenter interface {
    Title() string
    SearchPlaceholder() string
    Record() model.Model
    Items() []Item
    Filter(term string) []Item // local case-insensitive match over Label+Description; "" returns all
    Reload() error
    Selected() string
    Select(id string) model.Model // unknown id → nil, selection unchanged
    Deselect()
}
type Saver interface { Save(payload model.Model) error }     // capability, by type assertion
type Deleter interface { Delete(id string) error }           // capability, by type assertion
```

Finally, the current `CrudView` API has a harness hole: `New()` wires the CRUD
behavior into PUBLIC callback fields (`OnSelect`, `OnSave`, …), so a consumer
that assigns its own callback **silently destroys the wiring** — the demo does
exactly this (`cv.OnSelect = toast` replaces the form-fill logic). This plan
moves the behavior into the component (unexported), leaving public fields as
purely ADDITIVE hooks.

## Constraints

- **No standard library** in any package here (all may compile to WASM): use
  `github.com/tinywasm/fmt` (`Convert`, `Contains`, `Errf`) — never `strings`,
  `errors`, `strconv`.
- Embed `dom.Element` **by value** (already done — keep it).
- No `internal/` folders. No local re-declaration of interfaces that exist in
  `view` — assert `view.Saver` / `view.Deleter` directly.
- CSS stays in `css.go`, SVG in `svg.go` (`//go:build !wasm` splits untouched).
- Unexported test seams are allowed: tests live in the same package.

## Stage 1 — Dependency bump

Files: `go.mod`, `go.sum`.

1. `go get github.com/tinywasm/view@v0.2.0`, then `go mod tidy`.
2. Expect compile failures in `crudview` — fixed in Stages 2–3.

## Stage 2 — crudview: self-wired behavior (`crudview/crudview.go` + `crudview/crud.go`)

### 2a. Struct changes (`crudview.go`)

```go
type CrudView struct {
    Element // value embed — NEVER *dom.Element

    Title             string
    Form              Component      // what Render paints (may stay nil in standalone mode)
    Presenter         view.Presenter
    SearchPlaceholder string

    // Additive user hooks — called AFTER the built-in behavior. Assigning them
    // can never disable list→form fill, save or delete wiring.
    OnSelect  func(it view.Item)
    OnNew     func()
    OnSaved   func(err error)
    OnDeleted func(id string, err error)
    OnCancel  func()

    // internal
    form     *form.Form // typed handle set by New; nil when standalone
    items    *SignalNodes
    selected *SignalString
    search   *SignalString
    canSave  *SignalBool
    canDelete *SignalBool
}
```

Delete the old public fields `OnSave func(done func(err error))` and
`OnDelete func(id string, done func(err error))` — they are replaced by the
result-shaped hooks `OnSaved` / `OnDeleted`. Acceptance:
`grep -n "OnSave " crudview/crudview.go crudview/crud.go` → no matches with the
old `done func` signature.

### 2b. Behavior methods (unexported, same file — they are the test seam)

```go
// selectAction: card click / driver Select
func (v *CrudView) selectAction(it view.Item) {
    v.selected.Set(it.ID)
    v.canDelete.Set(it.ID != "")
    rec := v.Presenter.Select(it.ID)
    if v.form != nil {
        _ = v.form.LoadValues(rec) // nil record → LoadValues resets; not an error
    }
    if v.OnSelect != nil { v.OnSelect(it) }
}

// newAction: "+" button
func (v *CrudView) newAction() {
    v.selected.Set("")
    v.canDelete.Set(false)
    v.Presenter.Deselect()
    if v.form != nil { v.form.Reset() }
    if v.OnNew != nil { v.OnNew() }
}

// cancelAction: "↺" button
func (v *CrudView) cancelAction() {
    if v.form != nil { v.form.Reset() }
    if v.OnCancel != nil { v.OnCancel() }
}

// saveAction: save button / driver Save. Only reachable when saver != nil.
func (v *CrudView) saveAction(saver view.Saver) {
    err := v.form.Validate()
    if err == nil {
        record := v.Presenter.Record()
        if err = v.form.SyncValues(record); err == nil {
            if err = saver.Save(record); err == nil {
                _ = v.Reload()
            }
        }
    }
    if err != nil { Log(err.Error()) }
    if v.OnSaved != nil { v.OnSaved(err) }
}

// deleteAction: delete button / driver Delete. Only reachable when deleter != nil.
func (v *CrudView) deleteAction(deleter view.Deleter, id string) {
    err := deleter.Delete(id)
    if err == nil {
        v.selected.Set("")
        v.canDelete.Set(false)
        _ = v.Reload()
    } else {
        Log(err.Error())
    }
    if v.OnDeleted != nil { v.OnDeleted(id, err) }
}
```

Wire `Render()` to these:

- Card click handler (inside `filter()`): replace the current body with
  `v.selectAction(it)`.
- Capability discovery at the top of `Render()`:
  ```go
  saver, _ := v.Presenter.(view.Saver)     // nil when not saveable
  deleter, _ := v.Presenter.(view.Deleter) // nil when not deletable
  ```
  Render the Save button only when `saver != nil` (click →
  `v.saveAction(saver)`), the Delete button only when `deleter != nil` (click →
  `v.deleteAction(deleter, v.selected.Get())`, guarded by non-empty id as
  today). New and Cancel buttons render whenever `hasSource` (click →
  `v.newAction()` / `v.cancelAction()`). Guard `v.Presenter != nil` before the
  assertions (standalone mode).
- Delete the public `Select(id string)` method; `selectAction` replaces it.

### 2c. Filtering delegates to the Presenter (`filter()`)

`view.Presenter.Filter(term)` now owns the matching (harness rule: *"the glue
is written once, in the library that owns it"*). Replace the body of `filter()`
so it iterates `v.Presenter.Filter(v.search.Get())` and ONLY builds DOM cards —
delete the local `Convert(...).ToLower()` / `Contains` matching block.
Acceptance: `grep -n "ToLower" crudview/crudview.go` → no matches.

### 2d. `crud.go` — `New()` shrinks

```go
func New(cfg Config) (*CrudView, error) {
    if cfg.Presenter == nil {
        return nil, fmt.Errf("crudview.New: Presenter is required")
    }
    f, err := form.New(cfg.ParentID, cfg.Presenter.Record())
    if err != nil {
        return nil, err // a record with no widgets fails HERE, loudly
    }
    f.HideSubmit()
    return &CrudView{
        Title:             cfg.Presenter.Title(),
        Form:              f,
        form:              f,
        Presenter:         cfg.Presenter,
        SearchPlaceholder: cfg.Presenter.SearchPlaceholder(),
    }, nil
}
```

All `cfg.Presenter.CanSave()` / `CanDelete()` calls disappear. Acceptance:
`grep -rn "CanSave\|CanDelete" crudview/` → no matches outside tests… and after
Stage 3, none at all.

## Stage 3 — Test migration (`crudview/consumer_test.go`, `crudview/crudview_test.go`, wasm variants)

Consumer-shaped tests now go through the REAL stack (harness rule): build the
presenter with `view.New` + `conformance.FakeCaller` instead of the hand-rolled
`fakePresenter`. Delete `fakePresenter` entirely.

1. Keep the `Device` model (it already declares `input.Text()` widgets). Add:
   ```go
   func (d *Device) Item() view.Item {
       return view.Item{ID: d.Id, Label: d.Name, Description: d.Ip}
   }
   ```
   and a `DeviceList` implementing `model.ModelSlice` exactly like
   `conformance.MockList` does (methods `IsNil`, `DecodeFields` no-op, `Schema`
   nil, `Pointers` nil, `Len`, `At`, `Append`).
2. Build presenters per test:
   ```go
   caller := &conformance.FakeCaller{ Reply: func(op string, into model.Decodable) { … } }
   p := view.New(caller, &Device{}, "device_list",
       func() model.ModelSlice { return &DeviceList{} },
       view.WithTitle("…"), view.WithSaveOp("device_save"), view.WithDeleteOp("device_delete"))
   ```
   Cases that need a Reload error use `FakeCaller{Err: …}`.
3. Re-express the existing cases against the new seams: `selectAction`,
   `saveAction` (get the saver via `p.(view.Saver)`), `deleteAction`,
   `v.OnSaved`/`v.OnDeleted` hooks instead of `done` callbacks. Case 1
   (widget-less record fails) keeps using a record whose schema is
   `model.Text()` — it must still error out of `New`.
4. Update `consumer_wasm_test.go` / `crudview_wasm_test.go` to the same API.
5. Delete `DeviceNoWidgets`'s `CanSave`-era scaffolding if any remains.

## Stage 4 — Adopt the view conformance suite

New file: `crudview/conformance_test.go` (package `crudview`, no build tag).

```go
func TestViewConformance(t *testing.T) {
    conformance.Run(t, conformance.Factory{
        New: func(t *testing.T, p view.Presenter) conformance.Driver {
            v, err := New(Config{ParentID: "conformance", Presenter: p})
            if err != nil {
                t.Fatalf("crudview.New: %v", err)
            }
            v.Init(&fakeCtx{})
            return conformance.Driver{
                Mount:  func() { _ = v.Reload() },
                Labels: func() []string { return cardLabels(v) },
                Select: func(id string) { v.selectAction(view.Item{ID: id}) },
                SetField: func(name, value string) { v.form.SetValues(name, value) },
                Save: func() {
                    if s, ok := p.(view.Saver); ok { v.saveAction(s) }
                },
                Delete: func() {
                    if d, ok := p.(view.Deleter); ok { v.deleteAction(d, v.selected.Get()) }
                },
            }
        },
    })
}

// cardLabels extracts the visible label of each rendered card. A card renders as
// <li class='…'>LABEL<span class='…'>DESC</span></li>; the label is the text
// between the first '>' and the following '<' of el.String().
func cardLabels(v *CrudView) []string {
    els := v.items.Get()
    labels := make([]string, 0, len(els))
    for _, el := range els {
        s := el.String()
        start := fmt.Index(s, ">") // if fmt lacks Index, scan bytes with a for loop
        end := -1
        for i := start + 1; i < len(s); i++ {
            if s[i] == '<' { end = i; break }
        }
        if start >= 0 && end > start {
            labels = append(labels, s[start+1:end])
        }
    }
    return labels
}
```

Anti-footgun: `Select` passes only the ID because `selectAction` re-resolves
the record from the Presenter — the Item's Label/Description are not needed for
selection semantics. Do not look items up by label.

Acceptance: `go test ./crudview -run TestViewConformance -v` → every clause of
`conformance.Run` passes.

## Stage 5 — platformd demo (`platformd/web/client.go`)

Rewrite the CRUD part of the demo over the real stack. Delete `mockModel`,
`mockPresenter` and `mockCaller` (note: `mockCaller` implements a Caller
signature that no longer exists — it is dead code).

1. Demo model (same shape as the test `Device`, widgets included):
   ```go
   import "github.com/tinywasm/form/input"

   var deviceDef = model.Definition{
       Name: "device",
       Fields: model.Fields{
           {Name: "id", Type: input.Text()},
           {Name: "name", Type: input.Text(), NotNull: true},
           {Name: "ip", Type: input.Text()},
       },
   }
   ```
   plus `Device` struct implementing `model.Model` + `Itemizer` + codec methods
   and `DeviceList` (`model.ModelSlice`) — copy the shapes from
   `crudview/consumer_test.go` Stage 3.
2. Demo caller — the app layer legitimately owns its transport double. Exact
   `router.Caller` contract to implement:
   ```go
   Call(op string, args model.Encodable, into model.Decodable, done func(err error))
   Dispatch(op string, args model.Encodable)
   ```
   `demoCaller.Call`: when `op == "device_list"`, fill `into.(*DeviceList)`
   with three canned devices ("Pc Administracion"/192.168.122.10, "Pc
   Ventas"/192.168.122.11, "Servidor Web"/192.168.122.20) via `Append()`;
   always finish with `done(nil)`. `Dispatch`: no-op.
3. Presenter + view:
   ```go
   pres := view.New(&demoCaller{}, &Device{}, "device_list",
       func() model.ModelSlice { return &DeviceList{} },
       view.WithTitle("Computadores"),
       view.WithSearchPlaceholder("Buscar..."),
       view.WithSaveOp("device_save"),
       view.WithDeleteOp("device_delete"),
   )
   cv, err := crudview.New(crudview.Config{ParentID: "crud", Presenter: pres})
   if err != nil { panic(err) }
   cv.OnSelect  = func(it view.Item) { m.p.Notify(Msg.Info, "Seleccionado: "+it.Label, 2000) }
   cv.OnNew     = func() { m.p.Notify(Msg.Info, "Nuevo", 2000) }
   cv.OnSaved   = func(err error) { if err == nil { m.p.Notify(Msg.Success, "Guardado", 2000) } }
   cv.OnDeleted = func(id string, err error) { if err == nil { m.p.Notify(Msg.Error, "Eliminado "+id, 2000) } }
   cv.OnCancel  = func() { m.p.Notify(Msg.Info, "Cancelado", 2000) }
   ```
   The hooks are now additive — assigning `OnSelect` no longer disables the
   list→form fill.

## Stage 6 — Verification

1. `go build ./...` and `go test ./...` green.
2. `GOOS=js GOARCH=wasm go vet ./crudview ./platformd/web` compiles (wasm-tagged
   tests and the demo build for the wasm target).
3. Manual (done by the human after release, via the tinywasm MCP): load the
   platformd demo and confirm `browser_get_errors` is empty and the CRUD panel
   renders form + list.

## Stages table

| # | Stage | Files | Gate |
|---|-------|-------|------|
| 1 | Bump view to v0.2.0 | `go.mod` | — |
| 2 | Self-wired crudview | `crudview/crudview.go`, `crudview/crud.go` | no `CanSave/CanDelete`, no `ToLower` in crudview |
| 3 | Test migration | `crudview/consumer_test.go`, `crudview/*_wasm_test.go`, `crudview/crudview_test.go` | `go test ./...` green |
| 4 | Conformance adoption | `crudview/conformance_test.go` (new) | `TestViewConformance` green |
| 5 | Demo rewrite | `platformd/web/client.go` | `GOOS=js GOARCH=wasm go vet ./platformd/web` clean |
| 6 | Verification | — | all of the above |
