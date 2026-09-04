---
PLAN: "refactor!: layout crudview tests use typed view.Backend doubles instead of router.Caller"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Do NOT run `gopush` or `codejob`.
>
> **This is a BREAKING-change follow-up (Fase D).** `github.com/tinywasm/view`
> v0.3.0 replaced its transport seam (`router.Caller` + op strings) with a
> domain seam. Production code in this repo does NOT change (`crudview` works
> on `view.Presenter`, whose interface is untouched) — only the test doubles
> in `crudview/*_test.go` migrate. **Do not touch any other repo from here.**

# PLAN — `layout`: `crudview` tests to typed doubles (`view.Backend`)

## Why (read this before writing code)

`view.New` no longer takes a `router.Caller`. Its signature is now:

```go
func New(b Backend, record model.Model, opts ...Option) Presenter
```

Only `WithTitle` and `WithSearchPlaceholder` options still exist
(`WithSaveOp`/`WithUpdateOp`/`WithDeleteOp`/`WithArgs` are gone). The old test
double `conformance.FakeCaller` (a `router.Caller` with a `Reply` callback
that fills a `model.Decodable`) no longer exists either; it is replaced by
`conformance.FakeBackend`:

```go
type FakeBackend struct {
	Rows []model.Model // what List returns

	Calls         int
	SavedRecords  []model.Model
	UpdatedIDs    []string
	UpdatedFields []string
	UpdatedRecord model.Model
	DeletedIDs    []string
	Err           error
}
// implements view.Backend + BackendSaver + BackendUpdater + BackendDeleter
```

`FakeBackend` always carries every capability. Tests that assert the ABSENCE
of a capability (no `Updater` → no edit button) cannot use it — they use the
minimal double this plan defines in Stage 1.

## Repo rules

- Public library → **English** in code, comments, identifiers, errors.
- `crudview` compiles to WASM: test files with `//go:build wasm` use
  `github.com/tinywasm/fmt`, never stdlib `fmt`/`strings`. Untagged test files
  compile in BOTH worlds — a helper defined in `consumer_test.go` (untagged)
  is visible to the wasm tests too. Do NOT duplicate helpers per build tag.
- `gotest`, never `go test`. First prerequisite on a fresh machine:
  `go install github.com/tinywasm/devflow/cmd/gotest@latest`.
- `go.mod` already requires `github.com/tinywasm/view v0.3.0`. If it does not
  resolve, run `go get github.com/tinywasm/view@v0.3.0` first — then migrate.
- `crudview/conformance_test.go` and `crudview/crud.go` need NO change (the
  former only adapts an already-built presenter; the latter mentions
  `view.New` only in a doc comment). Do not touch them.

---

## Stage 1 — shared pieces in `crudview/consumer_test.go` (untagged)

`consumer_test.go` already defines the `Device` record (line ~39) and the
`DeviceList` slice (line ~67) used by every test in the package.

1. **Delete** the `DeviceList` struct, all its methods, and any
   `var _ model.ModelSlice = (*DeviceList)(nil)` line. Nothing references it
   after this plan.
2. **Add** (next to `Device`) the minimal double for capability-absence cases:

```go
// listSaveDeleteBackend implements view.Backend + BackendSaver +
// BackendDeleter, but NOT BackendUpdater: the double for every test that
// asserts update UI stays absent.
type listSaveDeleteBackend struct {
	rows []model.Model
	saved []model.Model
	deleted []string
}

func (b *listSaveDeleteBackend) List() ([]model.Model, error) {
	out := make([]model.Model, len(b.rows))
	copy(out, b.rows)
	return out, nil
}

func (b *listSaveDeleteBackend) Save(recs []model.Model) error {
	b.saved = append(b.saved, recs...)
	return nil
}

func (b *listSaveDeleteBackend) Delete(ids []string) error {
	b.deleted = append(b.deleted, ids...)
	return nil
}
```

3. Fix imports of `consumer_test.go` that become unused (`model` stays only if
   other code in the file still names it — check with the compiler, then
   `gotest`).

## Stage 2 — mechanical translation, every `view.New` + `Reply` site

Apply this translation to every test site listed in the table. The pattern:

```go
// before
caller := &conformance.FakeCaller{
	Reply: func(op string, into model.Decodable) {
		if op == "device_list" {
			dl := into.(*DeviceList)
			d1 := dl.Append().(*Device)
			d1.Id = "id-1"
			...
		}
	},
}
p := view.New(caller, &Device{}, "device_list",
	func() model.ModelSlice { return &DeviceList{} }, opts...)

// after
fb := &conformance.FakeBackend{
	Rows: []model.Model{
		&Device{Id: "id-1", ...},
		...
	},
}
p := view.New(fb, &Device{})
```

Rules: keep `view.WithTitle` / `view.WithSearchPlaceholder` where the site
already passes them (`consumer_wasm_test.go`, `crudview_wasm_test.go`); drop
`WithSaveOp`/`WithUpdateOp`/`WithDeleteOp` everywhere. An empty
`&conformance.FakeCaller{}` becomes `&conformance.FakeBackend{}`. A `Reply`
that only sets `reloaded = true` (`consumer_test.go` `TestConsumer_ListOp`)
becomes a `FakeBackend` and the assertion becomes `fb.Calls != 0`. Remove
imports (`model`, `view/conformance` where now unused) per file until each
file compiles; `gotest` is the backstop.

