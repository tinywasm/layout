---
PLAN: "`tinywasm/layout`: crudview mobile master-detail direction fix (form enters from the left)"
---
> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Repo rules: `AGENTS.md` at this repo's root — read it first (especially
> "Component Contract — ONE way (signals)" and "Hot reload — do NOT compile
> manually").
> No master plan: this change is scoped entirely to this repo. It touches the
> `crudview` package and — for the follow-up width fix (Stage 5) — the sibling
> `platformd` package, which is the root cause of the bare-strip defect.
> `tinywasm/dom` and `tinywasm/components` are NOT touched — the scroll-snap
> primitive (`dom.Reference.ScrollIntoView`) stays as-is and in use; the
> chevron-left icon is registered locally in `crudview/svg.go` like the other
> crud icons.

## Context (zero-context summary)

`crudview` is a two-column master-detail view. Desktop is a CSS grid,
`[FORM (2fr) | LIST (1fr)]` — form on the left. Below 640px it becomes a
horizontal `scroll-snap` strip instead (`crudview/css.go`'s
`Media("(max-width: 640px)")` block): today that strip is
`[LIST (100vw, order:1) | FORM (90vw, order:2)]` — **list on the left**.
Selecting a row calls `dom.Get(v.detailPanelID()).ScrollIntoView()`
(`crudview.go` ~207-211), sliding the strip right to reveal the form — which
therefore enters **from the right**, the opposite side from where it lives on
desktop. Reported as unintuitive: the form should always enter from the left,
matching desktop's layout.

Decision (confirmed with the user, do not revisit): **keep the native
scroll-snap swipe** — it is the most intuitive mobile gesture and costs zero
JS/bytes — and only fix the physical ordering so it matches desktop
(form=left, list=right/default), rather than replacing scroll-snap with a
JS-driven overlay. A "‹ back" button is ADDED alongside the swipe as an
explicit affordance, not a replacement for it.

**The mechanism**: the list must stay the default-visible panel with zero
mount-time JS, while the form must be physically positioned to its LEFT. A
plain LTR flex row cannot show its 2nd-order item by default without an
explicit initial scroll (which needs a JS hook this framework's component
contract does not have — see "Component Contract — ONE way" in `AGENTS.md`:
no `OnMount`). The fix is `direction: rtl` on the scroll container: in RTL,
flex `order:1` renders at the container's "start" edge, which is the
**right**, and the browser's native default scroll position (0) already
rests there — showing `order:1` (the list) by default, with zero JS. `order:2`
(the form) then renders to the LEFT of it, matching desktop. Each panel gets
`direction: ltr` reset on it directly so its OWN content (list rows, form
fields, icons) is not mirrored — only the outer strip's flow direction flips.

`scroll-snap-align: start/end` are logical (relative to inline start/end), so
they do not need to change: `start` already means "the container's start
edge", which under `direction:rtl` is the right edge — exactly where the list
(order:1) already sits. Net diff: **only `direction` properties are added**;
`order` and `scroll-snap-align` values are unchanged.

## Stages

### Stage 1 — `crudview/css.go`: `direction:rtl` on the mobile strip

Current block (`crudview/css.go` ~293-318):

```go
		// ── Mobile master-detail (horizontal scroll-snap) ──────────────────────
		// Below the breakpoint, the two-column grid becomes a horizontal
		// scroll-snap strip: the list panel (100vw) then the form panel
		// (--cv-detail-width, 90vw — a single token so the peek is easy to
		// tune). DOM order stays article-then-aside (desktop semantics
		// unchanged); `order` reverses the VISUAL sequence to list-first
		// without touching markup. Selecting a row snaps the strip to the
		// form (crudview.go's selectAction calls dom Reference.ScrollIntoView()
		// on the form panel); swiping back to the list is a plain native
		// scroll, no JS required. Zeroing gap/padding here keeps the
		// 100vw/90vw math exact — the base rule's gap+padding are for the
		// desktop grid.
		Media("(max-width: 640px)",
			Rule(clsModuleContent,
				Display(Flex_),
				Padding(Zero),
				RawRule("flex-direction:row; overflow-x:auto; overflow-y:hidden; "+
					"scroll-snap-type:x mandatory; scroll-behavior:smooth; gap:0"),
			),
			Rule(clsArticleContend,
				RawRule("flex:0 0 var(--cv-detail-width, 90vw); scroll-snap-align:end; order:2"),
			),
			Rule(clsAsideWrap,
				RawRule("flex:0 0 100vw; scroll-snap-align:start; order:1"),
			),
		),
```

