# PLAN — `layout/platformd` (Platform Dashboard Skeleton)

> Reusable, fully-typed UI shell for tinywasm "platform"-style apps: persistent navigation, sliding module panels, integrated notification system. **Carcasa** only — no business content. The consumer injects modules; the skeleton injects nothing.

Visual/behavioral spec: the CSS embedded in **Appendix A** below is the **source of truth** for this layout. Translate it 1:1 into typed rules using `tinywasm/css`.
Independent sibling: `layout/rightpanel` — used **by composition** inside each module, not by inheritance.

---

## 1. Goals & non-goals

### Goals
- Provide a typed Go API (`tinywasm/dom` + `tinywasm/css`) that renders the platform "carcasa": header + navigation rail/bar + module panel area + message area.
- Single source of truth for the **responsive split** between desktop and mobile (CSS-first, no JS for layout/responsive).
- Built-in **route system** based on URL hash (`#module-id`) — modules slide in from the left when activated.
- Built-in **notification API** (`Notify(level, message)`) that internally renders into the correct viewport slot.
- Be trivially reusable: any project provides a list of typed `Module` entries (id, label, icon, view) and gets a working app shell.

### Non-goals
- Login/auth UI (consumer responsibility — exposed as `UserBlock` slot).
- Module content (consumer responsibility — each `Module` provides its own `Component`).
- Theming primitives beyond exposing CSS tokens for overrides.
- Replicating `rightpanel`'s two-column inner layout — modules that need it use `rightpanel` as their `View`.

---

## 2. Responsive behavior (per user spec)

| Viewport | Navigation rail | Module panel direction |
| --- | --- | --- |
| **Desktop** (landscape + hover) | **Vertical rail fixed to the RIGHT** edge; collapsed (icon only), expands on hover. | Modules slide in **from the left** into the main area. |
| **Mobile** (portrait / no-hover) | **Horizontal bar fixed at the TOP**. | Modules slide in **from the left** as full-screen panels. |

Breakpoint: media query `(orientation: landscape) and (hover: hover)` switches mobile→desktop, matching `menu.css` reference. Mobile is the default (mobile-first CSS).

CSS tokens exposed (with sensible fallbacks):
- `--pd-menu-size`, `--pd-header-height`, `--pd-content-height`
- `--pd-bg`, `--pd-nav-bg`, `--pd-nav-bg-active`
- `--pd-text`, `--pd-text-active`
- `--pd-slide-duration` (default `0.4s`)

---

## 3. Module sliding & routing

### Responsibility split
- **`platformd`** owns: hash routing, active-route state, slide-in/out animation, panel positioning (left→right reveal).
- **`rightpanel`** owns: nothing about routing. It's a content layout (main + aside) used **inside** a module if desired. The user's question "should rightpanel be responsible for the slide?" → **No.** Composition: `Module.View = &rightpanel.RightPanel{...}`.

### Mechanism
- Each registered module renders as a `<section class="pd-panel">` positioned `translateX(-100%)` (off-screen left).
- The currently active module gets class `pd-panel-active` → `translateX(0)`.
- Active state is driven by `window.location.hash` (`#module-id`); a tiny hash listener in `platformd.go` (wasm build) toggles the class. No JS framework; just `dom.AddEventListener("hashchange", ...)`.
- SSR / `!wasm` builds: render with the first module active (or one declared `Default: true`) so the static HTML is meaningful.

---

## 4. Notification API (typed)

**Do not declare a local `Level` type.** Reuse `fmt.MessageType` from `github.com/tinywasm/fmt` (file `messagetype.go`) which already defines the canonical message classification via the `fmt.Msg` struct (`Normal`, `Info`, `Error`, `Warning`, `Success`, `Debug`, etc.).

```go
import . "github.com/tinywasm/fmt"

// On the Platform instance:
func (p *Platform) Notify(t MessageType, msg string)

// Usage by the consumer:
p.Notify(Msg.Success, "Plataforma cargada")
p.Notify(Msg.Error,   "Conexión perdida")
p.Notify(Msg.Warning, "Modo limitado")
p.Notify(Msg.Info,    "Sincronizando…")
```

