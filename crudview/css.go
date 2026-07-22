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
	cPanel   = "var(--color-background, #ffffff)"        // white cards / bars
	cInset   = "var(--color-surface-variant, #d7d7dd)"   // gray inset frame
	cBorder  = "var(--color-outline-variant, #cfcfd6)"   // hairline borders
	cAccent  = "var(--color-primary, #3f88bf)"           // title band / buttons
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
		Rule(clsModuleContent,
			Display(Grid),
			Width(Str("100%")),
			Height(Str("100%")),
			Padding(Str("var(--cv-mag-pri, .5rem)")),
			BorderRadius(Str(".4em")),
			// Fill the stage and split it 2:1 (form : list) with fractions instead
			// of viewport units, so the view fits any container width.
			// NOTE: adjacent RawRules are concatenated without a separator, so
			// grid-template and gap MUST share one RawRule with an explicit ';'.
			RawRule("grid-template: none / 2fr 1fr; gap: var(--cv-mag-pri, .5rem)"),
		),

		Rule(clsArticleContend,
			Flex(None),
			Display(Grid),
			MinWidth(Str("0")),
			// No 'controls' row: the single crud button lives in the aside (list)
			// column now, not under the form — see clsAsideContend below.
			RawRule("grid-template: 'title' var(--cv-title-height, 8vh) 'article' 1fr / 100%"),
		),

		Rule(clsArticleContendFullPage,
			Display(Grid),
			RawRule("grid-template: 'title' var(--cv-title-height, 8vh) 'article' 89vh / 96vw"),
		),

		Rule(clsAsideContend,
			Flex(None),
			Display(Grid),
			MinWidth(Str("0")),
			// search (top) → list (fills) → the single crud button (bottom).
			RawRule("grid-template: 'aside-search' var(--cv-controls-height, 9vh) 'aside-list' 1fr 'aside-actions' var(--cv-controls-height, 9vh) / 100%; gap: var(--cv-mag-sec, .3rem)"),
		),

		// ── Title band ────────────────────────────────────────────────────────
		Rule(clsTitleContainer,
			GridArea(Str("title")),
			Display(Flex_),
			AlignItems(Center),
			Background(Str(cAccent)),
			BorderRadius(Str(".4em .4em 0 0")),
		),

		Rule("."+string(clsTitle)+" h1",
			MarginLeft(Str(".7em")),
			Color(Str(cOnAcc)),
			FontSize(Str("1.5rem")),
		),

		// ── Form area ─────────────────────────────────────────────────────────
		Rule("article",
			GridArea(Str("article")),
			Background(Str(cPanel)),
			Display(Flex_),
			FlexDirection(Column),
			MinHeight(Str("0")),
		),

		Rule(clsBoxContent,
			FlexGrow(Str("1")),
			Display(Flex_),
			FlexDirection(Column),
			MinHeight(Str("0")),
			Overflow(Auto),
			Background(Str(cInset)),
			BorderRadius(Str(".4em")),
			Margin(Str("var(--cv-mag-sec, .3rem)")),
			Padding(Str("var(--cv-mag-pri, .5rem)")),
		),

		// The form fills the inset frame (tinywasm/form renders a bare <form> with
		// no class; it is the only direct child here).
		Rule("."+string(clsBoxContent)+" > form",
			Width(Str("100%")),
		),

		// ── Single crud button (bottom of the list column) ─────────────────────
		Rule(clsAsideActions,
			GridArea(Str("aside-actions")),
			Display(Flex_),
			Background(Str(cPanel)),
			Padding(Str(".2em")),
			BorderRadius(Str("0 0 .4em .4em")),
		),

		Rule(clsBtnCrud,
			Flex(Auto),
			Margin(Str(".2em")),
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
		Rule(clsAsideList,
			GridArea(Str("aside-list")),
			Background(Str(cPanel)),
			BorderRadius(Str(".4em .4em 0 0")),
			Display(Flex_),
			FlexDirection(Column),
			MinHeight(Str("0")),
			PaddingTop(Str("var(--cv-mag-sec, .3rem)")),
		),

		Rule(clsListaBox,
			Display(Flex_),
			Height(Str("100%")),
			FlexDirection(Column),
			Background(Str(cInset)),
			BorderRadius(Str(".4em")),
			Margin(Str("0 var(--cv-mag-sec, .3rem)")),
			Padding(Str(".3rem")),
			Overflow(Hidden),
		),

		// Row cards, the ⋮ options menu and the selected highlight all live in the
		// targetlist component (SRP) — crudview only frames the list panel above.

		// ── Search ────────────────────────────────────────────────────────────
		Rule(clsAsideSearch,
			GridArea(Str("aside-search")),
			Display(Flex_),
			Background(Str(cPanel)),
			Padding(Str(".2em")),
			BorderRadius(Str("0 0 .4em .4em")),
			AlignItems(Center),
		),

		Rule("."+string(clsAsideSearch)+" label",
			Background(Str(cAccent)),
			Color(Str(cOnAcc)),
			Padding(Str(".4em")),
			BorderRadius(Str(".3em 0 0 .3em")),
			Display(Flex_),
			AlignItems(Center),
		),

		Rule("."+string(clsAsideSearch)+" input",
			FlexGrow(Str("1")),
			Border(Str("1px solid "+cBorder)),
			BorderLeft(Str("none")),
			Padding(Str(".4em")),
			BorderRadius(Str("0 .3em .3em 0")),
		),

		Rule(clsIcon16,
			Width(Px(16)),
			Height(Px(16)),
			Decl{Prop: "fill", Val: "currentColor"},
		),
	)

	return s
}
