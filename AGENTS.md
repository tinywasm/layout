# Agent Guide — `tinywasm/layout`

Constraints for agents working on layout (e.g. `platformd`). Read this before any change.

---

## Construction Harness — typed & explicit (the TinyWasm approach)

This library is part of TinyWasm's **construction harness**: the typed, explicit API is what keeps an
agent that doesn't know the library from building wrong code. The compiler must reject mistakes; what
it can't catch becomes a `devMode` warning — never a silent failure.

- **Typed over `any`** — no generic slots; typed builder methods (like `tinywasm/json`), reusing `fmt` types. Anything reactive goes only through a signal binding (`BindText`/`Bind*`), which requires a signal.
- **Explicit names** — `Text` (static) vs `BindText` (reactive); reading the call states intent.
- **Illegal states unrepresentable** — dynamic content has ONE path, typed to require a signal.
- **Minimal public surface** — export only what the author types; engine plumbing stays unexported.
- **Docs are minimal "how" instructions, not long skills** — if a rule must be *remembered*, close it
  with types, not prose.

(Ecosystem rationale: `tinywasm/docs/ARNES_DE_CONSTRUCCION.md`.)

---

## Component Contract — ONE way (signals)

Layout components implement **only** `Render() *dom.Element` (+ optional `Init(ctx dom.Ctx)`).
There is **NO** `OnMount`/`OnUpdate`/`OnUnmount` and **NO** manual `Update()` (unexported in `dom`).

- One-time setup (hash listener, initial route activation) goes in `Init(ctx)`. The framework runs
  it ONCE — do **not** hand-roll a `mounted bool` guard.
- State the UI shows lives in **typed signals**: e.g. `menuOpen *SignalBool`, `active *SignalString`,
  notifications as a `SignalNodes` rendered with `container.BindChildren(s)`.
- Event handlers only mutate signals (e.g. `menuOpen.Toggle()`); the bound DOM patches
  surgically. **Never** re-render the whole platform, never a Virtual DOM.
- For state changed outside a handler (timers, programmatic notifications): `Set` the signal — it
  patches directly, even from a goroutine. Register teardown with `ctx.OnCleanup(fn)`.

## No Generics

Zero generic functions (follow `tinywasm/fmt` codec rule "cero any, cero map"). Concrete typed
signals only: `SignalString`/`SignalBool`/`SignalNodes`, `DeriveString`/`DeriveBool`, `Bind*`,
`Show`. Never `Signal[T]`.

## Minimal Public Surface

Export only the platform's user-facing API (`Platform`, module registration, `Notify`). Unexport
internal helpers and anything only this package uses. Struct fields stay unexported.

## WASM / TinyGo

- Reactive code in `//go:build wasm`; `!wasm` stubs where called from tag-less code.
- No Go stdlib: use `github.com/tinywasm/fmt`. DOM only via `github.com/tinywasm/dom`, never
  `syscall/js`. `switch` not `map`. No `defer/recover`. Embed `dom.Element` by value.

## Testing

```bash
go install github.com/tinywasm/devflow/cmd/gotest@latest
gotest
```

- `gotest`, never `go test`. Stdlib assertions only. Dual WASM/stdlib; real DOM in WASM tests.
- Cover frequent use cases: menu toggle patches only the nav class; toasts (keyed list) insert/remove
  a single row; routing activates once; `Init` runs once (no `mounted` guard). Publish with `gopush 'message'`.

## Documentation First

Update docs **before** code and before `gopush`: `docs/ARCHITECTURE.md` (platform lifecycle:
`Init(ctx)` once, signals, no `Update()`), update any lifecycle diagram (`flowchart TD`, no
`subgraph`, `<br/>` for breaks), and re-index `README.md` so every `docs/` file is linked.
