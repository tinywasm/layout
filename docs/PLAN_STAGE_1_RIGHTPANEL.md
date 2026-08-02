Master → [PLAN.md](PLAN.md) | Next → [Stage 2](PLAN_STAGE_2_CRUDVIEW.md)

# Stage 1 — `rightpanel` becomes the one skeleton

Read [PLAN.md](PLAN.md) first: it carries the rules and the visual contract that
apply here.

**This stage does not touch `crudview`.** It repairs and completes `rightpanel`
so that stage 2 has something correct to compose. After this stage the demo
still works exactly as before for `#crud` (crudview is untouched) and `#mod1`/
`#mod2` change appearance — that is intended and is checked below.

No upstream dependency: `widget.Filterable` and `components/searchbar` are not
needed until stage 2.

---

## 1. Fix the inverted flow (`rightpanel/css.go`)

This is a **defect**, not a preference. Today:

```go
Root(style.Stack(style.SpaceNone), …)                  // children: main, aside
Part(widget.Part("main"), style.Split(…), …)           // children: header, article
```

The root stacks the two columns vertically and `main` splits the title away from
the body. Measured on the running demo at `#mod1`, `rp__header` and
`rp__article` share `top=42` and sit at `left=764` and `left=1270` — side by
side.

The two options are **swapped**:

- `Root`: `style.Stack(style.SpaceNone)` → `style.Split(style.SplitTwoThirds, style.Space2)`
- `Part("main")`: `style.Split(style.SplitTwoThirds, style.Space2)` → `style.Stack(style.Space2)`

`Space2` on the main stack, not `SpaceNone`: it is the gap between the title band
and the body, and it must match the Split's gutter so the frame reads as one
rhythm.

### Acceptance

`grep -n "Split(" rightpanel/css.go` → exactly one hit, inside `Root(`.

---

## 2. Unify the module frame

Decided by the author (see [ARQ_REFACTOR.md](ARQ_REFACTOR.md)): every module in
the shell wears the same frame, and the frame is the blue one `#crud` already
has. `#mod1`/`#mod2` change; that is the point.

Three parts adopt `crudview`'s current values so that stage 2 can compose without
`#crud` shifting a pixel. **Copy the values exactly; do not average them with
what is there.**

| Part | Now | Becomes | Source |
|---|---|---|---|
| `Root` | `Stack(SpaceNone), Fill(), As(Panel), EdgeToEdge()` | `Split(SplitTwoThirds, Space2), Fill(), As(Primary), Pad(Space2), EdgeToEdge()` | crudview `Root` |
| `main` | `Split(SplitTwoThirds, Space2), Fill()` | `Stack(Space2), Fill()` | crudview `detail` |
| `header` | `Stack(Space1), Pad(Space2), As(Panel), EdgeToEdge()` | `Row(Space1), PadInline(Space2), ControlBox(), KeepSize()` | crudview `title` |
| `title` | `FontSize(TextXl), As(Secondary), EdgeToEdge()` | `As(Primary), FontSize(Text2xl), FontWeight(WeightBold)` | crudview `title-text` |

Carry these comments over with the values — they are the contract:

```go
		// Pad is what turns the primary surface into a visible frame: without it
		// the content and the aside sit flush against the panel edges and the
		// module reads as stacked rectangles instead of one view.
		// One gutter everywhere: the frame's pad is the Split's gap, so the
		// margin touching the four sides measures the same as the seam between
		// the two columns.
```

```go
		// The heading needs its own indent: it sits directly on the primary
		// surface with no card of its own to inset it. PadInline, not Pad — the
		// indent is horizontal; vertically the band answers to --control-height,
		// the same token every control measures by, so the frame reads as one
		// rhythm.
```

```go
		// Explicit now that the reset stops <h1> from carrying 2em of its own.
```

⚠️ **Do NOT touch `Part("article")`, `Part("title-row")` or `Part("controls")`
in this stage.** `article` in particular is reconciled in stage 2 with a measured
before/after; changing it here would make that measurement meaningless.

---

## 3. The aside becomes a three-band column

`rightpanel` today renders two bands (`aside-header`, `aside-content`). A CRUD
column has three: filter on top, list in the middle, actions at the bottom. The
third band is missing, and that gap is why `crudview` built its own aside.

### 3a. New slot in `rightpanel/rightpanel.go`

In the `RightPanel` struct, immediately after `Aside`:

```go
	// AsideFooter is rendered at the bottom of the aside panel, below the
	// content (e.g. a primary action button). It keeps its size while the
	// content between it and AsideControls takes the slack.
	AsideFooter Component
```

Add the class in the `var (...)` block, after `clsAsideContent`:

```go
	clsAsideFooter  = NameRightPanel.Class("aside-footer")
```

In `Render()`, extend the aside guard and append the band. The guard currently
reads:

```go
	if r.AsideControls != nil || r.Aside != nil {
```

It becomes:

```go
	if r.AsideControls != nil || r.Aside != nil || r.AsideFooter != nil {
```