CSS class derivation: use `MessageType.String()` (already implemented) lowercased to build `pd-msg-{string}` — e.g. `Msg.Success` → `pd-msg-success`, `Msg.Error` → `pd-msg-error`. Only these five styled variants are required for v1: `Normal`, `Info`, `Success`, `Warning`, `Error`. Other `MessageType` values fall back to `pd-msg-normal`.

Internal:
- Two slots rendered in the DOM: `<div class="pd-msg-desktop">` (inside header) and `<div class="pd-msg-mobile">` (below top bar on mobile).
- CSS hides the slot not matching the current viewport — Notify writes the message into BOTH; only the visible one is shown. Avoids needing JS to detect viewport.
- Each notification is a `<div class="pd-msg pd-msg-{type}">` with auto-dismiss timer (configurable token).

---

## 5. Public API

```go
package platformd

import (
    . "github.com/tinywasm/dom"
)

// Module describes one registered route/page in the shell.
// Pure data: the consumer creates these and passes them to Platform.
type Module struct {
    ID      string     // hash slug, e.g. "products" → routed via "#products"
    Label   string     // text shown next to icon in expanded rail
    Icon    Component  // icon component (usually a tinywasm/icons Symbol)
    View    Component  // the module's content (often a *rightpanel.RightPanel)
    Default bool       // if true and no hash set, this module is shown initially
}

// Platform is the typed skeleton root.
type Platform struct {
    *Element

    // AppName appears in the header (left side, near UserBlock).
    AppName string

    // UserBlock slot — usually an avatar/name/logout link. Optional.
    UserBlock Component

    // Modules registered in order — appearance order in the nav rail.
    Modules []Module

    // internal: notification queue, active module
    // (unexported state)
}

// Render builds the DOM tree (implements ViewRenderer).
func (p *Platform) Render() *Element

// Notify queues a typed notification in the proper viewport slot.
func (p *Platform) Notify(level Level, msg string)

// Activate programmatically switches to a module by ID
// (also updates window.location.hash on wasm builds).
func (p *Platform) Activate(moduleID string)
```

### DOM structure rendered

```
<div class="pd-root">
  <header class="pd-header">
    <div class="pd-user-block">{UserBlock}</div>
    <div class="pd-msg-desktop"></div>
    <h2 class="pd-area">{active module Label}</h2>
  </header>

  <div class="pd-msg-mobile"></div>

  <nav class="pd-menu">
    <ul class="pd-navbar">
      <li class="pd-nav-item">
        <a href="#{ID}" class="pd-nav-link" data-id="{ID}">
          {Icon}
          <span class="pd-link-text">{Label}</span>
        </a>
      </li>
      ... per module
    </ul>
  </nav>

  <main class="pd-stage">
    <section class="pd-panel" id="{ID}" data-id="{ID}">
      {View}
    </section>
    ... per module
  </main>
</div>
```

---

## 6. File layout

```
layout/platformd/
├── platformd.go        # Platform type, Module type, Render, Notify, Activate
├── platformd_test.go   # render + routing + notification tests
├── ssr.go              # //go:build !wasm — SSRInstance(), RenderCSS()
├── tokens.go           # //go:build !wasm — CSS token declarations
├── docs/
│   ├── PLAN.md         # this file
│   └── DESIGN.md       # (later) detailed rationale + diagrams
└── README.md
```

All CSS is typed via `github.com/tinywasm/css` — **no raw `.css` files, no CSS-as-strings inside Go**. The CSS in Appendix A is the **visual spec** (reference only); the actual implementation uses `Rule`, `Media`, `Declare`, `Token`, `Class` and the typed property/value DSL exclusively. See the "Translation example" in Appendix A for the mandatory pattern.

---

## 7. CSS translation map

