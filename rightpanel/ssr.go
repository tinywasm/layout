//go:build !wasm

package rightpanel

import (
	. "github.com/tinywasm/css"
)

var (
	tokenTitleHeight    = Token{Name: "--rp-title-height", Fallback: "8vh"}
	tokenContentHeight  = Token{Name: "--rp-content-height", Fallback: "89vh"}
	tokenControlsHeight = Token{Name: "--rp-controls-height", Fallback: "3vh"}
	tokenMainWidth      = Token{Name: "--rp-main-width", Fallback: "66vw"}
	tokenAsideWidth     = Token{Name: "--rp-aside-width", Fallback: "30vw"}
	tokenGap            = Token{Name: "--rp-gap", Fallback: "var(--space-2)"}
	tokenBorderColor    = Token{Name: "--rp-border-color", Fallback: "var(--color-muted)"}
	tokenBg             = Token{Name: "--rp-bg", Fallback: "var(--color-surface)"}
	tokenAsideBg        = Token{Name: "--rp-aside-bg", Fallback: "var(--color-on-surface)"}
	tokenTitleColor     = Token{Name: "--rp-title-color", Fallback: "var(--color-secondary)"}
)

// SSRInstance returns a new instance of RightPanel for SSR asset collection.
func SSRInstance() *RightPanel { return &RightPanel{} }

// RenderCSS implements CSSProvider.
func (r *RightPanel) RenderCSS() *Stylesheet {
	return NewStylesheet(
		Root(
			Declare(tokenTitleHeight, "8vh"),
			Declare(tokenContentHeight, "89vh"),
			Declare(tokenControlsHeight, "3vh"),
			Declare(tokenMainWidth, "66vw"),
			Declare(tokenAsideWidth, "30vw"),
			Declare(tokenGap, Space2.Var()),
			Declare(tokenBorderColor, ColorMuted.Var()),
			Declare(tokenBg, ColorSurface.Var()),
			Declare(tokenAsideBg, ColorOnSurface.Var()),
			Declare(tokenTitleColor, ColorSecondary.Var()),
		),

		Rule(clsWrapper,
			Display(Flex_),
			FlexDirection(Str("row")),
			Width(Pct(100)),
			Height(tokenContentHeight),
			RawRule("overflow: hidden;"),
		),

		Rule(clsMain,
			Display(Grid),
			RawRule("grid-template-rows: auto 1fr;"),
			Width(tokenMainWidth),
			Height(Pct(100)),
			RawRule("overflow: hidden;"),
			RawRule("border-right: 0.1vw solid "+tokenBorderColor.Var()+";"),
		),

		Rule(clsHeader,
			Display(Flex_),
			FlexDirection(Str("column")),
			Background(tokenBg),
			Padding(Space1, Space2),
		),

		Rule(clsTitleRow,
			Display(Flex_),
			FlexDirection(Str("row")),
			AlignItems(Center),
			Gap(tokenGap),
			MinHeight(tokenTitleHeight),
		),

		Rule(Selector("."+string(clsTitleRow)+" h1"),
			FontSize(TextXl),
			Color(tokenTitleColor),
			Margin(Zero),
		),

		Rule(clsHeadControls,
			Display(Flex_),
			FlexDirection(Str("row")),
			AlignItems(Center),
			MinHeight(tokenControlsHeight),
			RawRule("padding-bottom: "+Space1.Var()+";"),
		),

		Rule(clsArticle,
			RawRule("overflow-y: auto;"),
			Padding(Space2),
			Background(ColorSurface),
			BorderRadius(RadiusMd, RadiusMd, Zero, Zero),
		),

		Rule(Selector("."+string(clsArticle)+"::-webkit-scrollbar"),
			Width(Str("0.2em")),
			Background(None),
		),

		Rule(Selector("."+string(clsArticle)+"::-webkit-scrollbar-thumb"),
			Background(ColorMuted),
			BorderRadius(RadiusSm),
		),

		Rule(clsAside,
			Display(Grid),
			RawRule("grid-template-rows: auto 1fr;"),
			Width(tokenAsideWidth),
			Height(Pct(100)),
			RawRule("overflow: hidden;"),
		),

		Rule(clsAsideHeader,
			Display(Flex_),
			FlexDirection(Str("row")),
			AlignItems(Center),
			MinHeight(tokenControlsHeight),
			Padding(Space1, Space2),
			Background(tokenAsideBg),
		),

		Rule(clsAsideContent,
			RawRule("overflow-y: auto;"),
			Padding(Space2),
			Background(tokenAsideBg),
		),

		Rule(Selector("."+string(clsAsideContent)+"::-webkit-scrollbar"),
			Width(Str("0.2em")),
			Background(None),
		),

		Rule(Selector("."+string(clsAsideContent)+"::-webkit-scrollbar-thumb"),
			Background(ColorMuted),
			BorderRadius(RadiusSm),
		),

		Media("(max-width: 640px)",
			Rule(clsWrapper,
				FlexDirection(Str("column")),
				Height(Auto),
			),
			Rule(clsMain,
				Width(Pct(100)),
				Height(Auto),
			),
			Rule(clsAside,
				Width(Pct(100)),
				Height(Auto),
			),
			Rule(clsArticle,
				RawRule("overflow-y: visible;"),
			),
			Rule(clsAsideContent,
				RawRule("overflow-y: visible;"),
			),
		),
	)
}