Replace with (only `direction` added; `order`/`scroll-snap-align` unchanged —
see Context for why):

```go
		// ── Mobile master-detail (horizontal scroll-snap, RTL-anchored) ─────────
		// Below the breakpoint, the two-column grid becomes a horizontal
		// scroll-snap strip that must (a) show the list by default with zero
		// mount-time JS and (b) place the form to the list's LEFT, matching
		// desktop's [FORM|LIST] order — so the form always enters from the
		// left, never the right. A plain LTR row can't do both: showing an
		// order:2 item by default needs an initial scroll, and this
		// framework's component contract has no mount hook (AGENTS.md,
		// "Component Contract — ONE way"). `direction:rtl` solves it: order:1
		// renders at the container's START edge, which RTL makes the RIGHT —
		// exactly where the browser's native default scroll position (0)
		// already rests. So order:1 (list) shows by default at the right,
		// and order:2 (form) sits physically to its LEFT, matching desktop —
		// with no scroll manipulation needed. `direction:ltr` is reset on
		// EACH panel directly so its own content (rows, fields, icons) isn't
		// mirrored; only the outer strip's flow flips. `scroll-snap-align:
		// start/end` are logical keywords (relative to inline start/end), so
		// they read correctly unchanged: `start` already means "the
		// container's start edge", which is now the right, where the list
		// already sits. Selecting a row still snaps via crudview.go's
		// selectAction → dom.Reference.ScrollIntoView() on the form panel;
		// swiping is still a plain native scroll, no JS required — same
		// mechanism as before, just re-anchored. Zeroing gap/padding here
		// keeps the 100vw/90vw math exact — the base rule's gap+padding are
		// for the desktop grid.
		Media("(max-width: 640px)",
			Rule(clsModuleContent,
				Display(Flex_),
				Padding(Zero),
				RawRule("flex-direction:row; overflow-x:auto; overflow-y:hidden; "+
					"scroll-snap-type:x mandatory; scroll-behavior:smooth; gap:0; direction:rtl"),
			),
			Rule(clsArticleContend,
				RawRule("direction:ltr; flex:0 0 var(--cv-detail-width, 90vw); scroll-snap-align:end; order:2"),
			),
			Rule(clsAsideWrap,
				RawRule("direction:ltr; flex:0 0 100vw; scroll-snap-align:start; order:1"),
			),
			Rule(clsBackBtn,
				Display(Flex_),
			),
		),
```

Also add, in the base (desktop, non-media) rule set — anywhere alongside the
other always-on rules, e.g. next to `clsIcon16` — a rule hiding the new back
button outside mobile:

```go
		Rule(clsBackBtn,
			Display(None),
		),