| Source CSS (Appendix A) | Becomes in `ssr.go` |
| --- | --- |
| `body.css` block | Root `pd-root` grid: `grid-template-areas` switching between mobile/desktop. |
| `menu.css` block | `pd-menu` + `pd-navbar` + `pd-nav-link` + hover-expand + `.hash-selected` → `pd-nav-active`. |
| `slider-panel.css` block | `pd-panel` positioning + horizontal translate + `pd-panel-active` transition. |
| `user-message.css` block | `pd-msg-desktop` / `pd-msg-mobile` + `pd-msg-{level}` color variants. |
| `add-default.css` (tokens) | Root CSS custom properties + universal reset rule. |

Class renames: original CSS uses ad-hoc names (`.menu-container`, `.slider_panel`, `#USER_AREA`, `.hash-selected`, etc.). In the typed port everything is prefixed `pd-*` (see DOM tree in §5). Map:

| Original | Typed name |
| --- | --- |
| `.menu-container` | `pd-menu` |
| `.navbar-container` | `pd-navbar` |
| `.navbar-item` | `pd-nav-item` |
| `.navbar-link` | `pd-nav-link` |
| `.link-text` | `pd-link-text` |
| `.fa-primary` | `pd-nav-icon` |
| `.hash-selected` | `pd-nav-active` |
| `.slider_panel` | `pd-panel` |
| `.slider_panel:target` | `.pd-panel.pd-panel-active` (class toggled by JS, not `:target`) |
| `#USER_NAME` | `pd-user-block` |
| `#user-desktop-messages` | `pd-msg-desktop` |
| `#user-mobile-messages` | `pd-msg-mobile` |
| `#USER_AREA` | `pd-area` |
| `.off` / `.err` / `.ok` | `pd-msg-info` / `pd-msg-error` / `pd-msg-success` (names derived from `fmt.MessageType.String()` lowercased) |
| `#horizontal-mobile-message` | `pd-orientation-warn` |

Per project rules: **no stdlib `strings`** — use `github.com/tinywasm/fmt` (Builder/Join) for any string composition. Tokens declared as `Token{Name, Fallback}` and referenced via `.Var()`. Color/space tokens that already exist in `tinywasm/css` (e.g. `ColorPrimary`, `Space2`) must be reused; only declare new tokens for layout-specific values (`--pd-menu-size`, `--pd-header-height`, `--pd-slide-duration`, etc.).

---

## 8. Example consumer — `web/client.go`

A minimal demo registering 3 modules, each using `rightpanel` as its content, to show composition:

```go
//go:build wasm

package main

import (
    . "github.com/tinywasm/dom"
    . "github.com/tinywasm/fmt"
    "github.com/tinywasm/layout/platformd"
    "github.com/tinywasm/layout/rightpanel"
)

// Tiny model stub so rightpanel has an ID source.
type mod struct{ name string }
func (m mod) ModelName() string { return m.name }

func main() {
    p := &platformd.Platform{
        AppName: "Demo Platform",
        Modules: []platformd.Module{
            {
                ID: "mod1", Label: "Módulo 1", Default: true,
                Icon: iconHome(), // consumer-provided icon component
                View: &rightpanel.RightPanel{
                    Module: mod{"mod1"}, Title: "Módulo 1",
                },
            },
            {
                ID: "mod2", Label: "Módulo 2",
                Icon: iconProducts(),
                View: &rightpanel.RightPanel{
                    Module: mod{"mod2"}, Title: "Módulo 2",
                },
            },
            {
                ID: "mod3", Label: "Módulo 3",
                Icon: iconInfo(),
                View: &rightpanel.RightPanel{
                    Module: mod{"mod3"}, Title: "Módulo 3",
                },
            },
        },
    }
    Body().Add(p)
    Render()

    // demo of typed Notify — uses fmt.MessageType / fmt.Msg
    p.Notify(Msg.Success, "Plataforma cargada")
}
```

This file goes in the new directory `layout/platformd/web/client.go` (created in stage 5).

---

## 9. Stages

