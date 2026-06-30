# tinywasm/layout — Plan: `platformd` State as Signals

> **Master:** tinywasm/docs/PLAN.md · **Engine:** tinywasm/dom/docs/PLAN.md
> **Module:** `github.com/tinywasm/layout`
> **Type:** Breaking-aligned migration.

---

## Prerequisites

```bash
# Canonical test runner (WASM tests run against a real DOM). Required: external agents have no global gotest.
go install github.com/tinywasm/devflow/cmd/gotest@latest
```

## Development Rules

- **Documentation First:** update `docs/ARCHITECTURE.md` (platform lifecycle) before code.
- **WASM only:** reactive code in `//go:build wasm`.
- **TinyGo idioms:** `switch` not `map`; embed `Element` by value.
- **Tests:** `gotest` (never `go test`); stdlib only; dual WASM/stdlib. Publish with `gopush 'msg'`.
- **Minimal public API:** export only what a component *user* types; unexport anything only this package uses (helpers, field models, single-use constants). State lives in unexported fields exposed via signals.

## Signals API recap (from the dom engine — self-contained)

```go
open  := dom.NewBool(false); open.Set(true)              // patches bound class/attr surgically
notes := dom.NewNodes()                                  // observable list of rendered rows (*Element)
ul.BindChildren(notes)                                   // notes.Set(buildToasts()) → surgical insert/remove (keyed)
dom.Show(open, renderOverlay)                            // mount/unmount subtree
el.BindClass("nav--open", open)                          // reactive class (SignalBool)
```

State the UI shows lives in a typed `Signal` (**no generics**). No `Update()`. `Init(ctx dom.Ctx)`
runs once; `ctx.OnCleanup(fn)` for teardown.

---

## Context

`platformd` is the canonical proof of the old footgun: its `OnMount`
(platformd/platformd.go:75-106) hand-rolls `if p.mounted { return }`
because `OnMount` fired on every `Update()`. It also has five `Update()` calls (lines 157, 167, 239,
254, 282) — all become signal `Set`s.

---

## Change

1. **State → signals** on the `Platform` struct: `menuOpen *dom.SignalBool`,
   `notifications *dom.SignalNodes` (rendered toast rows), `active *dom.SignalString` (current module
   ID); keep the raw `[]Notification` as a plain field that `Notify`/`dismiss` map into rows. Initialize
   in `Init`. **Delete the `mounted` field** (line 63) — the framework guarantees `Init` runs once.

2. **`OnMount` → `Init(ctx dom.Ctx)`:** keep the hash listener and initial activation; register the
   listener cleanup via `ctx.OnCleanup`:

```go
func (p *Platform) Init(ctx dom.Ctx) {
	p.menuOpen = dom.NewBool(false)
	p.notifications = dom.NewNodes()
	p.active = dom.NewString("")
	dom.OnHashChange(func(hash string) { // dom.OnHashChange — all browser APIs go through tinywasm/dom
		if len(hash) > 0 && hash[0] == '#' { p.Activate(hash[1:]) } // Activate sets p.active → patches UI
	})
	// initial activation (hash or default module) — unchanged logic, sets p.active
}
```

3. **Replace every `Update()` with a signal `Set`:**
   - hamburger/overlay handlers (157, 167): `p.menuOpen.Toggle()` / `p.menuOpen.Set(false)`; bind the
     nav with `BindClass("nav--open", p.menuOpen)` (and/or wrap the mobile overlay in
     `Show(p.menuOpen, …)`). The `Update()` lines are deleted.
   - `Notify` (239): append to the raw slice, then `p.notifications.Set(buildToasts())` (each row an
     `*Element` with `.Key(n.ID)`); the container `ul.BindChildren(p.notifications)` inserts **one** row.
   - `dismiss` (254): remove from the raw slice, `p.notifications.Set(buildToasts())` → removes **one** row.
   - `Activate` (282): `p.active.Set(moduleID)`; the active module section is `Show`/`BindChildren`-driven off
     `p.active`. Keep the `SetHash` side-effect. No `Update()`.

4. The `p.mu` mutex is unrelated and may stay.

---

## Documentation (do FIRST)

- **`docs/ARCHITECTURE.md`**: update the platform lifecycle — `OnMount`+`mounted` guard replaced by
  `Init(ctx)` (once); `menuOpen`/`notifications`/`active` are signals; remove "calls `Update()`" notes.
- If a lifecycle diagram exists, update it (`flowchart TD`, no `subgraph`).
- **`README.md`**: re-index `docs/`.

## Tests — frequent use cases (`gotest`)

Stdlib assertions only; dual WASM/stdlib:

- **stdlib:** activation logic (hash → active module), notification add/remove over signals.
- **wasm (real DOM):**
  - **menu toggle:** flipping `menuOpen` patches only the nav class (assert node identity).
  - **toasts (keyed list):** `Notify`/`dismiss` insert/remove a single row; untouched toasts keep identity.
  - **routing:** `Activate` swaps the active section once; `Init` runs once (no `mounted` guard needed).
- **In-browser (tinywasm MCP):** hamburger/overlay toggle menu; hash navigation activates modules
  once; toasts appear/dismiss; `menuOpen` survives; `browser_get_errors` clean.

## Done When

- `platformd` implements `Render()` + `Init(ctx)` only; `mounted` field/guard deleted; all five
  `Update()` calls replaced by signal `Set`s; toasts are a `SignalNodes` list bound with `BindChildren`.
- **Docs:** ARCHITECTURE.md (+ diagram if present) updated; `README.md` re-indexed. **Tests:** use-case tests pass under `gotest`.
