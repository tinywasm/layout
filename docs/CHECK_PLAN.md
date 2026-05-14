# PLAN — `layout` module: post-execution cleanup

> This plan covers the remaining work after Jules' initial implementation of `layout/platformd`. It is blocked on `tinywasm/css v0.1.0` which introduces (a) `NewStylesheet` rename and (b) 9 new typed property functions. Both consumers in this module (`platformd` and `rightpanel`) must be updated atomically once that version is published.

---

## 1. Blocker — wait for `tinywasm/css v0.1.0`

Do not start any stage below until `github.com/tinywasm/css v0.1.0` is available.

`v0.1.0` delivers:
- **`New` → `NewStylesheet` rename** (breaking change, hence major bump).
- **9 new property functions**: `MarginLeft`, `MarginRight`, `PaddingBottom`, `ListStyle`, `All`, `OverflowY`, `GridTemplateRows`, `GridTemplateColumns`, `BorderRight`.

Once published, bump `layout/go.mod`:

```bash
go get github.com/tinywasm/css@v0.1.0
go mod tidy
```

---

## 2. `platformd/ssr.go` — RawRule sweep + `NewStylesheet`

### 2.1 Rename constructor

```go
// Before
return New(
// After
return NewStylesheet(
```

### 2.2 Replace non-vendor `RawRule` with typed equivalents

Every line below has a `// TODO(css-dsl)` comment in the current file. After bumping to v0.1.0 all of them have a typed replacement.

| Current `RawRule` | Typed replacement |
| --- | --- |
| `RawRule("list-style: none;")` | `ListStyle(None)` |
| `RawRule("margin-left: 5px;")` | `MarginLeft(Px(5))` |
| `RawRule("margin-left: 0;")` ×2 | `MarginLeft(Zero)` |
| `RawRule("margin-left: auto;")` ×2 | `MarginLeft(Auto)` |
| `RawRule("margin-left: -100vw;")` | `MarginLeft(Vw(-100))` |
| `RawRule("margin-left: .4rem;")` | `MarginLeft(Rem(0.4))` |
| `RawRule("margin-right: .4rem;")` | `MarginRight(Rem(0.4))` |
| `RawRule("grid-template-columns: 1fr 3fr 1fr;")` | `GridTemplateColumns(Str("1fr 3fr 1fr"))` |
| `RawRule("all: initial;")` | `All(Initial)` |

### 2.3 Permanent `RawRule` (vendor-prefixed — leave as-is)

These stay. They are vendor-only with no standard typed equivalent planned:

```go
RawRule("-webkit-box-sizing: border-box;")
RawRule("-moz-box-sizing: border-box;")
RawRule("-webkit-user-select: none;")
RawRule("-khtml-user-select: none;")
RawRule("-moz-user-select: none;")
RawRule("-ms-user-select: none;")
```

### 2.4 Verify

```bash
go build ./platformd/...
go test ./platformd/...
grep -n "RawRule" platformd/ssr.go   # only vendor lines remain
```

---

## 3. `rightpanel/ssr.go` — RawRule sweep + `NewStylesheet`

### 3.1 Rename constructor

```go
// Before
return New(
// After
return NewStylesheet(
```

### 3.2 Replace all `RawRule` with typed equivalents

`rightpanel/ssr.go` currently has zero non-vendor `RawRule` with typed equivalents from v0.0.5 — but those were not yet migrated because the sweep was deferred. After v0.1.0:

| Current `RawRule` | Typed replacement |
| --- | --- |
| `RawRule("overflow: hidden;")` ×3 | `Overflow(Hidden)` |
| `RawRule("overflow-y: auto;")` ×2 | `OverflowY(Auto)` |
| `RawRule("overflow-y: visible;")` ×2 | `OverflowY(Visible)` |
| `RawRule("grid-template-rows: auto 1fr;")` ×2 | `GridTemplateRows(Str("auto 1fr"))` |
| `RawRule("border-right: 0.1vw solid "+token.Var()+";")` | `BorderRight(Vw(0.1), Str("solid"), token)` |
| `RawRule("padding-bottom: "+Space1.Var()+";")` | `PaddingBottom(Space1)` |

> `rightpanel` should have **zero** `RawRule` after this sweep — all its usages are standard properties now typed in the DSL.

