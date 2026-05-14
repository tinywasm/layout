# PLAN — Fix missing root tag in Platform / RightPanel `Render()`

## Problem

After switching `Platform` and `RightPanel` from `*Element` to value-embed
`Element` (per the `components` skill), the rendered HTML for the
`platformd` demo shows a malformed root tag:

```html
<body>
  &lt; id='1' class='pd-root'&gt;
  <header class="pd-header">...
```

Captured via `mcp__tinywasm__browser_evaluate_js`. Note the literal
`&lt; id='1' ...&gt;` — the opening tag has no element name.

## Root Cause

`Platform.Render()` currently does:

```go
func (p *Platform) Render() *Element {
    p.Element.Add(clsRoot.AsAttr())
    // ...mutates p.Element via p.Add(...)
    return &p.Element
}
```

The embedded `Element` is zero-value: its unexported `tag` field is `""`.
`tag` is package-private to `tinywasm/dom`; outside the package the only
way to obtain an `*Element` with a tag set is to call a constructor like
`Div(...)`, `Section(...)`, etc.

Returning `&p.Element` therefore always emits `<` + attrs + `>` with no
tag name, which the DOM library serializes as a broken opening tag and
the browser HTML-escapes when injected.

This is the same anti-pattern the `components` skill warns against
implicitly: the skill's canonical example builds the tree with
constructors and returns that:

```go
func (c *MyComponent) Render() *Element {
    return Div(clsRoot.AsAttr(),
        H2(clsTitle.AsAttr(), c.Title),
        Div(clsBody.AsAttr()),
    )
}
```

The embedded `Element` exists only to satisfy the `Component` interface
(`GetID`, `SetID`, `Update`) — it is **not** the rendered root. Mutating
it with `p.Add(...)` and returning `&p.Element` confuses identity (the
component) with content (the rendered tree).

## Fix

Rewrite `Render()` in both `platformd` and `rightpanel` so it builds the
DOM tree exclusively through constructors and returns the freshly-built
root, never `&p.Element`.

### platformd/platformd.go

```go
func (p *Platform) Render() *Element {
    root := Div(clsRoot.AsAttr())

    header := Header(clsHeader.AsAttr())
    userBlock := Div(clsUserBlock.AsAttr())
    if p.UserBlock != nil {
        userBlock.Add(p.UserBlock)
    }
    header.Add(userBlock)

    msgDesktop := Div(clsMsgDesktop.AsAttr()).ID("pd-msg-desktop")
    for _, n := range p.notifications {
        msgDesktop.Add(p.renderNotification(n))
    }
    header.Add(msgDesktop)

    activeLabel := ""
    for _, mod := range p.Modules {
        if mod.ID == p.activeModuleID {
            activeLabel = mod.Label
            break
        }
    }
    header.Add(H2(clsArea.AsAttr()).Text(activeLabel))
    root.Add(header)

    msgMobile := Div(clsMsgMobile.AsAttr()).ID("pd-msg-mobile")
    for _, n := range p.notifications {
        msgMobile.Add(p.renderNotification(n))
    }
    root.Add(msgMobile)

    nav := Nav(clsMenu.AsAttr())
    navbar := Ul(clsNavbar.AsAttr())
    for _, mod := range p.Modules {
        item := Li(clsNavItem.AsAttr())
        link := A("#"+mod.ID, clsNavLink.AsAttr()).Attr("data-id", mod.ID)
        if mod.ID == p.activeModuleID {
            link.Add(clsNavActive.AsAttr())
        }
        if mod.Icon != nil {
            link.Add(mod.Icon)
        }
        link.Add(Span(clsLinkText.AsAttr()).Text(mod.Label))
        item.Add(link)
        navbar.Add(item)
    }
    nav.Add(navbar)
    root.Add(nav)

    stage := Main(clsStage.AsAttr())
    for _, mod := range p.Modules {
        panel := Section(clsPanel.AsAttr()).
            ID(mod.ID).
            Attr("data-id", mod.ID)
        if mod.ID == p.activeModuleID {
            panel.Add(clsPanelActive.AsAttr())
        }
        if mod.View != nil {
            panel.Add(mod.View)
        }
        stage.Add(panel)
    }
    root.Add(stage)

    root.Add(Div(clsOrientationWarn.AsAttr()))
    return root
}
```

The function no longer touches `p.Element`. All mutation happens on local
constructor results.

### rightpanel/rightpanel.go

`RightPanel.Render()` already builds its tree via a local `wrapper :=
Div(...)` and returns `wrapper`, so no functional change is required —
only verify it does not depend on the (now-removed) `if r.Element == nil`
init.

## Why the embedded `Element` still matters

`*Platform` is the value passed to `dom.Append("body", p)`. `Append`
calls `p.GetID()` / `p.SetID()` (promoted from `Element`) and tracks the
component for `Update()` / `OnMount()`. The embedded `Element` carries
the component identity. The tree built by `Render()` is what the user
sees; identity is then *injected* into that tree by `dom.Append` via
`injectComponentID(root, component.GetID())` (see
`dom_frontend.go:308`). Two distinct concerns:

| Concern | Lives in |
|---|---|
| Component identity (`GetID`, `Update`) | embedded `Element` |
| Rendered DOM tree | the `*Element` returned by `Render()` |

Conflating them (returning `&p.Element`) breaks the DOM tree because the
identity element is never given a tag.

## Files to change

- `layout/platformd/platformd.go` — rewrite `Render()` per above.
- `layout/rightpanel/rightpanel.go` — verify `Render()` still works
  without the lazy-init block (already removed).




