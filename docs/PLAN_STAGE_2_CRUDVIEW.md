← [Stage 1](PLAN_STAGE_1_RIGHTPANEL.md) | Master → [PLAN.md](PLAN.md) | Next → [Stage 3](PLAN_STAGE_3_GUARDS.md)

# Stage 2 — `crudview` becomes a controller

Read [PLAN.md](PLAN.md) first: it carries the rules and the visual contract.

`crudview` stops rendering a grid. It builds a `rightpanel.RightPanel`, fills its
slots, and keeps only what is genuinely its own: the CRUD state machine.

```
ANTES                                DESPUÉS
crudview                             crudview
├── root  Split/Fill/Primary/Pad     └── controller
├── detail                               └── builds rightpanel.RightPanel
├── aside / aside-content                    ├── Title        ← v.Title
├── search / search-icon / -input            ├── Article      ← the form
├── actions                                  ├── AsideControls← v.Filter
├── mobile strip                             ├── Aside        ← the targetlist
└── state machine                            └── AsideFooter  ← the toggle button
```

---

## Dependencies — verify before starting

```sh
go doc github.com/tinywasm/widget Filterable                # M1, must exist
go doc github.com/tinywasm/components/searchbar SearchBar   # M2, must exist
go doc github.com/tinywasm/layout/rightpanel RightPanel     # stage 1: AsideFooter, ShowMain
```

If `AsideFooter` or `ShowMain` are absent, **stage 1 has not landed — stop.**

⚠️ **Never declare a local `Filterable`/`FilterSource` interface here.** The type
is `widget.Filterable`. Recreating it downstream is the defect this whole plan
removes.

---

## 0. Capture the baseline FIRST

This stage moves CSS between two files and the demo must come out identical. Do
not trust the eye. With the demo running at `http://localhost:8080/#crud`,
record these values **before touching any code**, and paste the JSON into the PR
description:

```js
JSON.stringify(['.crudview', '.crudview__detail', '.crudview__article',
  '.crudview__fields', '.crudview__aside', '.crudview__list',
  '.crudview__title', '.crudview__action', '.crudview__search'
].map(s => {
  const e = document.querySelector(s); if (!e) return [s, null];
  const b = e.getBoundingClientRect(), c = getComputedStyle(e);
  return [s, Math.round(b.width), Math.round(b.height), Math.round(b.top),
          Math.round(b.left), c.backgroundColor, c.padding, c.borderRadius];
}))
```

At the end of the stage, run the equivalent against the new selectors and
compare. **Any difference in width/height/top/left is a defect in this stage**,
not an improvement.

---

## 1. `crudview/crudview.go` — the slot and the seam

### 1a. Struct

- **DELETE** the field `SearchPlaceholder string`.
- **ADD**, right after `Form Component`:

```go
	// Filter is the control that narrows the list — a searchbar.SearchBar, a
	// calendar, a select. nil paints no controls band at all. If it implements
	// widget.Filterable, crudview wires it to the list filter in Init.
	//
	// The type is Component, not widget.Filterable, on purpose: a control with
	// no live output (a static legend, a chip strip that navigates elsewhere) is
	// a legitimate occupant and must not be forced to implement a callback it
	// would never fire.
	Filter Component
```

- **ADD** an internal field next to `list`:

```go
	panel *rightpanel.RightPanel // the skeleton this controller fills
```

### 1b. Wire the seam in `Init`

After the `v.list = &targetlist.TargetList{…}` assignment, before the
`v.confirmDelete` block:

```go
	// The filter control reports terms; crudview owns what a term means.
	// Assigned here, not in Render, so it survives a re-render and so a host
	// that never renders — the conformance driver — still filters.
	if src, ok := v.Filter.(widget.Filterable); ok {
		src.OnFilterChange(func(term string) {
			v.search.Set(term)
			v.filter()
		})
	}
```

`widget` is already imported by this file.

### 1c. Delete the search markup and constants

- From the `var (...)` class block **DELETE**: `clsModuleContent`,
  `clsArticleContend`, `clsArticleContendFullPage`, `clsAsideContend`,
  `clsTitleContainer`, `clsTitle`, `clsAsideActions`, `clsAsideWrap`,
  `clsAsideSearch`, `clsSearchInput`, `clsSearchIcon`, `clsIcon16`.
  **KEEP**: `clsArticle`, `clsBoxContent`, `clsBtnCrud`, `clsBtnCrudIconHidden`,
  `clsListaBox`, and every `clsDelConfirm*`.