and after the `if r.Aside != nil { … }` block, before `wrapper.Child(aside)`:

```go
		if r.AsideFooter != nil {
			aside.Child(Div().Set(clsAsideFooter.AsAttr()).Child(r.AsideFooter))
		}
```

Also extend the struct's usage doc comment (the `// Usage:` block) with
`AsideFooter: myActionButton,` so autocomplete has a worked example — the harness
requires the signature to guide without a manual.

### 3b. The aside's rules (`rightpanel/css.go`)

Adopt `crudview`'s aside values, and add the new part:

| Part | Now | Becomes | Source |
|---|---|---|---|
| `aside` | `As(Panel), Stack(SpaceNone), Fill()` | `Stack(Space1), As(Panel), Pad(Space1), Fill()` | crudview `aside` |
| `aside-header` | `Row(Space2), Pad(Space1), As(Panel)` | `KeepSize()` | see below |
| `aside-content` | `As(Panel), Pad(Space2), Scroll(), Fill()` | `Fill(), Stack(SpaceNone)` | crudview `aside-content` |
| `aside-footer` | — (new) | `Row(Space1), KeepSize()` | crudview `actions` |

`aside-header` sheds every treatment on purpose. Comment it:

```go
		// The controls band carries no surface, padding or radius of its own:
		// whatever the consumer puts here — a search bar, a calendar, a select —
		// brings its own. A second frame around a control that already has one
		// reads as a box inside a box. All the band owes it is a refusal to be
		// squashed when the content below grows, which is what KeepSize buys.
```

---

## 4. The mobile strip moves here

`crudview` owns a mobile master-detail behaviour that `rightpanel` lacks. It is
the only capability blocking the fusion, so it moves to the skeleton — the
skeleton is what owns the frame.

**Move these rules from `crudview/css.go` into `rightpanel/css.go`**, remapping
the part names. Do not retype the values; copy them.

| From `crudview` | To `rightpanel` |
|---|---|
| `On(css.Mobile, "", MasterDetail(Most), Pad(SpaceNone))` | same, on `""` |
| `On(css.Mobile, Part("title"), Docked(Parent, EdgeTop, SideStart, Space4), Row(Space1), As(Primary), Round(RadiusMd), Pad(Space2), Raise(Floating), Width(Content))` | on `Part("header")` |
| `On(css.Mobile, Part("title-text"), FontSize(TextBase), FontWeight(WeightBold))` | on `Part("title")` |
| `On(css.Mobile, Part("detail"), PadEdge(EdgeTop, Space12))` | on `Part("main")` |
| `On(css.Mobile, Part("aside"), PadEdge(EdgeTop, Space12))` | on `Part("aside")` |

⚠️ **Leave the `crudview` copies in place for now.** Stage 2 deletes them.
Deleting them here breaks `#crud` between two dispatches, and this stage must
leave the demo working.

⚠️ **Do NOT move the FAB rules** — `On(css.Mobile, Part("action"), …)`,
`Part("action-new")`, `Part("action-cancel")`. The toggle button is `crudview`'s
own widget, not a frame concern. They stay where they are, permanently.

Carry these comments with the rules:

```go
		// On a phone the desktop Split becomes a horizontal scroll-snap strip:
		// the aside is what shows on arrival, and selecting an item slides the
		// main panel in from the left, leaving a sliver of the aside on the right
		// so it is obvious where you came from.
		// Pad(SpaceNone) is part of the contract, not a detail: the panels are
		// sized as a share of the scroll container, and any padding on it makes
		// each panel that much narrower than the window, so a strip of the
		// neighbour shows through at rest.
```

```go
		// Reserve the band the floating header and the hamburger occupy, so the
		// content starts below them instead of underneath.
```

---

## 5. The snap behaviour moves too (`rightpanel/rightpanel.go`)

`crudview` drives the strip with `showPanel`/`detailPanelID`/`listPanelID`. Since
`rightpanel` now owns the strip, it owns the scrolling. **Add** these to
`rightpanel` — the `crudview` copies are deleted in stage 2:

```go
// MainPanelID and AsidePanelID identify the two scroll-snap targets of the
// mobile strip. A host that drives the snap (see ShowMain/ShowAside) does not
// need them; they are exported because a host may want to link to a panel.
func (r *RightPanel) MainPanelID() string  { return r.panelID() + ".main" }
func (r *RightPanel) AsidePanelID() string { return r.panelID() + ".aside" }

// ShowMain brings the main panel into view; ShowAside brings the aside back.
//
// Both are no-ops on a wide screen, and that guard is the whole point:
// ScrollIntoView walks EVERY scrollable ancestor, not just the nearest one the
// caller had in mind. Side by side there is nothing to scroll here, so an
// unguarded call reached the platform's module deck instead and slid the whole
// application to the next module.
func (r *RightPanel) ShowMain()  { r.showPanel(r.MainPanelID()) }
func (r *RightPanel) ShowAside() { r.showPanel(r.AsidePanelID()) }

func (r *RightPanel) showPanel(id string) {
	strip, ok := Get(r.GetID())
	if !ok || !strip.ScrollsX() {
		return
	}
	if el, ok := Get(id); ok {
		el.ScrollIntoView()
	}
}
```

