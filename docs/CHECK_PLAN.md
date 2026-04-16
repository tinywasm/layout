# PLAN: tinywasm/layout — RightPanel layout

## Project purpose

`tinywasm/layout` provides pre-built, consistent, reusable UI layout skeletons for
tinywasm modules. A layout defines WHERE elements are placed (structure, grid, spacing)
but NOT what those elements contain — that is the consumer's responsibility.

Layouts are built exclusively with `tinywasm/dom` primitives (`*dom.Element`, factory
functions like `dom.Div()`, `dom.Section()`, etc.). No HTML templates, no text/template,
no embed of raw HTML.

Each layout lives in its own sub-package so consumers import only what they use:

```
github.com/tinywasm/layout/rightpanel   ← this plan
github.com/tinywasm/layout/fullpage     ← future
github.com/tinywasm/layout/storefront   ← future (e-commerce)
```

**Prerequisite:** Execute `tinywasm/dom` PLAN.md first. This plan depends on
`github.com/tinywasm/dom` having `CssVars`, `ThemeCSS`, and `theme.css` available.

---

## Layout: RightPanel

### Visual structure

```
┌─────────────────────────────────────────────────────────┐
│  div#<ModelName()>  (.rp-wrapper)                        │
│  ┌────────────────────────────────┐  ┌────────────────┐ │
│  │ section.rp-main                │  │ aside.rp-aside │ │
│  │ ┌────────────────────────────┐ │  │ ┌────────────┐ │ │
│  │ │ div.rp-header              │ │  │ │rp-aside-hdr│ │ │
│  │ │  <h1>Title</h1>  [Head]    │ │  │ │[AsideCtrls]│ │ │
│  │ │  [HeadControls]            │ │  │ └────────────┘ │ │
│  │ └────────────────────────────┘ │  │ ┌────────────┐ │ │
│  │ ┌────────────────────────────┐ │  │ │rp-aside-   │ │ │
│  │ │ article.rp-article         │ │  │ │content     │ │ │
│  │ │  [Article]                 │ │  │ │[Aside]     │ │ │
│  │ └────────────────────────────┘ │  │ └────────────┘ │ │
│  └────────────────────────────────┘  └────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

Desktop: rp-main (~66vw) + rp-aside (~30vw) side by side.
Mobile (<640px): rp-main full width, rp-aside stacked below or hidden (CSS-only).

### Slots (all optional — nil = not rendered)

| Field         | Type          | Position                        | Typical use             |
|---------------|---------------|---------------------------------|-------------------------|
| Title         | string        | `<h1>` in rp-header             | Module name             |
| Head          | dom.Component | Beside `<h1>` in rp-header      | Status badge, icon      |
| HeadControls  | dom.Component | Below title row in rp-header    | Select with search      |
| Article       | dom.Component | rp-article (main content area)  | Table, form, list       |
| AsideControls | dom.Component | Top of rp-aside                 | Search input + filters  |
| Aside         | dom.Component | Content area of rp-aside        | Detail panel, info card |

### Module identification

`RightPanel` does NOT generate its own ID. The consumer passes a value that satisfies:

```go
type Module interface {
    ModelName() string
}
```

The return value of `ModelName()` becomes the `id` attribute of the root `div.rp-wrapper`.
This avoids reflect and integrates naturally with the orm model pattern already used in
the ecosystem.

---

## Files to create

### Directory layout

```
layout/
├── go.mod
├── go.sum
├── layout.go          ← package doc only (already created by gonew)
├── LICENSE
├── README.md
├── docs/
│   └── PLAN.md        ← this file
└── rightpanel/
    ├── rightpanel.go
    ├── rightpanel.css
    ├── ssr.go
    └── rightpanel_test.go
```

---

### `go.mod` — update after creating files

Add dependency on `tinywasm/dom`:

```
require github.com/tinywasm/dom v<latest>
```

Run `go get github.com/tinywasm/dom@latest` to resolve.

---

### `rightpanel/rightpanel.go`

```go
package rightpanel

import "github.com/tinywasm/dom"

// Module is the interface the consumer must satisfy to provide the layout ID.
// Any struct with a ModelName() string method qualifies (e.g. ORM model structs).
type Module interface {
    ModelName() string
}