- From the `const (...)` block **DELETE** `iconCrudSearch` and
  `defaultSearchPlaceholder`. **KEEP** `iconCrudNew` and `iconCrudCancel`.
- **DELETE** `func renderIcon(icon svg.Icon) *Element` at the end of the file —
  the search card was its only caller. Keep the `svg` import: `iconCrudNew` and
  `iconCrudCancel` are still `svg.Icon` constants.

### 1d. Rewrite `Render()`

Replace the whole body with the composition below. It is the same tree the demo
renders today, expressed through the skeleton's slots.

```go
func (v *CrudView) Render() *Element {
	hasSource := v.Presenter != nil

	// The form's card. crudview owns this one because the shape of a form area
	// is a CRUD decision, not a frame decision.
	boxContent := Div().Set(clsBoxContent.AsAttr())
	if v.Form != nil {
		boxContent.Child(v.Form)
	}
	article := Article().Set(clsArticle.AsAttr()).Child(boxContent)

	v.panel = &rightpanel.RightPanel{
		Title:   v.Title,
		Article: article,
	}

	if hasSource {
		// The list — its own inset card inside the aside's content band.
		v.panel.Aside = Div().Set(clsListaBox.AsAttr()).Child(v.list)

		// The filter is the consumer's control. crudview supplies no card
		// around it: rightpanel's controls band already keeps its size, and a
		// second frame around a control that has one reads as a box in a box.
		v.panel.AsideControls = v.Filter

		// Single toggle button — "+" when nothing is selected, "↺" when a row
		// is; Editar/Eliminar live in the targetlist row's ⋮ menu instead.
		toggle := Button().Set(clsBtnCrud.AsAttr()).
			// NOT "btn_..." — actionbutton's global `button[name*="btn"]` rule
			// matches any button whose name contains that substring and, being a
			// type+attribute selector, outranks this class; it was silently
			// injecting a stray margin. This button is crudview-owned, so its
			// name must not accidentally opt back in.
			Attr("name", "cv-crudtoggle").
			BindStateFunc(widget.Open, v.active).
			Child(
				iconCrudNew.Render(string(NameCrudView.Class("action-new"))).
					BindStateFunc(widget.Open, v.active),
				iconCrudCancel.Render(string(NameCrudView.Class("action-cancel"))).
					BindStateFunc(widget.Open, v.active),
			)
		toggle.On("click", func(Event) { v.toggleAction() })
		v.panel.AsideFooter = toggle
	}

	root := v.panel.Render()

	if hasSource {
		// v.confirmDelete's Show() wraps its content in a bare, class-less div
		// even while hidden. As an un-placed child of the frame's grid it got
		// auto-placed into an implicit second row and doubled the bottom gutter.
		// clsDelConfirmMount is position:fixed, which removes it from grid item
		// participation entirely, regardless of visibility.
		root.Child(Div().Set(clsDelConfirmMount.AsAttr()).Child(v.confirmDelete))
	}

	return root
}
```

Add `"github.com/tinywasm/layout/rightpanel"` to the import block.

⚠️ **`v.panel` must be assigned in `Render()`, before `ShowMain`/`ShowAside` can
be called.** `Render()` runs before any user interaction, so the nil guards in
1e are belt-and-braces, not dead code — keep them.

### 1e. Delegate the snap

**DELETE** `showPanel`, `detailPanelID` and `listPanelID` — stage 1 moved them to
`rightpanel`. Replace every call site:

| Was | Becomes |
|---|---|
| `v.showPanel(v.detailPanelID())` in `selectAction` | `if v.panel != nil { v.panel.ShowMain() }` |
| `v.showPanel(v.detailPanelID())` in `newAction` | `if v.panel != nil { v.panel.ShowMain() }` |
| `v.showPanel(v.listPanelID())` in `undoAction` | `if v.panel != nil { v.panel.ShowAside() }` |