| # | Stage | Output | Verify |
| --- | --- | --- | --- |
| 1 | **Public types** | `platformd.go`: `Module`, `Platform`, `Level`, method signatures (empty bodies). | `go build ./...` compiles. |
| 2 | **DOM rendering** | Implement `Render()`: header, nav, stage with panels, message slots. No animation/routing yet. | Unit test snapshots HTML tree for a 3-module fixture. |
| 3 | **CSS (SSR)** | `ssr.go` + `tokens.go`: translate the 4 reference CSS files into typed `Stylesheet`, including mobile-first + landscape/hover media query. | `RenderCSS()` returns non-empty; manual visual check via SSR snapshot. |
| 4 | **Routing + slide** | Hash listener on wasm; `Activate()` toggles `pd-panel-active` and `pd-nav-active`. SSR picks `Default` (or first) module. | Test: simulate hashchange events and assert active classes. |
| 5 | **Notify** | Implement `Notify()` queue, auto-dismiss timer (CSS-driven via `animation`). | Test: call `Notify`, assert message node added/removed. |
| 6 | **Example** | Create `web/client.go` with 3 modules using `rightpanel`. Build with tinywasm/app. | Open in browser: rail on right (desktop), top bar (mobile), panels slide, Notify visible. |
| 7 | **Docs** | `README.md` with usage snippet; `docs/DESIGN.md` if non-obvious choices need recording. | Read-through. |

Each stage ends with `go test ./layout/platformd/...` green before moving on.

---

## 10. Open questions (resolved here for record)

- **Q: Should the slide be `rightpanel`'s responsibility?**
  A: No. `rightpanel` is a content layout (main + aside). The slide is part of `platformd`'s routing. Modules that want the two-column look set `Module.View = &rightpanel.RightPanel{...}` — composition, not inheritance.
- **Q: Login form?** Out of scope. Consumer can render it inside a module or via `UserBlock`.
- **Q: Icons?** Consumer-provided `Component`. `platformd` doesn't register icons itself; this keeps the package free of icon dependencies.
- **Q: Active route source of truth?** URL hash (`window.location.hash`). `Activate()` updates it; CSS reacts via the class toggle, not via `:target` (we need programmatic control for the initial Default and for SSR).

---

## 11. Acceptance criteria

- `go build ./...` and `go test ./...` pass for `layout/platformd`.
- Example at `layout/platformd/web/client.go` builds to wasm and runs.
- Visually: on a desktop browser the rail sits on the right edge and expands on hover; on a narrow mobile viewport the rail becomes a top bar.
- Switching modules via clicking nav items slides the panel in from the left in ≤ `--pd-slide-duration`.
- `Notify(LevelError, "boom")` shows a styled error toast in the correct viewport slot and auto-dismisses.
- Zero references to stdlib `strings`; all CSS typed via `github.com/tinywasm/css` — no CSS-as-strings, no `.css` files. `grep -nE "\\.[a-z][a-z0-9_-]+\\s*\\{" *.go` must return empty.

---

## Appendix A — Source CSS (visual spec ONLY — do NOT copy as strings)

> ⚠️ **Critical implementation rule**
>
> The CSS blocks below are the **visual/behavioral specification**, written in plain CSS only so a human can read them. They are **NOT** to be pasted into Go as strings.
>
> The implementation in `ssr.go` MUST be **100% typed via `github.com/tinywasm/css`** — using `New()`, `Root()`, `Rule()`, `Media()`, `Keyframes()`, `Declare()`, `Token{}`, `Class`, typed property functions (`Display`, `Width`, `Background`, `Transition`, …) and typed values (`Px`, `Em`, `Pct`, `Str`, `Hex`, …).
>
> **Forbidden**:
> - `RawRule("…")` for whole declarations that have a typed equivalent. (Use `RawRule` only as a last-resort escape hatch for properties not yet covered by the DSL — e.g. `grid-template-areas` multi-line strings, vendor-prefixed properties — and document why.)
> - Concatenating CSS as Go strings.
> - `fmt.Sprintf` to build selectors or values when a typed primitive exists.
> - Any literal `.foo { … }` blocks inside Go.
>
> Reference implementation pattern to follow exactly: `layout/rightpanel/ssr.go` (already in the repo) — same package layout, same `SSRInstance()` + `RenderCSS()` shape, same use of `Token` + `Class` + `Rule` + `Media`.

