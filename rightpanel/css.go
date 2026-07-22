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
			Display(Grid),
			RawRule("grid-template-rows: "+tokenTitleHeight.Var()+" "+tokenContentHeight.Var()),
			GridTemplateColumns(Str("1fr")),
			RawRule(`grid-template-areas: "title" "content"`),
			Width(Pct(100)),
			Height(Pct(100)),
			Overflow(Hidden),
		),

		Rule(clsMain,
			RawRule("grid-area: content"),
			Display(Grid),
			RawRule("grid-template-columns: "+tokenMainWidth.Var()+" "+tokenAsideWidth.Var()),
			RawRule(`grid-template-areas: "main aside"`),
			Width(Pct(100)),
			Height(Pct(100)),
			Overflow(Hidden),
		),

		Rule(clsHeader,
			RawRule("grid-area: title"),
			Display(Flex_),
			FlexDirection(Str("column")),
			Background(tokenBg),
			Padding(Zero, Space2),
		),

		Rule(clsTitleRow,
			Display(Flex_),
			FlexDirection(Str("row")),
			AlignItems(Center),
			Gap(tokenGap),
			Height(tokenTitleHeight),
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
			PaddingBottom(Space1),
		),

		Rule(clsArticle,
			RawRule("grid-area: main"),
			OverflowY(Auto),
			Padding(Space2),
			Background(ColorSurface),
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
			RawRule("grid-area: aside"),
			Display(Grid),
			GridTemplateRows(Str("auto 1fr")),
			Height(Pct(100)),
			Overflow(Hidden),
			RawRule("border-left: 0.1vw solid "+tokenBorderColor.Var()),
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
			OverflowY(Auto),
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

		// NOTE: form-field appearance is NOT defined here. It lives in the global
		// form skin `github.com/tinywasm/components/fieldset`, which the consumer
		// opts into once. rightpanel only styles its own panel structure.

		Media("(max-width: 640px)",
			Rule(clsWrapper,
				RawRule("grid-template-rows: "+tokenTitleHeight.Var()+" auto"),
				Height(Auto),
			),
			Rule(clsMain,
				GridTemplateColumns(Str("1fr")),
				RawRule(`grid-template-areas: "main" "aside"`),
				Width(Pct(100)),
				Height(Auto),
			),
			Rule(clsAside,
				Width(Pct(100)),
				Height(Auto),
			),
			Rule(clsArticle,
				OverflowY(Visible),
			),
			Rule(clsAsideContent,
				OverflowY(Visible),
			),
		),
	)
}