Keep the surrounding comments — the reason each call exists (*"it focused the
first field; on a phone that field is on the panel next door"*, *"the list is the
resting view"*) does not change.

### Acceptance

- `grep -rn "SearchPlaceholder\|clsAsideSearch\|iconCrudSearch\|renderIcon\|showPanel\|detailPanelID\|listPanelID" crudview/` → empty.
- `grep -rn "Input(\"search\")" crudview/` → empty.
- `grep -rn "FilterSource\|interface{ OnFilterChange" crudview/` → empty.

---

## 2. `crudview/css.go` — delete the frame

`RenderSheet()` keeps only the parts `crudview` still paints. **DELETE** these
`Part(...)` blocks with their comments:

`detail`, `detail-full`, `aside`, `aside-content`, `actions`, `title`,
`title-text`, `icon`, `search`, `search-icon`, `search-input`.

**DELETE** the `Root(...)` options and replace the whole `Root(` block with
nothing — `crudview` no longer owns a root. `style.For(v)` still needs a first
call; use:

```go
	return style.For(v).
		// crudview paints no frame: rightpanel owns the root, the columns and
		// the mobile strip. What remains here are the widgets this controller
		// puts INTO those slots.
		Part(widget.Part("fields"), …).
```

**DELETE** every `On(css.Mobile, …)` rule **except** the three FAB ones:

```go
		On(css.Mobile, widget.Part("action"), …)         // KEEP
		On(css.Mobile, widget.Part("action-new"), …)     // KEEP
		On(css.Mobile, widget.Part("action-cancel"), …)  // KEEP
```

`Part("action")`'s mobile rule docks to `style.Parent`. Its anchor is now
`rightpanel`'s root — which is `position: relative` only if something sets it.
**Verify after the change** that the FAB still lands at the bottom-right of the
aside panel on a 375×812 viewport. If it escapes to the page, the fix is a
`style.Anchor()` on `rightpanel`'s root, **reported and fixed in `rightpanel`**,
never patched with a wrapper here.

**KEEP unchanged**: `fields`, `article`, `list`, `action`, `action-new`,
`action-cancel`, `delconfirm-mount`, `delconfirm-body`, `delconfirm-actions`,
`delconfirm-btn`, `delconfirm-btn-danger`, and the two `When(...)` rules.

### The `article` reconciliation — measure, do not guess

`rightpanel`'s `article` band is `As(Page), Pad(Space2), Scroll(), Fill()`;
`crudview`'s own `article` is `As(Panel), Round(RadiusMd), Pad(Space1),
Stack(SpaceNone), Fill()`. After composition they nest:
`rp__article > crudview__article > crudview__fields`.

That is one layer more than today. Compare the stage-0 baseline against the
result:

- If `.crudview__fields` keeps its box (width/height/top/left within 1px), leave
  both layers alone and move on.
- If it shifts, the cause is `rp__article`'s `Pad(Space2)` doubling with
  `crudview__article`'s `Pad(Space1)`. The fix is to **delete `crudview`'s own
  `article` part** and let `rp__article` be the card, adjusting nothing in
  `rightpanel`. Re-measure.
- Report the outcome either way in the PR description. Do **not** add a
  compensating negative margin.

---

## 3. `crudview/crud.go`

In `type Config struct`, after `Presenter`:

```go
	// Filter is the control that narrows the list. Optional: nil renders no
	// controls band. When nil, New installs a searchbar.SearchBar carrying the
	// presenter's placeholder — the ergonomic default, not a decision imposed:
	// pass any widget.Filterable to replace it.
	Filter dom.Component
```

Add `"github.com/tinywasm/dom"` as a **named** import (this file uses plain
identifiers; a dot import would shadow them) and
`"github.com/tinywasm/components/searchbar"`.

In `New`, before the `v := &CrudView{…}` literal:

```go
	filter := cfg.Filter
	if filter == nil {
		filter = &searchbar.SearchBar{Placeholder: cfg.Presenter.SearchPlaceholder()}
	}
```

In the literal: **DELETE** `SearchPlaceholder: cfg.Presenter.SearchPlaceholder(),`
and **ADD** `Filter: filter,`.

> The default keeps `view.Presenter.SearchPlaceholder()` alive and meaningful,
> and keeps every existing consumer compiling. It is the ergonomic path for a
> host that does not care which control it gets — **not** the path the demo
> takes. See stage 4.

---

## 4. The demo uses the slot

This is not cleanup. The demo is the **consumer-shaped proof** the construction
harness requires before an API is considered published:

> *"An API is not published until a consumer-shaped test, inside the library
> itself, proves it… if that test is awkward to write, the API is awkward to
> use, and you have found the defect before shipping it."*

A slot that only the default ever fills is a slot nobody has proven. If injecting
a `SearchBar` from `client.go` turns out to be awkward, that is a defect in this
stage's API, discovered here rather than by the first real consumer.

### 4a. `platformd/web/client.go`

In `func (m mod) View() Component`, inside the `if m.name == "crud"` branch:

1. **DELETE** the presenter option line `view.WithSearchPlaceholder("Buscar..."),`.
   The placeholder now travels with the control that displays it. Leaving it
   would give two sources for one string and let them disagree — the demo would
   still render correctly and teach the wrong pattern.
2. Pass the control explicitly:

```go
		cv, err := crudview.New(crudview.Config{
			ParentID:  "crud",
			Presenter: pres,
			// The filter is the application's choice, not the layout's. This is
			// the line a deployment swaps for a calendar or a category select;
			// nothing in crudview or rightpanel changes when it does.
			Filter: &searchbar.SearchBar{Placeholder: "Buscar..."},
		})
```

3. Add `"github.com/tinywasm/components/searchbar"` to the import block.

Leave the `cv.OnNew` / `OnSaved` / `OnDeleted` / `OnCancel` wiring untouched.

### 4b. If `view.WithSearchPlaceholder` becomes unused

After 4a, `client.go` no longer calls it. Check:

```sh
grep -rn "WithSearchPlaceholder" .
```

Hits remaining inside `crudview/crud.go`'s default path (via
`Presenter.SearchPlaceholder()`) are expected — that is the default still
working. **Do not delete `WithSearchPlaceholder` from `tinywasm/view`**; it is
another repository and the option is still the right way to configure the
default. Note the reduced usage in the PR description and stop there.

### 4c. Verify in the running app

Per `docs/ROADMAP.md`, every change is verified on the real page, on desktop
**and** an emulated phone, in **both** themes — not by reading the diff.

| Check | How | Expected |
|---|---|---|
| Desktop `#crud` | screenshot at 1024×768 | identical to the stage-0 baseline: blue frame, form left, search+list+`+` right |
| Desktop `#mod1` | screenshot | title **above** content, blue frame (changed by stage 1, correct here) |
| Filtering | type `backend` in the bar | the list narrows; clearing restores all rows |
| Mobile | emulate 375×812, `#crud` | one panel visible; tapping a row swipes to the form; the FAB sits bottom-right; the floating title is present |
| Dark theme | toggle via the user menu | no hardcoded colour survives — every surface follows the token |
| Console | read the log | clean; no warning about a missing sprite id or an unresolved `use href` |

Paste the desktop and mobile screenshots into the PR description alongside the
stage-0 measurement.

---

## 5. Tests

### 5a. `crudview/consumer_stylesheet_test.go`

- In the control-height loop, drop `".crudview__search {"`, keep
  `".crudview__action {"`.
- **DELETE** both `.crudview__search-icon` asserts.
- **ADD** the extraction guard:

```go
	// The frame and the filter left this package. A rule reappearing here means
	// someone re-hardcoded a skeleton into the controller.
	for _, gone := range []string{".crudview__search", ".crudview__detail",
		".crudview__aside", ".crudview__title", ".crudview__actions"} {
		if fmt.Contains(cssStr, gone) {
			t.Errorf("%s must not exist: rightpanel owns the frame", gone)
		}
	}
```

- The markup↔stylesheet pairing test now sees `rp__*` classes in the markup and
  `crudview__*` in the sheet. It already filters both sides by the `crudview`
  prefix, so `rp__*` is ignored — **verify that is still true** after the change
  and, if the filter is on the element instead, widen it to accept both prefixes.
  Give the rendered `v` a `Filter` so no band is missing.

### 5b. New `crudview/compose_test.go`

No build tag. Reuse `fakeCtx` and the presenter fixtures from `consumer_test.go`,
and `cardLabels` from `conformance_test.go`.

```go
// fakeFilter is a filter control with no markup beyond a marker attribute.
type fakeFilter struct {
	dom.Element
	sink func(term string)
}

func (f *fakeFilter) OnFilterChange(fn func(term string)) { f.sink = fn }
func (f *fakeFilter) Render() *dom.Element {
	return html.Div().Attr("data-testid", "fake-filter")
}
```

⚠️ **`html.Element` does not exist.** `github.com/tinywasm/html` dot-imports
`dom` and its builders return `*dom.Element`; a dot import does not re-export the
type. Name the type through `dom`, build nodes through `html`.

Cases:

1. `TestCrudView_RendersThroughRightPanel` — markup contains `class='rp'`,
   `rp__main`, `rp__aside`, `rp__aside-header`, `rp__aside-footer`, and does
   **not** contain `crudview__detail` or `crudview__search`.
2. `TestCrudView_FilterSlotIsRendered` — with `Filter: &fakeFilter{}` the markup
   contains `data-testid='fake-filter'` inside the `rp__aside-header` band.
3. `TestCrudView_NoFilterPaintsNoControlsBand` — with `Filter: nil` the markup
   contains neither `rp__aside-header` nor `type='search'`.
4. `TestCrudView_FilterableDrivesTheList` — build over the test presenter with
   `Filter: &fakeFilter{}`, `Init`, then `f.sink("backend")` → `cardLabels(v)`
   returns only the matching row; `f.sink("")` → all rows return.
5. `TestCrudView_NewInstallsADefaultFilter` — `New(Config{…})` with no `Filter`
   yields a `CrudView` whose `Filter` is non-nil and satisfies
   `widget.Filterable`.

> Attribute quoting: `tinywasm/dom` renders attributes with **single quotes**.
> Assert with single quotes or these tests silently never match.

### 5c. Untouched

`TestConsumer_SearchFiltering` drives `v.search.Set(…)` directly and still works
— the signal and `filter()` survive. **Do not rewrite it.**
`fakeNoWidgetsPresenter.SearchPlaceholder()` stays: it implements
`view.Presenter`, which still declares the method.
`conformance_test.go` needs no change.

---

## Stages table

| # | File | What lands | Blocks |
|---|---|---|---|
| 0 | — | baseline measurement | 2, 4c |
| 1 | `crudview/crudview.go` | slot, seam, composition, snap delegation | 2, 4, 5 |
| 2 | `crudview/css.go` | delete the frame; reconcile `article` by measurement | 4c, 5a |
| 3 | `crudview/crud.go` | `Config.Filter` + default | 4a |
| 4 | `platformd/web/client.go` | the demo injects the control; verify in the app | — |
| 5 | tests | retarget + new compose tests | — |

1 and 2 land together — 2 deletes what 1 stops referencing. 4 is **not**
optional: it is the consumer-shaped proof, and it is the only step that
exercises the slot for real.

---

## Definition of done

1. `gotest` green at the module root.
2. `grep -rn "crudview__search\|crudview__detail\|crudview__aside\|crudview__title" .` → empty.
3. `grep -c "Split(\|MasterDetail(" crudview/css.go` → 0.
4. `grep -rn "components/searchbar" crudview/` → hits **only** `crud.go`.
5. `grep -n "searchbar.SearchBar" platformd/web/client.go` → one hit: the demo
   fills the slot explicitly instead of relying on the default.
6. `grep -n "WithSearchPlaceholder" platformd/web/client.go` → empty.
7. The stage-0 measurement, re-run, matches within 1px on every row — or the
   deviation is explained in the PR description under the `article`
   reconciliation.
8. The 4c verification table is fully green, with the desktop and mobile
   screenshots in the PR description. A stage that compiles and passes unit tests
   but was never opened in a browser is **not** done.

## Out of scope

- **`view.Presenter.SearchPlaceholder()`.** Still declared, still feeding
  `crudview.New`'s default, still another repository's decision. The demo stops
  using it; the option does not go away.
- **Building a second filter control to prove the slot twice.** One real consumer
  is the proof. A calendar is its own component, later.
- **The delete-confirmation modal.** Its markup and flow are unchanged.
- **Building a calendar or select filter.** This stage only opens the seam.
- **Any value change in `rightpanel`.** If the composition needs something the
  skeleton lacks (e.g. `Anchor()` on the root for the FAB), fix it **there** and
  say so — never wrap or compensate here.