### 3.3 Verify

```bash
go build ./rightpanel/...
go test ./rightpanel/...
grep -n "RawRule" rightpanel/ssr.go   # must return empty
```

---

## 4. `platformd` — pending items from original CHECK_PLAN.md

The following stages from the original plan were not fully executed by Jules:

### 4.1 `Notify` signature correction

The implemented signature is:
```go
func (p *Platform) Notify(t MessageType, msg string, durationMs int)
```

The original plan specified:
```go
func (p *Platform) Notify(t MessageType, msg string)
```

The `durationMs` parameter makes the auto-dismiss opt-in at every call site, which is ergonomically correct for a reusable shell (some apps want persistent messages). **Keep the current signature** — it is an improvement over the plan. No change needed.

### 4.2 `platformd_test.go` — extend test coverage

Current tests only assert that `RenderHTML()` is non-empty and contains class names. Missing assertions per the original plan:

| Test | What to assert |
| --- | --- |
| `TestPlatform_Render_DefaultModule` | When no hash is set, the module with `Default: true` gets class `pd-panel-active`. |
| `TestPlatform_Activate` | Calling `Activate("mod2")` sets `activeModuleID`, re-renders, `pd-panel-active` moves to mod2's panel. |
| `TestPlatform_Notify_Renders` | Calling `Notify(Msg.Error, "boom", 0)` adds a `pd-msg-error` node to both `pd-msg-desktop` and `pd-msg-mobile` slots. |
| `TestPlatform_Notify_Dismiss` | Calling `Notify(Msg.Info, "hi", 100)` adds the node; after dismiss the notification list is empty. |
| `TestRenderCSS_NonEmpty` | `SSRInstance().RenderCSS().String()` is non-empty and contains `pd-root`. |

### 4.3 `platformd/README.md` — complete the usage section

Current README has a skeleton. It must include:

- Full `web/client.go` usage example (3 modules with `rightpanel`).
- `Notify` usage with `Msg.*` variants and duration.
- CSS token override table (list all `--pd-*` tokens with their defaults).
- Note that `Module.View` accepts any `Component` — not only `rightpanel`.

---

## 5. Final verification (entire module)

After all stages above:

```bash
# Build and test everything
go build ./...
go test ./...

# Zero non-vendor RawRule in layout packages
grep -rn "RawRule" rightpanel/ssr.go platformd/ssr.go \
  | grep -v "webkit\|khtml\|moz-\|ms-user"
# must return empty

# Zero use of old New() constructor
grep -rn "\bNew(" --include="*.go" rightpanel/ platformd/ \
  | grep -v "_test\|//\|js\.Global"
# must return empty

# Zero stdlib strings import
grep -rn "\"strings\"" --include="*.go" .
# must return empty
```

---

## 6. Stages summary

| # | Blocked on | Stage | Verify |
| --- | --- | --- | --- |
| 0 | — | `go get github.com/tinywasm/css@v0.1.0` + `go mod tidy` | `go build ./...` green |
| 1 | Stage 0 | `platformd/ssr.go`: `NewStylesheet` + non-vendor RawRule sweep | `go test ./platformd/...` green; grep clean |
| 2 | Stage 0 | `rightpanel/ssr.go`: `NewStylesheet` + all RawRule sweep | `go test ./rightpanel/...` green; grep empty |
| 3 | Stage 1 | `platformd_test.go`: add 5 missing test cases from §4.2 | `go test ./platformd/...` green with new cases |
| 4 | Stage 3 | `platformd/README.md`: complete usage section per §4.3 | Read-through |
| 5 | Stages 1–4 | Final grep checks from §5 | All three greps return empty |

---

## 7. Acceptance criteria

- `go build ./...` and `go test ./...` green across the entire `layout` module.
- `grep -rn "RawRule" rightpanel/ssr.go` returns **empty**.
- `grep -rn "RawRule" platformd/ssr.go | grep -v "webkit\|khtml\|moz-\|ms-user"` returns **empty**.
- `grep -rn "\bNew(" --include="ssr.go" .` returns **empty** (all callers use `NewStylesheet`).
- `platformd_test.go` has passing tests for Default module, Activate, Notify render, Notify dismiss, and RenderCSS.
- `platformd/README.md` includes token override table and full usage example.
