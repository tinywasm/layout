//go:build !wasm

package rightpanel

import (
	"github.com/tinywasm/css"
)

// SSRInstance returns a new instance of RightPanel for SSR asset collection.
func SSRInstance() *RightPanel { return &RightPanel{} }

// RenderCSS implements dom.CSSProvider.
func (r *RightPanel) RenderCSS() *css.Stylesheet {
	return css.New(
		css.Root(
			css.Declare(TokenTitleHeight, "8vh"),
			css.Declare(TokenContentHeight, "89vh"),
			css.Declare(TokenControlsHeight, "3vh"),
			css.Declare(TokenMainWidth, "66vw"),
			css.Declare(TokenAsideWidth, "30vw"),
			css.Declare(TokenGap, css.Space2.Var()),
			css.Declare(TokenBorderColor, css.ColorMuted.Var()),
			css.Declare(TokenBg, css.ColorSurface.Var()),
			css.Declare(TokenAsideBg, css.ColorOnSurface.Var()),
			css.Declare(TokenTitleColor, css.ColorSecondary.Var()),
		),

		css.Rule(ClsWrapper,
			css.Display(css.Flex_),
			css.FlexDirection(css.Str("row")),
			css.Width(css.Pct(100)),
			css.Height(TokenContentHeight),
			css.RawRule("overflow: hidden;"),
		),

		css.Rule(ClsMain,
			css.Display(css.Grid),
			css.RawRule("grid-template-rows: auto 1fr;"),
			css.Width(TokenMainWidth),
			css.Height(css.Pct(100)),
			css.RawRule("overflow: hidden;"),
			css.RawRule("border-right: 0.1vw solid "+TokenBorderColor.Var()+";"),
		),

		css.Rule(ClsHeader,
			css.Display(css.Flex_),
			css.FlexDirection(css.Str("column")),
			css.Background(TokenBg),
			css.Padding(css.Space1, css.Space2),
		),

		css.Rule(ClsTitleRow,
			css.Display(css.Flex_),
			css.FlexDirection(css.Str("row")),
			css.AlignItems(css.Center),
			css.Gap(TokenGap),
			css.MinHeight(TokenTitleHeight),
		),

		css.Rule(css.Selector("."+string(ClsTitleRow)+" h1"),
			css.FontSize(css.TextXl),
			css.Color(TokenTitleColor),
			css.Margin(css.Zero),
		),

		css.Rule(ClsHeadControls,
			css.Display(css.Flex_),
			css.FlexDirection(css.Str("row")),
			css.AlignItems(css.Center),
			css.MinHeight(TokenControlsHeight),
			css.RawRule("padding-bottom: "+css.Space1.Var()+";"),
		),

		css.Rule(ClsArticle,
			css.RawRule("overflow-y: auto;"),
			css.Padding(css.Space2),
			css.Background(css.ColorSurface),
			css.BorderRadius(css.RadiusMd, css.RadiusMd, css.Zero, css.Zero),
		),

		css.Rule(css.Selector("."+string(ClsArticle)+"::-webkit-scrollbar"),
			css.Width(css.Str("0.2em")),
			css.Background(css.None),
		),

		css.Rule(css.Selector("."+string(ClsArticle)+"::-webkit-scrollbar-thumb"),
			css.Background(css.ColorMuted),
			css.BorderRadius(css.RadiusSm),
		),

		css.Rule(ClsAside,
			css.Display(css.Grid),
			css.RawRule("grid-template-rows: auto 1fr;"),
			css.Width(TokenAsideWidth),
			css.Height(css.Pct(100)),
			css.RawRule("overflow: hidden;"),
		),

		css.Rule(ClsAsideHeader,
			css.Display(css.Flex_),
			css.FlexDirection(css.Str("row")),
			css.AlignItems(css.Center),
			css.MinHeight(TokenControlsHeight),
			css.Padding(css.Space1, css.Space2),
			css.Background(TokenAsideBg),
		),

		css.Rule(ClsAsideContent,
			css.RawRule("overflow-y: auto;"),
			css.Padding(css.Space2),
			css.Background(TokenAsideBg),
		),

		css.Rule(css.Selector("."+string(ClsAsideContent)+"::-webkit-scrollbar"),
			css.Width(css.Str("0.2em")),
			css.Background(css.None),
		),

		css.Rule(css.Selector("."+string(ClsAsideContent)+"::-webkit-scrollbar-thumb"),
			css.Background(css.ColorMuted),
			css.BorderRadius(css.RadiusSm),
		),

		css.Media("(max-width: 640px)",
			css.Rule(ClsWrapper,
				css.FlexDirection(css.Str("column")),
				css.Height(css.Auto),
			),
			css.Rule(ClsMain,
				css.Width(css.Pct(100)),
				css.Height(css.Auto),
			),
			css.Rule(ClsAside,
				css.Width(css.Pct(100)),
				css.Height(css.Auto),
			),
			css.Rule(ClsArticle,
				css.RawRule("overflow-y: visible;"),
			),
			css.Rule(ClsAsideContent,
				css.RawRule("overflow-y: visible;"),
			),
		),
	)
}