```

(The mobile media query's own `Rule(clsBackBtn, Display(Flex_))` above
overrides this under the breakpoint.)

**Verify empirically, do not assume**: `direction:rtl` + `scroll-snap` has had
cross-browser inconsistencies historically. After this stage, use the MCP
browser tools (`browser_emulate_device`, `browser_screenshot`,
`browser_evaluate_js` with `getBoundingClientRect`) to confirm: (1) the list
is what shows on first load at the mobile breakpoint, physically on the
right; (2) the form is physically to its left; (3) selecting a row snaps the
form fully into view flush-left with the list peeking ~10vw on the right. If
any of these are wrong, the fix is tuning `scroll-snap-align`/`order` on these
same two rules — do NOT reach for JS.

### Stage 2 — `crudview/crudview.go`: back button + list panel id

1. Add a class constant next to the others near the top of the file (same
   `var (... Class = ...)` block that already declares `clsAsideWrap`,
   `clsBtnCrud`, etc.):

   ```go
   clsBackBtn Class = "cv-back"
   ```

2. Add a stable id getter for the list/aside panel, mirroring the existing
   `detailPanelID()` (`crudview.go` ~186-190):

   ```go
   // listPanelID identifies the list/aside panel — the mobile "‹ back"
   // button's scroll target (see css.go's "(max-width: 640px)" block).
   func (v *CrudView) listPanelID() string {
   	return v.GetID() + ".list"
   }
   ```

   Set this id on the `asideWrap` element in `Render()` (currently
   `asideWrap := Div().Set(clsAsideWrap.AsAttr()).Child(searchCard, listCard, actionsCard)`
   — add `.ID(v.listPanelID())` to that chain).

3. In `Render()`, inside the `clsTitleContainer` child (currently
   `articleCont.Child(Div().Set(clsTitleContainer.AsAttr()).Child(Div().Set(clsTitle.AsAttr()).Child(H1().Text(v.Title))))`),
   add a back button as the FIRST child of the title container, before the
   title `Div`:

   ```go
   back := Button().Set(clsBackBtn.AsAttr()).
   	Attr("name", "cv-back"). // NOT "btn_..." — see cv-crudtoggle's comment on the actionbutton name collision
   	Child(iconCrudBack.Render(string(clsIcon16)))
   back.On("click", func(Event) {
   	if el, ok := Get(v.listPanelID()); ok {
   		el.ScrollIntoView()
   	}
   })
   articleCont.Child(Div().Set(clsTitleContainer.AsAttr()).
   	Child(back).
   	Child(Div().Set(clsTitle.AsAttr()).
   		Child(H1().Text(v.Title))))
   ```

   This handler is intentionally NOT `undoAction` — it only changes which
   panel is in view (same effect as a manual swipe), it does not deselect or
   clear the form, matching the swipe gesture's own behavior.

4. Add the icon declaration next to the other `iconCrud*` constants
   (`crudview.go` ~39-41):

   ```go
   iconCrudBack = svg.Icon("icon-crud-back") // "‹" — mobile-only back-to-list button
   ```

5. In `crudview/svg.go` (the `!wasm`-tagged sprite-definition file, alongside
   `iconCrudNew`/`iconCrudCancel`/`iconCrudSearch`), add:

   ```go
   sprite.Define(iconCrudBack, "0 0 320 512", sprite.Path("M9.4 233.4c-12.5 12.5-12.5 32.8 0 45.3l192 192c12.5 12.5 32.8 12.5 45.3 0s12.5-32.8 0-45.3L109.2 256 246.6 118.6c12.5-12.5 12.5-32.8 0-45.3s-32.8-12.5-45.3 0l-192 192z")),
   ```

   (Font Awesome `chevron-left`, same license/style family as the other crud
   icons already in this file.)

### Stage 3 — tests

- `grep -rn "detailPanelID\|listPanelID\|cv-back" crudview/*_test.go` to find
  what needs updating.
- Add/adjust an assertion that `Render()` output contains `cv-back` and that
  the aside/list panel carries an id (mirrors whatever pattern the existing
  `detailPanelID` test uses, e.g. checking for `id='<id>.list'` in the HTML
  string).
- Do not remove or weaken the existing `selectAction` scroll-snap coverage —
  that mechanism is unchanged.
- Run `gotest` (never `go test`).

### Stage 4 — docs

- `docs/ROADMAP.md`: update the mobile master-detail bullet — direction is now
  `direction:rtl`-anchored so form enters from the left (matching desktop);
  the "‹ back" button is new; swipe still works, unchanged mechanism.
- `AGENTS.md`: this repo's `AGENTS.md` already has the sections this plan
  relies on (Component Contract, hot reload); no new rule needed there beyond
  what's already true. If you find it needs a line about "mobile detail
  direction uses `direction:rtl`, not JS scroll-position tricks", add ONE
  short bullet near "Component Contract" — do not restructure the file.

### Stage 5 — mobile width fix (root cause in `platformd`)

Found during Stage 1–4 live verification: a gray strip of bare stage showed on
one side and the view didn't use the full mobile width. The cause was NOT in
crudview: `platformd/css.go`'s mobile (base, non-media) `clsPanelActive` rule
sized the active module panel at `Width(tokenContentWidth)` — 96vw — so every
module panel was ~15px narrower than the viewport, uncovering the stage
(`ColorSecondary`) beside it AND making crudview's container 15px narrower than
its own `100vw`/`90vw` scroll-snap panels, which then overflowed by the gap.

- `platformd/css.go`: change the mobile `clsPanelActive` `Width(tokenContentWidth)`
  → `Width(Pct(100))` (fills the screen, matching the desktop-media override
  that is already `Pct(100)`). Do NOT touch the desktop override.
- Remove the now-dead token: `tokenContentWidth` in `platformd/tokens.go`, and
  its two `Declare(tokenContentWidth, "96vw")` calls in `platformd/css.go`
  (base `Root(...)` and the desktop-media `Root(...)`). Verify dead first:
  `grep -rn "tokenContentWidth\|pd-content-width" --include='*.go' platformd`
  must show ONLY those three definition/declaration lines and no other consumer.
- `crudview/css.go`: decouple the scroll-snap panels from the viewport —
  `flex:0 0 100vw` → `flex:0 0 100%` and the `--cv-detail-width` fallback
  `90vw` → `90%`. A scroll-snap child must be sized by its scroll CONTAINER,
  not the viewport; this makes crudview self-correct to whatever width the host
  panel gives it, so the class of bug cannot recur even if the platform panel
  width changes again.
- Verify live (MCP, mobile emulation): the active panel, `.cv-module-content`
  and `.cv-aside-wrap` all measure the full viewport width (no `left:-15`
  overflow, no bare strip); the form view still peeks the list ~10% on the
  right. Check a placeholder module (`#mod1`) too — its panel is now full-width
  as well (the shared fix benefits every module). Desktop is unaffected (its
  `clsPanelActive` override was not touched).

## Anti-footguns (do NOT do)

- **Do NOT touch `tinywasm/dom`.** `ScrollIntoView()` stays exactly as-is and
  in use (by both `selectAction` and the new back button). No new dom
  primitives, no `OnMount`/`Defer` — the component contract in `AGENTS.md`
  forbids lifecycle hooks beyond `Init`/`Render`.
- **Do NOT touch `tinywasm/components`.** The back icon is local to
  `crudview/svg.go`, same as the other crud icons — do not add it to the
  shared `components` package.
- **Do NOT change `order` or `scroll-snap-align` values** unless live MCP
  verification (Stage 1) shows the peek/snap is wrong — the reasoning in
  Context says they should be correct unchanged; if you must change them,
  update the Context section's reasoning to match reality, don't silently
  diverge from the documented intent.
- **Do NOT make `back`'s click handler call `undoAction`.** It must only move
  the viewport, not clear the selection/draft — that would surprise a user
  who just wanted to glance at the list.
- **Do NOT hand-build `<svg>` for the back icon** — use `svg.Icon` +
  `sprite.Define` + `.Render()`, per `AGENTS.md`'s "SVG icons" section.
- Never run `gopush` or `codejob` from this plan — report back when stages
  are done and verified.

## Verification (end-to-end)

- `go build ./...` and `GOOS=js GOARCH=wasm go build ./...` clean; `gofmt -l`
  empty; `gotest` green (vet/race/tests/wasm/coverage).
- `GOOS=js GOARCH=wasm go list -deps ./crudview | grep tinywasm/svg/sprite`
  MUST be empty (the icon leak check `AGENTS.md` requires).
- MCP at `http://localhost:8080/#crud` (restart the dev daemon after CSS
  edits, not just reload — this session has repeatedly seen stale duplicate
  stylesheet generations survive a plain browser reload):
  - **Desktop**: unchanged — form left, list right, back button hidden
    (`display:none`). Light + dark.
  - **Mobile** (emulate ≤640px): list shows by default, physically on the
    right, with the form peeking ~10vw on the left. Swiping (or selecting a
    row) brings the form in from the left, snapping flush, with the list
    peeking ~10vw on the right. The "‹" button returns to the list (same
    visual result as swiping back) without clearing the selection. Light +
    dark.

## Stages table

| # | Stage | Files | Done |
|---|---|---|---|
| 1 | `direction:rtl` mobile strip + back-button display rules | `crudview/css.go` | ☑ |
| 2 | Back button, list panel id, icon | `crudview/crudview.go`, `crudview/svg.go` | ☑ |
| 3 | Tests | `crudview/*_test.go` | ☑ |
| 4 | Docs | `docs/ROADMAP.md`, `AGENTS.md` | ☑ |
| 5 | Mobile width fix (root cause) | `platformd/css.go`, `platformd/tokens.go`, `crudview/css.go` | ☑ |
