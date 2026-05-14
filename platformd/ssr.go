//go:build !wasm

package platformd

import (
	. "github.com/tinywasm/css"
)

// SSRInstance returns a new instance of Platform for SSR asset collection.
func SSRInstance() *Platform { return &Platform{} }

// RenderCSS implements CSSProvider.
func (p *Platform) RenderCSS() *Stylesheet {
	return NewStylesheet(
		Root(
			Declare(tokenFontSizeNormal, "1.1rem"),
			Declare(tokenFontSizeSmall, ".6rem"),
			Declare(tokenColorPrimary, "#ffffff"),
			Declare(tokenColorSecondary, "#3f88bf"),
			Declare(tokenColorTertiary, "#c2c1c1"),
			Declare(tokenColorQuaternary, "#000000"),
			Declare(tokenColorGray, "#e9e9e9"),
			Declare(tokenColorSelection, "#ff9300"),
			Declare(tokenColorHover, "#ff95008e"),
			Declare(tokenColorSuccess, "#aadaff7c"),
			Declare(tokenColorError, "#f20707"),
			Declare(tokenMenuSize, "6vh"),
			Declare(tokenHeaderHeight, "5vh"),
			Declare(tokenContentHeight, "94vh"),
			Declare(tokenContentWidth, "100vw"),
			Declare(tokenSlideDur, "0.6s"),
			Declare(tokenTransitionWait, "0s"),
		),

		// Universal reset for platform elements
		Rule(Selector("."+string(clsRoot)+" *"),
			RawRule("-webkit-box-sizing: border-box;"), // TODO(css-dsl): add typed -webkit-box-sizing
			RawRule("-moz-box-sizing: border-box;"),    // TODO(css-dsl): add typed -moz-box-sizing
			Margin(Zero),
			Padding(Zero),
			BoxSizing(Str("border-box")),
			Outline(None),
			TextDecoration(None),
			ListStyle(None),
			ListStyleType(None),
			RawRule("-webkit-user-select: none;"), // TODO(css-dsl): add typed -webkit-user-select
			RawRule("-khtml-user-select: none;"),  // TODO(css-dsl): add typed -khtml-user-select
			RawRule("-moz-user-select: none;"),    // TODO(css-dsl): add typed -moz-user-select
			RawRule("-ms-user-select: none;"),     // TODO(css-dsl): add typed -ms-user-select
			UserSelect(None),
			FontFamily(Str("Arvo")),
		),

		/************ MOBILE (Default) ************/
		Rule(clsRoot,
			FontSize(Rem(0.8)),
			Background(tokenColorTertiary),
			MinHeight(Vh(100)),
			Display(Flex_),
			AlignItems(Center),
			JustifyContent(Center),
			Height(Vh(100)),
		),

		Rule(clsHeader,
			Position(Fixed),
			AlignSelf(FlexEnd),
			Bottom(Px(5)),
			MarginLeft(Px(5)),
		),

		Rule(Selector("."+string(clsRoot)+" a:link"), Color(tokenColorQuaternary)),
		Rule(Selector("."+string(clsRoot)+" a:visited"), Color(tokenColorQuaternary)),

		Rule(clsMenu,
			Background(tokenColorGray),
			Position(Fixed),
			Top(Zero),
			Width(Vw(100)),
			Height(tokenMenuSize),
			Transition(tokenSlideDur, Str("all"), Str("ease")),
			ZIndex(Str("3")),
		),

		Rule(clsNavbar,
			Display(Flex_),
			FlexDirection(Row),
			AlignItems(Center),
			JustifyContent(SpaceAround),
			Height(Pct(100)),
		),

		Rule(clsNavItem,
			Width(Pct(100)),
		),

		Rule(clsNavActive,
			Background(tokenColorSelection),
		),

		Rule(Selector("."+string(clsNavActive)+", ."+string(clsNavActive)+" > svg"),
			Color(tokenColorPrimary),
			Transition(Str("0.3s"), Str("all"), Str("ease")),
		),

		Rule(clsNavLink,
			Display(Flex_),
			FlexDirection(Column),
			AlignItems(Center),
			JustifyContent(Center),
			Height(tokenMenuSize),
			Transition(tokenSlideDur, Str("all"), Str("ease")),
		),

		Rule(clsNavIcon,
			Color(tokenColorSecondary),
			Height(Em(8)),
			Width(Em(8)),
			Transition(tokenSlideDur, Str("all"), Str("ease")),
		),

		Rule(clsLinkText,
			Display(None),
		),

		Rule(clsPanel,
			Top(Vh(-100)),
			Margin(Zero),
			ZIndex(Str("1")),
			Width(tokenContentWidth),
			Height(tokenContentHeight),
			Position(Absolute),
			Background(Hex("#000")),
			Transition(tokenSlideDur, Str("all"), Str("ease-in-out")),
		),

		Rule(clsPanelActive,
			Top(tokenMenuSize),
			Background(tokenColorSecondary),
		),

		Rule(clsMsgDesktop,
			Width(Pct(100)),
			JustifyContent(Center),
			Display(Flex_),
			AlignItems(Center),
		),

		Rule(clsArea,
			TextTransform(Uppercase),
			Color(tokenColorPrimary),
			TextAlign(Center),
			Display(Flex_),
			AlignItems(Center),
		),

		Rule(clsUserBlock,
			Display(Flex_),
			AlignItems(Center),
		),

		Rule(clsOrientationWarn,
			ZIndex(Str("4")),
			Position(Fixed),
			Top(Zero),
			Left(Zero),
			Width(Pct(100)),
			Height(Pct(100)),
			JustifyContent(Center),
			AlignItems(Center),
			Display(None),
			Color(tokenColorPrimary),
			TouchAction(None),
			Background(tokenColorSecondary),
		),

		Rule(clsMsgMobile,
			ZIndex(Str("4")),
			Position(Fixed),
			Top(Zero),
			Left(Zero),
			Width(Pct(100)),
			Height(Pct(100)),
			JustifyContent(Center),
			AlignItems(Center),
			Display(Flex_),
			Background(Str("rgba(255, 255, 255, 0.8)")),
			Visibility(Visible),
		),

		Rule(clsMsg,
			TextAlign(Center),
			Padding(Px(20)),
			Width(Pct(80)),
			FontSize(Em(1.3)),
			TextShadow(Em(0.1), Em(0.1), Em(0.1), Hex("#ffffff")),
			Animation(Str("pdFadeIn"), tokenTransitionWait),
		),

		// Message variants
		Rule(Selector(".pd-msg-info"), Color(tokenColorTertiary)),
		Rule(Selector(".pd-msg-error"), Color(tokenColorError)),
		Rule(Selector(".pd-msg-success"), Color(tokenColorSecondary)),
		Rule(Selector(".pd-msg-warning"), Color(tokenColorSelection)),

		Keyframes("pdFadeOut",
			At("from", Opacity(1.0), Visibility(Visible)),
			At("to", Opacity(0.7), Visibility(Hidden)),
		),

		Keyframes("pdFadeIn",
			At("from", Opacity(0.0), Visibility(Hidden)),
			At("to", Opacity(1.0), Visibility(Visible)),
		),

		/************ DESKTOP ************/
		MediaDesktop(
			Root(
				Declare(tokenHeaderHeight, "5vh"),
				Declare(tokenContentHeight, "95vh"),
				Declare(tokenMenuSize, "5vw"),
				Declare(tokenContentWidth, "95vw"),
			),

			Rule(clsRoot,
				Overflow(Hidden),
				Background(tokenColorGray),
				FontSize(Px(16)),
				Height(tokenContentHeight),
				Display(Grid),
				GridTemplate(Str(`"header         header" `+tokenHeaderHeight.Var()+`
					"module-content menu-container" `+tokenContentHeight.Var()+` /
					`+tokenContentWidth.Var()+` `+tokenMenuSize.Var())),
			),

			Rule(clsHeader,
				AlignSelf(Unset),
				Bottom(Auto),
				Position(Unset),
				MarginLeft(Zero),
				MaxHeight(tokenHeaderHeight),
				Width(Vw(100)),
				GridArea(Str("header")),
				Display(Grid),
				GridTemplateColumns(Str("1fr 3fr 1fr")),
				Background(tokenColorGray),
			),

			Rule(clsMenu,
				Top(tokenHeaderHeight),
				Width(tokenMenuSize),
				Height(tokenContentHeight),
				MarginLeft(Auto),
				Right(Zero),
				Background(tokenColorGray),
			),

			Rule(clsMenu.Hover(),
				Width(Em(10)),
			),

			Rule(Selector("."+string(clsMenu)+":hover ."+string(clsLinkText)),
				Display(Block),
			),

			Rule(clsNavbar,
				FlexDirection(Column),
			),

			Rule(clsNavItem,
				Height(Pct(100)),
			),

			Rule(clsNavLink,
				FlexDirection(Row),
				Height(Pct(100)),
			),

			Rule(Selector("."+string(clsNavLink)+" svg"),
				Margin(Zero, Em(0.5)),
			),

			Rule(Selector("."+string(clsNavLink)+":hover"),
				Background(tokenColorHover),
			),

			Rule(Selector("."+string(clsNavActive)+", ."+string(clsNavActive)+" > ."+string(clsLinkText)),
				Color(tokenColorPrimary),
			),

			Rule(clsLinkText,
				Color(tokenColorSecondary),
			),

			Rule(clsNavIcon,
				MaxWidth(Em(2.5)),
				MinWidth(Em(2.5)),
			),

			Rule(clsPanel,
				GridArea(Str("module-content")),
				Position(Unset),
				Top(tokenHeaderHeight),
				MarginLeft(Vw(-100)),
				BorderRadius(Zero),
			),

			Rule(clsPanelActive,
				MarginLeft(Zero),
			),

			Rule(clsMsg,
				Padding(Zero),
				Animation(Str("pdFadeIn"), tokenTransitionWait),
				Visibility(Visible),
				Width(Pct(100)),
			),

			Rule(clsUserBlock,
				FontSize(Calc(".5em + .5vh")),
				MarginLeft(Rem(0.4)),
			),

			Rule(clsArea,
				FontSize(Calc(".5em + .5vh")),
				MarginLeft(Auto),
				MarginRight(Rem(0.4)),
				TextAlign(RightText),
			),

			Rule(clsMsgDesktop,
				FontSize(Calc(".4em + .4vh")),
			),

			Rule(clsMsgMobile,
				All(Initial),
				Display(None),
			),
		),

		Media("(orientation: portrait)",
			Rule(clsOrientationWarn, Display(None)),
		),

		Media("(orientation: landscape) and (min-width: 600px) and (max-width: 1024px)",
			Rule(clsOrientationWarn, Display(Flex_)),
		),
	)
}