| File | Sites |
|---|---|
| `crudview/consumer_test.go` | ~8 `FakeCaller` sites (incl. `TestConsumer_ListOp`, `TestConsumer_ListRendersCards`, `TestConsumer_SaveWithFormData` — see Stage 3) |
| `crudview/consumer_stylesheet_test.go` | 6 sites (5 empty + 1 full-caps at ~line 229) |
| `crudview/bulk_stylesheet_test.go` | 5 empty sites |
| `crudview/crudview_test.go` | 1 empty site (~line 38) |
| `crudview/crudview_scope_test.go` | 1 `Reply` site (~line 19); keep its comment, rewording `FakeCaller` → `FakeBackend` |
| `crudview/compose_test.go` | `fakeListCaller()` → `fakeListBackend()` returning `*conformance.FakeBackend` with the same 3 rows; all 5 `view.New` sites use it |
| `crudview/consumer_wasm_test.go` | 1 empty site; keep `WithTitle` + `WithSearchPlaceholder` |
| `crudview/crudview_wasm_test.go` | 1 `Reply` site; keep `WithTitle`, drop `WithDeleteOp` (full caps come from the backend) |
| `crudview/rowclick_wasm_test.go` | 1 `Reply` site; drop `WithDeleteOp` |

## Stage 3 — behavioral assertions (wire → typed)

Three sites assert what the fake transport SAW; they now assert the typed
double's recordings:

1. `crudview/consumer_test.go` `TestConsumer_SaveWithFormData`: the
   `for _, c := range caller.Calls` loop hunting `op == "device_save"` plus
   `conformance.Payload(saveCall.Args)` / `conformance.Has(sent, "name", ...)`
   becomes: `fb.SavedRecords` must hold 1 record, asserted to `*Device` with
   `Name == "New Name"`. The comment about the envelope ("view ships batches
   inside its own envelope") dies with the code it described.
2. `crudview/bulk_wasm_test.go`
   `TestBulkDelete_ShipsEveryCheckedIDInOneCall`: `callsFor(caller,
   "device_delete")` + `Payload`/`Has` becomes `fb.DeletedIDs` containing
   exactly `["id-1", "id-3"]` (and NOT `"id-2"`). **Delete** the now-unused
   `callsFor` helper and the `conformance.FakeCall` reference.
3. `crudview/consumer_test.go` line ~321 (`len(caller.Calls) == 0`) and ~391,
   ~457 loops over `caller.Calls`: translate each to the matching typed field
   (`SavedRecords` for save ops, `DeletedIDs` for delete ops). Read each site
   before translating; the op name in the loop condition tells you which field.

## Stage 4 — capability-absence still absent

`setupBulkTest(t, withUpdate)` (`bulk_test.go`) and `mountBulk(t, withUpdate)`
(`bulk_wasm_test.go`) build `opts` with `WithDeleteOp` always and
`WithUpdateOp` when `withUpdate`. After the translation:

- `withUpdate == true` → `view.New(&conformance.FakeBackend{Rows: ...}, &Device{})`.
- `withUpdate == false` → `view.New(&listSaveDeleteBackend{rows: ...}, &Device{})`
  (no `Updater` → `TestEditButtonAbsentWithoutUpdater` keeps passing for the
  right reason: the backend has no update method, not a missing string).

The `Delete` capability is always present in both doubles, so every delete
test is unaffected. The `opts` slice and its `WithXOp` appends disappear.

## Stage 5 — green both worlds

- `gotest ./...` green (this runs the stdlib suite AND the wasm suite).
- `GOOS=js GOARCH=wasm go build -o /dev/null ./crudview/` succeeds.
- Sprite-leak check stays clean:
  `GOOS=js GOARCH=wasm go list -deps ./crudview | grep tinywasm/svg/sprite` →
  empty.

## Acceptance

- `grep -rn "FakeCaller\|FakeCall\|WithSaveOp\|WithUpdateOp\|WithDeleteOp\|WithArgs\|Reply(" --include="*.go" crudview/` → empty.
- `grep -rn "[^w]view\.New(" --include="*.go" crudview/` → empty (the remaining
  `crudview.New(` hits are the constructor under test, not `view.New`).
- `grep -rn "device_list\|DeviceList" --include="*.go" crudview/` → empty.
- `grep -rn "callsFor" --include="*.go" crudview/` → empty.
- `gotest ./...` green. WASM build clean. Sprite check empty.

## Stages

| # | Scope | Files |
|---|---|---|
| 1 | shared doubles, drop `DeviceList` | `crudview/consumer_test.go` |
| 2 | mechanical translation | the 9 files in the Stage-2 table |
| 3 | wire assertions → typed fields | `crudview/consumer_test.go`, `crudview/bulk_wasm_test.go` |
| 4 | capability-absence doubles | `crudview/bulk_test.go`, `crudview/bulk_wasm_test.go` |
| 5 | green both worlds | full `gotest`, WASM build, sprite check |

---

## NOT in this plan

`crudview` production code (`crudview.go`, `crud.go`, `css.go`, `svg.go` — the
`view.Presenter` contract they consume is unchanged), the pending
`PLAN_ROWCOUNT_OVERLAY.md` queued alongside this one (independent; either order
works, do not mix the two), anything outside this repo.