### Translation example (mandatory pattern)

The first block of `menu.css`:

```css
.menu-container {
    background-color: var(--color-gray);
    position: fixed;
    top: 0;
    width: 100vw;
    height: var(--menu-size);
    transition: 0.6s all ease;
    z-index: 3;
}
```

MUST be implemented as:

```go
//go:build !wasm
package platformd

import . "github.com/tinywasm/css"

var (
    clsMenu Class = "pd-menu"

    tokenMenuSize     = Token{Name: "--pd-menu-size",     Fallback: "6vh"}
    tokenSlideDur     = Token{Name: "--pd-slide-duration", Fallback: "0.6s"}
    tokenNavBg        = Token{Name: "--pd-nav-bg",        Fallback: "var(--color-gray)"}
)

func (p *Platform) RenderCSS() *Stylesheet {
    return New(
        Root(
            Declare(tokenMenuSize, "6vh"),
            Declare(tokenSlideDur, "0.6s"),
            Declare(tokenNavBg,    ColorGray.Var()), // reuse existing tinywasm/css token
        ),

        Rule(clsMenu,
            Background(tokenNavBg),
            Position(Fixed),
            Top(Zero),
            Width(Vw(100)),
            Height(tokenMenuSize),
            Transition(tokenSlideDur, Str("all"), Str("ease")),
            ZIndex(Str("3")),
        ),

        Media("(orientation: landscape) and (hover: hover)",
            Rule(clsMenu,
                Top(tokenHeaderHeight),
                Width(tokenMenuSize),
                Height(tokenContentHeight),
                Right(Zero),
                // …
            ),
            // …
        ),
    )
}
```

Every other CSS block in this appendix MUST be translated using this same approach. If a property/value is missing from the `tinywasm/css` DSL, the agent's job is to **add it to the DSL** (typed) — not to fall back to raw strings.

These five blocks are the original CSS files from the reference platform. **Port them with full type safety** (same selectors, properties, breakpoints, transitions) using the class renames in §7.

### A.1 `add-default.css` — tokens + universal reset

```css
:root {
    /* Font Sizes */
    --font-size-normal: 1.1rem;
    --font-size-small: .6rem;

    /* Colors */
    --color-primary: #ffffff;
    --color-secondary: #3f88bf;
    --color-tertiary: #c2c1c1;
    --color-quaternary: #000000;
    --color-gray: #e9e9e9;
    --color-selection: #ff9300;
    --color-hover: #ff95008e;
    --color-success: #aadaff7c;
    --color-error: #f20707;

    /* Layout Sizes */
    --menu-size: 6vh;
    --content-height: 94vh;
    --content-width: 100vw;

    /* Timing */
    --transition-wait: 0s;
}

* {
    -webkit-box-sizing: border-box;
    -moz-box-sizing: border-box;
    margin: 0;
    padding: 0;
    box-sizing: border-box;
    outline: none;
    text-decoration: none;
    list-style: none;
    list-style-type: none;
    -webkit-user-select: none;
    -khtml-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
    user-select: none;
    font-family: Arvo;
}
```

> Note: existing tokens in `tinywasm/css` (e.g. `ColorPrimary`, `ColorSecondary`) should be reused when they exist. Layout-specific tokens (`--menu-size`, `--content-height`, `--content-width`, `--header-height`, `--transition-wait`) become `--pd-*` Tokens declared in `tokens.go`.

### A.2 `body.css` — root grid (mobile + desktop)

