# Agent Guide — `tinywasm/layout`

Constraints for agents working on layout (e.g. `platformd`). Read this before any change.

---

## Construction Harness — the TinyWasm way (read first)

The typed, explicit code **is** the harness. Whoever writes against this library is
often an agent that does **not** know it; they must produce correct code guided only
by the signatures, and the compiler must **reject** whatever is wrong. Correctness
lives in the compiler and the signatures, not in a manual you must remember. A
harness moves correctness to the compiler; a manual moves it to the reader — the
first is orders of magnitude more reliable for someone with no context.

Every public API must hold to these principles:

1. **Typed over `any`.** No generic holes (`func(...any)`, `interface{}`) in the
   API — intent-typed methods, like the `tinywasm/json` writer (`String`, `Int`,
   `Bool`, `Object`, `Array`). `any` is allowed **only** at the I/O edge, never in
   the data. **Reuse already-declared types** (e.g. `fmt.KeyValue`) instead of
   inventing new ones. Generics with an `any` constraint are the same hole in
   disguise: a signature that does not name its real type is not self-describing.
2. **Explicit over implicit.** The name declares the intent; reading the call is
   enough to know what it does, without opening the implementation.
3. **Illegal states unrepresentable.** If something must not happen, it must not be
   writable. One intent = one path, typed to demand exactly what it needs.
4. **One way to do each thing.** A single construction pattern, with no alternatives
   that force a choice or a trip to the docs.
5. **Minimal public surface.** Export exactly what the author uses; internal
   machinery stays unexported — you cannot misuse what you cannot see.
6. **Fail at compile time, not at runtime.** Order of preference to catch an error:
   compile error → noisy dev-mode diagnostic → (never) silent failure.
7. **Self-describing signatures.** Autocomplete must be enough to build. If using
   the API needs a long document, the API is incomplete.

**Litmus test:** if an agent with no context produces correct code guided only by
autocomplete and a few-line example, the harness is closed. If it needs a manual to
avoid mistakes, something is still untyped.

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