// RightPanel is a two-column layout skeleton:
//   - Left: main content area with header (title + controls) and article.
//   - Right: aside panel with its own header (controls) and content.
//
// All slots are optional. A nil slot is simply not rendered.
// The layout does not define what the slots contain — that is the consumer's job.
//
// Usage:
//
//	panel := &rightpanel.RightPanel{
//	    Module:        myModel,          // implements ModelName() string
//	    Title:         "Users",
//	    HeadControls:  mySelectSearch,
//	    Article:       myTable,
//	    AsideControls: myFilterBar,
//	    Aside:         myDetailPanel,
//	}
//	panel.Render("app")
type RightPanel struct {
    *dom.Element

    // Module provides the ID for the root wrapper element.
    Module Module

    // Title is rendered as <h1> in the header.
    Title string

    // Head is rendered beside the <h1> (e.g. status badge, icon).
    Head dom.Component

    // HeadControls is rendered below the title row (e.g. select with search).
    HeadControls dom.Component

    // Article is the main content area.
    Article dom.Component

    // AsideControls is rendered at the top of the aside panel (e.g. search + filter).
    AsideControls dom.Component

    // Aside is the content area of the aside panel (e.g. detail view, info card).
    Aside dom.Component
}

// Render builds the layout element tree.
// Implements dom.ViewRenderer.
func (r *RightPanel) Render() *dom.Element {
    if r.Element == nil {
        r.Element = &dom.Element{}
    }

    // ── root wrapper ─────────────────────────────────────────────────────────
    id := ""
    if r.Module != nil {
        id = r.Module.ModelName()
    }

    wrapper := dom.Div().Class("rp-wrapper")
    if id != "" {
        wrapper.ID(id)
    }

    // ── main section ─────────────────────────────────────────────────────────
    main := dom.Section().Class("rp-main")

    // header row: title + Head slot + HeadControls slot
    header := dom.Div().Class("rp-header")

    titleRow := dom.Div().Class("rp-title-row")
    if r.Title != "" {
        titleRow.Add(dom.H1().Text(r.Title))
    }
    if r.Head != nil {
        titleRow.Add(r.Head)
    }
    header.Add(titleRow)

    if r.HeadControls != nil {
        header.Add(dom.Div().Class("rp-head-controls").Add(r.HeadControls))
    }
    main.Add(header)

    // article
    if r.Article != nil {
        main.Add(dom.Article().Class("rp-article").Add(r.Article))
    } else {
        main.Add(dom.Article().Class("rp-article"))
    }

    wrapper.Add(main)

    // ── aside panel ──────────────────────────────────────────────────────────
    if r.AsideControls != nil || r.Aside != nil {
        aside := dom.Aside().Class("rp-aside")

        if r.AsideControls != nil {
            aside.Add(dom.Div().Class("rp-aside-header").Add(r.AsideControls))
        }
        if r.Aside != nil {
            aside.Add(dom.Div().Class("rp-aside-content").Add(r.Aside))
        }

        wrapper.Add(aside)
    }

    return wrapper
}
```

---

### `rightpanel/rightpanel.css`

The CSS uses `--color-*` and `--mag-*` tokens defined by `tinywasm/dom`'s `theme.css`.
Layout-specific dimensions use `--rp-*` variables with fallbacks so the layout works
even if the consumer has not injected the dom theme.

```css
/* ── RightPanel layout tokens (with fallbacks) ───────────────── */
:root {
  --rp-title-height:    var(--title-height,    8vh);
  --rp-content-height:  var(--content-height,  89vh);
  --rp-controls-height: var(--controls-height, 3vh);
  --rp-main-width:      66vw;
  --rp-aside-width:     30vw;
  --rp-gap:             var(--mag-pri, 0.5rem);
  --rp-border-color:    var(--color-tertiary,  #94a3b8);
  --rp-bg:              var(--color-gray,      #f8fafc);
  --rp-aside-bg:        var(--color-quaternary,#1e293b);
  --rp-title-color:     var(--color-secondary, #7c3aed);
}

/* ── Wrapper ─────────────────────────────────────────────────── */
.rp-wrapper {
  display: flex;
  flex-direction: row;
  width: 100%;
  height: var(--rp-content-height);
  overflow: hidden;
}

/* ── Main section ────────────────────────────────────────────── */
.rp-main {
  display: grid;
  grid-template-rows:
    auto          /* rp-header  */
    1fr;          /* rp-article */
  width: var(--rp-main-width);
  height: 100%;
  overflow: hidden;
  border-right: 0.1vw solid var(--rp-border-color);
}

/* ── Header ──────────────────────────────────────────────────── */
.rp-header {
  display: flex;
  flex-direction: column;
  background: var(--rp-bg);
  padding: var(--mag-sec, 0.2rem) var(--mag-pri, 0.5rem);
}

.rp-title-row {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: var(--rp-gap);
  min-height: var(--rp-title-height);
}

.rp-title-row h1 {
  font-size: 1.5rem;
  color: var(--rp-title-color);
  margin: 0;
}

.rp-head-controls {
  display: flex;
  flex-direction: row;
  align-items: center;
  min-height: var(--rp-controls-height);
  padding-bottom: var(--mag-sec, 0.2rem);
}

/* ── Article ─────────────────────────────────────────────────── */
.rp-article {
  overflow-y: auto;
  padding: var(--mag-pri, 0.5rem);
  background: var(--color-gray, #f8fafc);
  border-radius: 0.4em 0.4em 0 0;
}

.rp-article::-webkit-scrollbar        { width: 0.2em; background: none; }
.rp-article::-webkit-scrollbar-thumb  { background: var(--color-tertiary, #94a3b8); border-radius: 0.1em; }

/* ── Aside panel ─────────────────────────────────────────────── */
.rp-aside {
  display: grid;
  grid-template-rows:
    auto   /* rp-aside-header  */
    1fr;   /* rp-aside-content */
  width: var(--rp-aside-width);
  height: 100%;
  overflow: hidden;
}

.rp-aside-header {
  display: flex;
  flex-direction: row;
  align-items: center;
  min-height: var(--rp-controls-height);
  padding: var(--mag-sec, 0.2rem) var(--mag-pri, 0.5rem);
  background: var(--rp-aside-bg);
}

.rp-aside-content {
  overflow-y: auto;
  padding: var(--mag-pri, 0.5rem);
  background: var(--rp-aside-bg);
}

.rp-aside-content::-webkit-scrollbar        { width: 0.2em; background: none; }
.rp-aside-content::-webkit-scrollbar-thumb  { background: var(--color-tertiary, #94a3b8); border-radius: 0.1em; }

/* ── Mobile (<640px): stack vertically ───────────────────────── */
@media (max-width: 640px) {
  .rp-wrapper {
    flex-direction: column;
    height: auto;
  }

  .rp-main {
    width: 100%;
    height: auto;
  }

  .rp-aside {
    width: 100%;
    height: auto;
  }

  .rp-article {
    overflow-y: visible;
  }

  .rp-aside-content {
    overflow-y: visible;
  }
}
```

---

### `rightpanel/ssr.go`

```go
//go:build !wasm

package rightpanel

import _ "embed"

//go:embed rightpanel.css
var css string

// RenderCSS implements dom.CSSProvider.
// tinywasm/site collects this during SSR to inject into <head>.
func (r *RightPanel) RenderCSS() string {
    return css
}
```

---

### `rightpanel/rightpanel_test.go`

```go
package rightpanel_test

import (
    "strings"
    "testing"

    "github.com/tinywasm/layout/rightpanel"
)

// stubModule implements Module for tests.
type stubModule struct{ name string }
func (s stubModule) ModelName() string { return s.name }

// stubComponent implements dom.Component for tests.
type stubComponent struct{ html string }
func (s *stubComponent) GetID() string      { return "stub" }
func (s *stubComponent) SetID(_ string)     {}
func (s *stubComponent) RenderHTML() string { return s.html }
func (s *stubComponent) Children() []any    { return nil }

func TestRightPanel_RenderHTML_WithAllSlots(t *testing.T) {
    panel := &rightpanel.RightPanel{
        Module:        stubModule{"users"},
        Title:         "Users",
        Head:          &stubComponent{"<span>badge</span>"},
        HeadControls:  &stubComponent{"<select></select>"},
        Article:       &stubComponent{"<table></table>"},
        AsideControls: &stubComponent{"<input type=search>"},
        Aside:         &stubComponent{"<ul></ul>"},
    }

    el := panel.Render()
    html := el.RenderHTML()

    checks := []struct {
        label, want string
    }{
        {"root id", "id='users'"},
        {"wrapper class", "class='rp-wrapper'"},
        {"main class", "class='rp-main'"},
        {"header class", "class='rp-header'"},
        {"title row", "class='rp-title-row'"},
        {"h1 title", "<h1>Users</h1>"},
        {"Head slot", "<span>badge</span>"},
        {"HeadControls slot", "<select></select>"},
        {"article class", "class='rp-article'"},
        {"Article slot", "<table></table>"},
        {"aside class", "class='rp-aside'"},
        {"aside header", "class='rp-aside-header'"},
        {"AsideControls slot", "<input type=search>"},
        {"aside content", "class='rp-aside-content'"},
        {"Aside slot", "<ul></ul>"},
    }

    for _, c := range checks {
        if !strings.Contains(html, c.want) {
            t.Errorf("[%s] expected %q in HTML:\n%s", c.label, c.want, html)
        }
    }
}

func TestRightPanel_RenderHTML_AsideOmittedWhenNil(t *testing.T) {
    panel := &rightpanel.RightPanel{
        Module:  stubModule{"orders"},
        Title:   "Orders",
        Article: &stubComponent{"<table></table>"},
        // No AsideControls, no Aside
    }

    html := panel.Render().RenderHTML()

    if strings.Contains(html, "rp-aside") {
        t.Error("expected rp-aside to be absent when both AsideControls and Aside are nil")
    }
}

func TestRightPanel_RenderHTML_NoModuleNoID(t *testing.T) {
    panel := &rightpanel.RightPanel{Title: "No ID"}
    html := panel.Render().RenderHTML()

    if strings.Contains(html, "id=") {
        t.Error("expected no id attribute when Module is nil")
    }
}
```

---

## Execution order for Jules

1. Update `go.mod`: run `go get github.com/tinywasm/dom@latest`
2. Create `rightpanel/rightpanel.go`
3. Create `rightpanel/rightpanel.css`
4. Create `rightpanel/ssr.go`
5. Create `rightpanel/rightpanel_test.go`
6. Run `go test ./rightpanel/...` — all tests must pass
7. Commit: `feat: add rightpanel layout`

---

## Design rules for future layouts

New layouts in this module (e.g. `fullpage/`, `storefront/`) must follow these rules:

1. **Only `tinywasm/dom` primitives** — no `text/template`, no raw HTML strings, no embed
   of `.html` files. Use `dom.Div()`, `dom.Section()`, etc.

2. **CSS in own file** — each layout has its own `.css` file. No global styles. Only
   classes prefixed with the layout abbreviation (e.g. `rp-` for rightpanel, `fp-` for
   fullpage, `sf-` for storefront).

3. **Consume `--color-*` tokens** — never hardcode colors. Use CSS variables from
   `tinywasm/dom`'s `theme.css` with fallback values.

4. **All slots optional** — a nil slot renders nothing. The layout must remain coherent
   with any combination of nil slots.

5. **Module interface for ID** — the root element ID comes from a `Module` interface
   (`ModelName() string`). No positional string arguments for the ID.

6. **SSR split** — CSS embed goes in `ssr.go` (`//go:build !wasm`).
   `RenderCSS() string` implements `dom.CSSProvider`.

7. **Mobile first** — all layouts must define a `@media (max-width: 640px)` block that
   stacks the layout vertically and makes it usable on small screens.

8. **No JS** — layout behavior (dark mode, responsiveness) must work without JavaScript.

9. **Tests** — verify presence of key classes and slot content via `strings.Contains`.
   At minimum: all slots present, aside omitted when nil, no id when Module is nil.

---

## Adding a new layout (example: storefront)

```
layout/
└── storefront/
    ├── storefront.go       # struct + Render()
    ├── storefront.css      # sf-* classes only
    ├── ssr.go              # !wasm, embed CSS, RenderCSS()
    └── storefront_test.go
```

Follow the same pattern as `rightpanel`. CSS prefix: `sf-`. Module interface: same
`Module` interface (can be copied or moved to `layout` root package if shared).