`panelID()` is a small unexported helper returning `r.Module.ModelName()` when
`Module != nil` and `r.GetID()` otherwise — the same id `Render()` already
stamps on the wrapper. Write it next to the above.

In `Render()`, stamp the two ids so the targets exist:

- `main.ID(r.MainPanelID())` right after `main := Section()…`
- `aside.ID(r.AsidePanelID())` right after `aside := Aside()…`

> `Get` and `ScrollIntoView` come from the dot-imported `dom`; the file already
> dot-imports it. No new import.

---

## 6. Tests

### 6a. `rightpanel/rightpanel_test.go`

- `TestRightPanel_RenderHTML_WithAllSlots`: add `AsideFooter: &stubComponent{"<button></button>"}`
  to the literal and two checks: `{"aside footer", "class='rp__aside-footer'"}`
  and `{"AsideFooter slot", "<button></button>"}`.
- `TestRightPanel_RenderHTML_AsideOmittedWhenNil` stays valid — it passes none of
  the three aside slots.
- **New** `TestRightPanel_AsideRendersForFooterAlone`: a panel with only
  `AsideFooter` set renders `rp__aside` and `rp__aside-footer`. Guards the
  extended guard condition in 3a.
- **New** `TestRightPanel_PanelIDsAreStamped`: with `Module: stubModule{"users"}`,
  the markup contains `id='users.main'` and `id='users.aside'`.

### 6b. `rightpanel/consumer_stylesheet_test.go`

Add asserts that the flow defect is fixed and cannot come back:

```go
	// The root splits the two columns; main stacks the title band above the body.
	// These were swapped, and the swap was invisible because no demo module
	// passed an Aside.
	if b := ruleBlock(cssStr, ".rp {"); !fmt.Contains(b, "flex-wrap: wrap") {
		t.Errorf(".rp must carry the Split flow, block:\n%s", b)
	}
	if b := ruleBlock(cssStr, ".rp__main {"); !fmt.Contains(b, "flex-direction: column") {
		t.Errorf(".rp__main must stack, not split, block:\n%s", b)
	}
```

> `ruleBlock` is already defined in that file (line 24) and
> `github.com/tinywasm/fmt` is already imported. Use both as they are — do not
> add a second helper.

Whatever test in that file pairs markup classes against stylesheet classes must
be given an `AsideFooter` so `.rp__aside-footer` appears in both.

---

## Stages table

| # | File | What lands |
|---|---|---|
| 1 | `rightpanel/css.go` | Split↔Stack swap |
| 2 | `rightpanel/css.go` | frame + header + title adopt crudview's values |
| 3 | `rightpanel/rightpanel.go`, `css.go` | `AsideFooter` slot + three-band aside |
| 4 | `rightpanel/css.go` | mobile strip copied in (crudview's copies stay) |
| 5 | `rightpanel/rightpanel.go` | `ShowMain`/`ShowAside` + panel ids |
| 6 | `rightpanel/*_test.go` | new and adjusted cases |

1 and 2 land together. 3, 4, 5 are independent of each other.

---

## Definition of done

1. `gotest` green at the module root.
2. `grep -n "Split(" rightpanel/css.go` → one hit, inside `Root(`.
3. `grep -c "css.Mobile" rightpanel/css.go` → 5.
4. `grep -n "Part(\"action\")" rightpanel/css.go` → empty (the FAB did not travel).
5. In the running demo, `#mod1` shows the title **above** the content inside a
   blue frame. Verify by measurement, not by eye:
   ```js
   const h = document.querySelector('#mod1 .rp__header').getBoundingClientRect();
   const a = document.querySelector('#mod1 .rp__article').getBoundingClientRect();
   // expected: a.top >= h.bottom   (stacked, not side by side)
   ```
6. `#crud` is unchanged — `crudview` was not touched in this stage. Confirm with a
   screenshot, not by reasoning: this stage moves the mobile rules into
   `rightpanel` while `crudview`'s copies are still live, and two `@media` blocks
   targeting the same viewport are exactly the kind of overlap that shows up
   only in a browser.
7. Per [ROADMAP.md](ROADMAP.md), verified in the running app on desktop **and**
   an emulated 375×812 phone, in **both** themes. Attach the screenshots to the
   PR description. A stage that compiles and passes unit tests but was never
   opened in a browser is not done.

## Out of scope

- **Anything in `crudview/`.** Stage 2.
- **`Part("article")`, `Part("title-row")`, `Part("controls")`.** Stage 2
  reconciles `article` with a measured baseline; the other two are already right.
- **The FAB's mobile rules.** They belong to `crudview` permanently.
- **Deleting the mobile rules from `crudview/css.go`.** Stage 2, so the demo
  never breaks between dispatches.
