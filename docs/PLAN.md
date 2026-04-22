# PLAN: layout/rightpanel — Component slot embedding rule

## Problem

Passing a `dom.Component` implementor with a nil `*dom.Element` pointer to any RightPanel slot
(Head, HeadControls, Article, AsideControls, Aside) causes a nil pointer panic in `dom` during render.

**Root cause:** documented in `tinywasm/dom/docs/PLAN.md`. Summary: `dom.renderToHTML` calls
`GetID()` on every Component child before rendering, which panics on a nil embedded pointer.

## Rule for consumers of RightPanel

All structs passed as RightPanel slots must embed `dom.Element` as a **value**, not a pointer.

```go
// ❌ Will panic at runtime
type MyArticle struct {
    *dom.Element
    ...
}

// ✅ Correct
type MyArticle struct {
    dom.Element
    ...
}
```

## Action items

- [x] Add this rule to the RightPanel usage docs in `rightpanel.go`
  - Added IMPORTANT note in RightPanel docstring referencing tinywasm/dom rules
- [ ] Add a compile-time or runtime check that rejects nil-embedded components
  - Deferred: runtime guard in dom_frontend.go is sufficient for now
- [ ] Consider accepting `dom.ViewRenderer` instead of (or in addition to) `dom.Component`
  for slots, since all practical consumers implement `Render() *dom.Element`
  - Deferred: architectural decision, beyond immediate scope