```css
/************ MOBILE ************/
body {
    font-size: .8rem;
    background-color: var(--color-tertiary);
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100vh;
}

body::before {
    background-size: 10em;
    background-position: center center;
    content: "";
    position: absolute;
    top: auto;
    left: 50%;
    transform: translateX(-50%);
    right: 50%;
    width: 10em;
    height: 10em;
    background-image: url("logo.png");
    background-repeat: no-repeat;
    filter: grayscale(100%) opacity(30%);
}

header {
    position: fixed;
    align-self: flex-end;
    bottom: 5px;
    margin-left: 5px;
}

a:link,
a:visited {
    text-decoration: none;
    color: var(--color-quaternary);
}

/************ DESKTOP ************/
@media only screen and (orientation: landscape) and (hover: hover) {
    :root {
        --header-height: 5vh;
        --content-height: 95vh;
        --menu-size: 5vw;
        --content-width: 95vw;
    }

    body {
        overflow: hidden;
        background-color: var(--color-gray);
        font-size: 16px;
        height: var(--content-height);
        display: grid;
        grid-template:
            "header         header" var(--header-height)
            "module-content menu-container" var(--content-height) /
            var(--content-width) var(--menu-size);
    }

    header,
    .menu-container {
        background-color: var(--color-gray);
    }

    header {
        align-self: unset;
        bottom: unset;
        position: unset;
        margin-left: 0;
        max-height: var(--header-height);
        width: 100vw;
        grid-area: header;
        display: grid;
        grid-template-columns: 1fr 3fr 1fr;
    }
}
```

> Port note: the `body::before` logo background and the `url("logo.png")` are **not** part of this layout — drop them. The layout renders inside `<div class="pd-root">`, not `<body>`; replicate the grid on `pd-root` so the package doesn't fight with the consumer's `<body>` styling.

### A.3 `menu.css` — navigation rail / top bar

```css
/************ MOBILE ************/
.menu-container {
    background-color: var(--color-gray);
    position: fixed;
    top: 0;
    width: 100vw;
    height: var(--menu-size);
    text-decoration: none;
    transition: 0.6s all ease;
    z-index: 3;
}

.navbar-container {
    list-style: none;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: space-around;
    height: 100%;
}

.navbar-item {
    width: 100%;
}

.hash-selected {
    background: var(--color-selection);
}

.hash-selected,
.hash-selected > svg {
    color: var(--color-primary);
    transition: 0.3s all ease;
}

.navbar-link {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: var(--menu-size);
    text-decoration: none;
    transition: 0.6s all ease;
}

.fa-primary {
    color: var(--color-secondary);
    height: 8em;
    width: 8em;
    transition: 0.6s all ease;
}

.link-text {
    display: none;
}

/************ DESKTOP ************/
@media only screen and (orientation: landscape) and (hover: hover) {
    .menu-container {
        top: var(--header-height);
        width: var(--menu-size);
        height: var(--content-height);
        margin-left: auto;
        right: 0;
    }

    .menu-container:hover {
        width: 10em;
    }

    .menu-container:hover .link-text {
        display: block;
    }

    .navbar-container {
        flex-direction: column;
    }

    .navbar-item {
        height: 100%;
    }

    .navbar-link {
        flex-direction: row;
        height: 100%;
    }

    .navbar-link svg {
        margin: 0 .5em;
    }

    .navbar-link:hover {
        background-color: var(--color-hover);
    }

    .hash-selected,
    .hash-selected > .link-text {
        color: var(--color-primary);
    }

    .link-text {
        color: var(--color-secondary);
    }

    .fa-primary {
        max-width: 2.5em;
        min-width: 2.5em;
    }

    .btn-url-disable {
        pointer-events: none;
        background: var(--color-tertiary);
    }

    .btn-selected {
        background: var(--color-selection);
    }
}
```

> Port note: per user spec the desktop rail is anchored to the **right** edge (`margin-left: auto; right: 0;`) — keep this. The mobile bar stays at the top (`top: 0`).

### A.4 `slider-panel.css` — module panel slide-in

```css
/************ MOBILE ************/
.slider_panel {
    top: -100vh;
    margin-left: 0;
    z-index: 1;
    width: var(--content-width);
    height: var(--content-height);
    position: absolute;
    background-color: #000;
    -webkit-transition: all .6s ease-in-out;
    -moz-transition: all .6s ease-in-out;
    -o-transition: all .6s ease-in-out;
    -ms-transition: all .6s ease-in-out;
    transition: all .6s ease-in-out;
}

.slider_panel:target {
    top: var(--menu-size);
    background-color: var(--color-secondary);
}

/************ DESKTOP ************/
@media only screen and (orientation: landscape) and (hover: hover) {
    .slider_panel {
        grid-area: module-content;
        position: unset;
        top: var(--header-height);
        margin-left: -100vw;
        border-radius: 0;
    }

    .slider_panel:target {
        margin-left: 0;
    }
}
```

