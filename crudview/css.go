//go:build !wasm

package crudview

import (
	. "github.com/tinywasm/css"
)

// Fallback palette. crudview reads theme tokens first, but every var() carries a
// Pa100T-derived fallback so the framing (white cards on a gray inset over the
// page) stays visible even when the active theme leaves these tokens undefined —
// which is exactly what made the view render flat and unreadable before.
const (
	cPanel  = "var(--color-background, #ffffff)"      // white cards / bars
	cInset  = "var(--color-surface-variant, #d7d7dd)" // gray inset frame
	cBorder = "var(--color-outline-variant, #cfcfd6)" // hairline borders
	cAccent = "var(--color-primary, #3f88bf)"         // title band / buttons
	// White text/icons on the saturated accent band. Not --color-on-primary:
	// some themes (e.g. the light Apple-style default) set on-primary to a near
	// black that is unreadable on the primary fill; white is the safe universal.
	cOnAcc = "#ffffff"
	cDisBg = "var(--color-outline-variant, #c2c1c1)"
	cDisFg = "var(--color-on-surface-variant, #6e6e73)"
)

// RenderCSS implements the SSR CSSProvider convention. The method name must be
// RenderCSS (not GenerateCSS) for tinywasm/ssr to detect and emit these rules.
// It is a method (not a free function) both to match the platformd/rightpanel
// convention and because the dot-imported css package already exports a free
// RenderCSS, which a package-level function here would collide with.
func (v *CrudView) RenderCSS() *Stylesheet {
	s := NewStylesheet(
		// The module's own background is the accent color — the "paper" every
		// panel sits on. White/gray panels float on it with a small margin,
		// so the accent shows through as a thin frame around them (padding
		// below, plus each panel's own margin) — that continuity, not a
		// border stroke, is what reads as "integrated" (ported from the
		Rule(clsModuleContent,
			Display(Grid),
			Width(Str("100%")),
			Height(Str("100%")),
			Padding(Space2),
			// Square, not rounded: this fills platformd's pd-panel-active
			// edge-to-edge, which is itself square (platformd/css.go) — a
			// radius here would round past the parent's flat corner and
			// show the parent's own background peeking through as a notch.
			// The inner article/aside cards keep their own rounding; only
			// this outermost layer must stay flush with the platform frame.
			BorderRadius(Zero),
			Background(Str(cAccent)),
			// Fill the stage and split it 2:1 (form : list) with fractions instead
			// of viewport units, so the view fits any container width. The row
			// track must be 1fr, not none/auto — an auto row only grows to its
			// content's intrinsic height, leaving the rest of the container's
			// height as bare background instead of stretching the columns to
			// fill it.
			// NOTE: adjacent RawRules are concatenated without a separator, so
			// grid-template and gap MUST share one RawRule with an explicit ';'.
			// Space2/Space1 (tinywasm/css's public spacing scale) replace the
			// old ad-hoc --cv-mag-* custom properties throughout this file —
			// those had drifted to different values (.5rem/.3rem/.2rem) for
			// what should read as the same "outer gutter" vs "inner gutter",
			// which is exactly what made the panel margins look inconsistent.
			RawRule("grid-template: 1fr / 2fr 1fr; gap: "+Space2.Var()),
		),

		Rule(clsArticleContend,
			Flex(None),
			Display(Grid),
			MinWidth(Str("0")),
			// Same min-height:0 grid-item fix as clsAsideWrap above (latent here
			// — the form is short today, but a long one would overflow the
			// column the same way).
			MinHeight(Str("0")),
			// No 'controls' row: the single crud button lives outside the aside
			// (list) card now — see clsAsideActions/clsAsideWrap below.
			RawRule("grid-template: 'title' var(--cv-title-height, 8vh) 'article' 1fr / 100%"),
		),

		Rule(clsArticleContendFullPage,
			Display(Grid),
			RawRule("grid-template: 'title' var(--cv-title-height, 8vh) 'article' 89vh / 96vw"),
		),

		// The aside is ONE white card (over the blue module background) that
		// holds all three pieces — search (top), list (middle), toggle button
		// (bottom). They are NOT separate cards: the white is a single shared
		// surface, and its thin padding + gap (Space1 — the "inner gutter"
		// used consistently everywhere in this file) are the only spacing,
		// showing as even white gutters between the pieces (the reference's
		// aside is a single white panel, not three floating boxes).
		Rule(clsAsideWrap,
			Flex(None),
			Display(Flex_),
			FlexDirection(Column),
			MinWidth(Str("0")),
			// min-height:0 is required on this grid item: without it the item's
			// default `min-height:auto` refuses to shrink below its content, so
			// when the list grows past the 1fr row the whole column overflows
			// downward and pushes the "+" button off the stage. With 0 it
			// respects the row and its inner list (clsAsideContend →
			// clsListaBox → tl-list, all already min-height:0) scrolls
			// internally, keeping the button pinned at the bottom.
			MinHeight(Str("0")),
			Background(Str(cPanel)),
			BorderRadius(Str(".4em")),
			Padding(Space1),
			RawRule("gap: "+Space1.Var()),
		),

		// List area — transparent passthrough that grows to fill the card
		// between search and button; clsListaBox is its gray inset.
		Rule(clsAsideContend,
			FlexGrow(Str("1")),
			Display(Flex_),
			FlexDirection(Column),
			MinWidth(Str("0")),
			MinHeight(Str("0")),
		),

		// ── Title band ────────────────────────────────────────────────────────
		// No background/radius of its own — it's the SAME accent color as
		// clsModuleContent's background, so the title sits flush on the
		// "paper" with no visible seam (the reference's title bar is just
		// white text directly on the module's own blue background).
		Rule(clsTitleContainer,
			GridArea(Str("title")),
			Display(Flex_),
			AlignItems(Center),
		),

		Rule("."+string(clsTitle)+" h1",
			MarginLeft(Str(".7em")),
			Color(Str(cOnAcc)),
			FontSize(Str("1.5rem")),
		),

		// ── Form area ─────────────────────────────────────────────────────────
		// The white card, rounded on all 4 corners — the accent "paper"
		// shows nowhere near it, so its shape is a purely stylistic rounded
		// rectangle, not a clipping concern.
		Rule("article",
			GridArea(Str("article")),
			Background(Str(cPanel)),
			BorderRadius(Str(".4em")),
			Display(Flex_),
			FlexDirection(Column),
			MinHeight(Str("0")),
		),

		// The gray inset — its small margin (Space1, the same "inner gutter"
		// clsAsideWrap uses) leaves a thin sliver of the white "article" card
		// visible around it (the reference's "white border"), giving the
		// panel visible depth instead of one flat surface.
		Rule(clsBoxContent,
			FlexGrow(Str("1")),
			Display(Flex_),
			FlexDirection(Column),
			MinHeight(Str("0")),
			Overflow(Auto),
			Background(Str(cInset)),
			BorderRadius(Str(".4em")),
			Margin(Space1),
			// 4 separate longhands, not a Padding()/PaddingTop() shorthand
			// combo (that corrupted the other 3 sides to empty values
			// downstream) nor a single multi-token Padding(a,b,c,d) call
			// (joinValues()'s output loses its spaces somewhere in the CSS
			// pipeline: `padding:var(...)var(...)var(...)`, no separators —
			// both verified via the live stylesheet). Top alone is Space3
			// (not Space2 like the other 3 sides): the first field's label
			// chip straddles ITS OWN box's top border by roughly half the
			// chip's height (~9px, fieldset/css.go) — at Space2 (8px) that
			// protrusion crossed this container's own border; Space3 (12px)
			// clears it with a small visible gap instead.
			PaddingTop(Space3),
			PaddingRight(Space2),
			PaddingBottom(Space2),
			PaddingLeft(Space2),
		),

		// The form fills the inset frame (tinywasm/form renders a bare <form> with
		// no class; it is the only direct child here).
		Rule("."+string(clsBoxContent)+" > form",
			Width(Str("100%")),
		),

		// ── Single crud button (bottom piece of the shared white aside card) ──
		// Just a flex row for the button — no background/padding of its own;
		// the white comes from clsAsideWrap, the gutter from its gap.
		Rule(clsAsideActions,
			Display(Flex_),
			MinHeight(Str("var(--cv-controls-height, 9vh)")),
		),

		Rule(clsBtnCrud,
			Flex(Auto),
			Padding(Str(".4rem")),
			Background(Str(cAccent)),
			Color(Str(cOnAcc)),
			BorderRadius(Str(".4em")),
			Border(Str("none")),
			Cursor(Pointer),
			Display(Flex_),
			AlignItems(Center),
			JustifyContent(Center),
			Transition(Str(".2s all ease")),
		),

		Rule("."+string(clsBtnCrud)+":hover:not(:disabled)",
			RawRule("filter: brightness(1.08)"),
		),

		Rule("."+string(clsBtnCrud)+":disabled",
			Background(Str(cDisBg)),
			Color(Str(cDisFg)),
			Cursor(Str("not-allowed")),
			Opacity(0.6),
		),

		// The button swaps its two icon children by hiding one — see the "toggle"
		// block in crudview.go's Render().
		Rule(clsBtnCrudIconHidden,
			Display(None),
		),

		// ── List ──────────────────────────────────────────────────────────────
		// The gray inset behind the row cards — fills the list area of the
		// shared white aside card. No margin of its own: the white gutter
		// around it comes from clsAsideWrap's padding + gap.
		Rule(clsListaBox,
			FlexGrow(Str("1")),
			Display(Flex_),
			FlexDirection(Column),
			MinHeight(Str("0")),
			Background(Str(cInset)),
			BorderRadius(Str(".4em")),
			Padding(Space1),
			Overflow(Hidden),
		),

		// Row cards, the ⋮ options menu and the selected highlight all live in the
		// targetlist component (SRP) — crudview only frames the list panel above.

		// ── Search (top piece of the shared white aside card) ──────────────────
		// A blue-label + gray-input pill sitting directly on the white card —
		// no background/padding of its own (the white + gutter come from
		// clsAsideWrap). Height matches clsAsideActions so search and button
		// read as equal-height siblings top and bottom of the same card.
		Rule(clsAsideSearch,
			Display(Flex_),
			AlignItems(Str("stretch")),
			MinHeight(Str("var(--cv-controls-height, 9vh)")),
		),

		Rule("."+string(clsAsideSearch)+" label",
			Background(Str(cAccent)),
			Color(Str(cOnAcc)),
			Padding(Str("0 .6em")),
			BorderRadius(Str(".3em 0 0 .3em")),
			Display(Flex_),
			AlignItems(Center),
		),

		Rule("."+string(clsAsideSearch)+" input",
			FlexGrow(Str("1")),
			Border(None),
			Background(Str(cInset)),
			Padding(Str("0 .6em")),
			BorderRadius(Str("0 .3em .3em 0")),
		),

		Rule(clsIcon16,
			Width(Px(16)),
			Height(Px(16)),
			Decl{Prop: "fill", Val: "currentColor"},
		),

		// "‹ back" button — mobile-only (see the "(max-width: 640px)" block
		// below, which overrides this to Display(Flex_)). Desktop shows both
		// columns at once, so there is nothing to "go back" to.
		Rule(clsBackBtn,
			Display(None),
			Border(None),
			Background(Str("transparent")),
			Color(Str(cOnAcc)),
			Cursor(Pointer),
			Padding(Str(".2em .4em .2em 0")),
			AlignItems(Center),
		),

		// ── Delete confirmation modal body ──────────────────────────────────────
		// Fixed, not static: takes this mount point (and modaldialog's own
		// hidden-state anchor div inside it) out of clsModuleContent's grid
		// item participation entirely, so it can never be auto-placed into
		// an implicit row (which was doubling the bottom gutter — see the
		// comment at its call site in crudview.go).
		Rule(clsDelConfirmMount,
			Position(Str("fixed")),
		),
		Rule(clsDelConfirmActions,
			Display(Flex_),
			JustifyContent(Str("flex-end")),
			RawRule("gap: "+Space1.Var()+"; margin-top: "+Space2.Var()),
		),
		Rule(clsDelConfirmBtn,
			Padding(Str(".4em .8em")),
			BorderRadius(Str(".3em")),
			Border(Str("1px solid "+cBorder)),
			Background(Str(cPanel)),
			Cursor(Pointer),
		),
		Rule(clsDelConfirmBtnDanger,
			Background(Str("var(--color-error, #c0392b)")),
			Color(Str(cOnAcc)),
			Border(Str("none")),
		),

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
				// Panels are sized in % of the scroll CONTAINER, not the
				// viewport (vw): the platform panel that hosts this view is not
				// guaranteed to equal the viewport width, and sizing a
				// scroll-snap child in vw against a narrower container overflows
				// it by the difference (that mismatch left a bare strip). % keeps
				// each panel exactly one container-width regardless of the host.
				RawRule("direction:ltr; flex:0 0 var(--cv-detail-width, 90%); scroll-snap-align:end; order:2"),
			),
			Rule(clsAsideWrap,
				RawRule("direction:ltr; flex:0 0 100%; scroll-snap-align:start; order:1"),
			),
			Rule(clsBackBtn,
				Display(Flex_),
			),
		),
	)

	return s
}