> Port note: **do not use `:target`** in the typed port. Use a class `pd-panel-active` toggled by the `hashchange` listener in `platformd.go` (wasm build) so initial state, SSR rendering, and `Activate()` programmatic switches all work uniformly. The slide direction stays: panel enters from the left (`margin-left: -100vw` → `0` on desktop; `top: -100vh` → `top: var(--menu-size)` on mobile, which is a top-down slide — user confirmed left-to-right; reconcile by using `margin-left: -100vw → 0` on mobile too, anchored below the top bar via `top: var(--menu-size)`).

### A.5 `user-message.css` — notification slots

```css
.permanent {
    color: var(--color-quaternary);
    font-size: 1.3em;
    text-shadow: 0.1em 0.1em 0.1em #ffffff;
}

.foka,
.ferr {
    border-width: 3px;
    border-style: solid;
}

.foka { border-color: var(--color-secondary); }
.ferr { border-color: var(--color-error); }

#USER_NAME { text-decoration: none; }
header #USER_NAME a { text-transform: capitalize; }

#user-desktop-messages {
    width: 100%;
    justify-content: center;
}

#USER_AREA {
    text-transform: uppercase;
    color: var(--color-primary);
    text-align: right;
}

#USER_NAME,
#user-desktop-messages,
#USER_AREA {
    display: flex;
    align-items: center;
}

.err { color: var(--color-error); }
.ok  { color: var(--color-secondary); }
.off { color: var(--color-tertiary); }

@keyframes fadeOut {
    from { opacity: 1; visibility: visible; }
    to   { opacity: .7; visibility: hidden; }
}

@keyframes fadeIn {
    from { opacity: 0; visibility: hidden; }
    to   { opacity: 1; visibility: visible; }
}

#horizontal-mobile-message,
#user-mobile-messages {
    z-index: 4;
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    justify-content: center;
    align-items: center;
    display: flex;
}

#user-mobile-messages {
    background: rgba(255, 255, 255, 0.8);
    visibility: hidden;
    animation: fadeOut var(--transition-wait);
}

#horizontal-mobile-message {
    color: var(--color-primary);
    touch-action: none;
    display: none;
    background: var(--color-secondary);
}

@media only screen and (orientation: portrait) {
    #horizontal-mobile-message { display: none; }
}

@media only screen and (orientation: landscape) and (min-width: 600px) and (max-width: 1024px) {
    #horizontal-mobile-message { display: flex; }
}

.off,
.err,
.ok {
    text-align: center;
    padding: 20px;
    width: 80%;
    font-size: 1.3em;
    text-shadow: 0.1em 0.1em 0.1em #ffffff;
}

/************ DESKTOP ************/
@media only screen and (orientation: landscape) and (hover: hover) {
    .off,
    .err,
    .ok {
        padding: 0;
        animation: fadeOut var(--transition-wait);
        visibility: hidden;
        width: 100%;
    }

    #USER_NAME {
        font-size: calc(.5em + .5vh);
        margin-left: .4rem;
    }

    #USER_AREA {
        font-size: calc(.5em + .5vh);
        margin-left: auto;
        margin-right: .4rem;
    }

    #user-desktop-messages {
        font-size: calc(.4em + .4vh);
    }

    #user-mobile-messages {
        all: initial;
    }
}
```

> Port note: drop the `.foka`/`.ferr`/`.permanent` legacy classes (not used in this layout). Keep `pd-msg-info` (`.off`), `pd-msg-success` (`.ok`), `pd-msg-error` (`.err`), plus an additional `pd-msg-warning` variant tied to `--color-selection`. The `#horizontal-mobile-message` becomes `pd-orientation-warn` and is part of the skeleton (shown automatically in incompatible landscape mobile range).
