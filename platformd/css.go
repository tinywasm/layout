//go:build !wasm

package platformd

import (
	. "github.com/tinywasm/css"
)

// RenderCSS implements CSSProvider.
func (p *Platform) RenderCSS() *Stylesheet {
	return NewStylesheet(
		Root(
			Declare(tokenFontSizeNormal, "1.1rem"),
			Declare(tokenFontSizeSmall, ".6rem"),
			Declare(tokenMenuSize, "4vw"),
			Declare(tokenHeaderHeight, "3vh"),
			Declare(tokenContentHeight, "97vh"),
			Declare(tokenContentWidth, "96vw"),
			Declare(tokenSlideDur, "0.6s"),
			Declare(tokenTransitionWait, "0s"),
		),

		// Universal reset. Selector is a plain `*` (specificity 0,0,0), NOT
		// `.pd-root *` (0,1,0): the scoped form ties with every component's
		// single-class rule and, being later in the bundle, silently zeroed their
		// padding/margin. A bare `*` lets any component class override the reset.
		Rule(Selector("*"),
			RawRule("-webkit-box-sizing: border-box;"),
			RawRule("-moz-box-sizing: border-box;"),
			Margin(Zero),
			Padding(Zero),
			BoxSizing(Str("border-box")),
			Outline(None),
			TextDecoration(None),
			ListStyle(None),
			ListStyleType(None),
			RawRule("-webkit-user-select: none;"),
			RawRule("-khtml-user-select: none;"),
			RawRule("-moz-user-select: none;"),
			RawRule("-ms-user-select: none;"),
			UserSelect(None),
			FontFamily(Str("Arvo")),
		),

		/************ MOBILE (Default) ************/

		Rule(clsRoot,
			FontSize(Rem(0.8)),
			Background(ColorMuted),
			Width(Vw(100)),
			Height(Vh(100)),
			Position(Str("relative")),
			Overflow(Hidden),
		),

		Rule(clsHeader,
			Position(Fixed),
			AlignSelf(FlexEnd),
			Bottom(Px(5)),
			MarginLeft(Px(5)),
			ZIndex(Str("5")),
		),

		Rule(Selector("."+string(clsRoot)+" a:link"), Color(ColorOnSurface)),
		Rule(Selector("."+string(clsRoot)+" a:visited"), Color(ColorOnSurface)),

		// Hamburger button — visible only on mobile
		Rule(clsHamburger,
			Position(Fixed),
			Top(Px(8)),
			Right(Px(8)),
			ZIndex(Str("10")),
			Width(Px(44)),
			Height(Px(44)),
			Background(ColorSecondary),
			BorderRadius(Px(6)),
			Display(Flex_),
			FlexDirection(Column),
			AlignItems(Center),
			JustifyContent(SpaceAround),
			Padding(Px(10), Px(8)),
			Cursor(Pointer),
			Border(Str("none")),
		),

		Rule(Selector("."+string(clsHamburger)+" span"),
			Display(Block),
			Width(Pct(100)),
			Height(Px(2)),
			Background(ColorOnSecondary),
			BorderRadius(Px(2)),
		),

		// Nav drawer — hidden off-screen right using transform (no reflow, GPU)
		Rule(clsMenu,
			Background(ColorSurface),
			Position(Fixed),
			Top(Zero),
			Right(Zero),
			Width(Vw(75)),
			Height(Vh(100)),
			ZIndex(Str("9")),
			Transform(Str("translateX(100%)")),
		),

		// Nav overlay backdrop
		Rule(clsNavOverlay,
			Position(Fixed),
			Top(Zero),
			Left(Zero),
			Width(Pct(100)),
			Height(Pct(100)),
			Background(Str("rgba(0,0,0,0.45)")),
			ZIndex(Str("8")),
			Display(None),
		),

		// Open state — keyframe animation so it fires on new DOM elements too
		Rule(Selector("."+string(clsMenuOpen)+" ."+string(clsMenu)),
			Transform(Str("translateX(0)")),
			Animation(Str("pdDrawerOpen"), tokenSlideDur, Str("ease")),
		),
		Rule(Selector("."+string(clsMenuOpen)+" ."+string(clsNavOverlay)),
			Display(Block),
		),

		Rule(clsNavbar,
			Display(Flex_),
			FlexDirection(Column),
			AlignItems(Str("stretch")),
			JustifyContent(SpaceAround),
			Height(Pct(100)),
		),

		Rule(Selector("."+string(clsNavbar)+" li:first-child"),
			AlignSelf(Str("flex-start")),
		),

		Rule(Selector("."+string(clsNavbar)+" li:last-child"),
			AlignSelf(Str("flex-end")),
		),

		Rule(clsNavItem,
			Width(Pct(100)),
			Flex(Str("1")),
		),

		// Selection highlight — same --color-selection token as crudview list rows
		// so "selected" reads identically across menu and list.
		Rule(clsNavActive,
			Background(ColorSelection),
		),

		// The active nav ICON is always white: a dark glyph reads poorly on the
		// orange highlight (worse in light mode). This is intentionally NOT
		// --color-on-selection (which is dark in light mode for readable list text).
		Rule(Selector("."+string(clsNavActive)+", ."+string(clsNavActive)+" > svg"),
			Color(Str("#ffffff")),
			Transition(Str("0.3s"), Str("all"), Str("ease")),
		),

		Rule(clsNavLink,
			Display(Flex_),
			FlexDirection(Row),
			AlignItems(Center),
			Height(Pct(100)),
			Padding(Zero, Rem(1)),
			Transition(tokenSlideDur, Str("all"), Str("ease")),
		),

		Rule(ClsNavIcon,
			Color(ColorPrimary),
			Height(Em(2.5)),
			Width(Em(2.5)),
			MinWidth(Em(2.5)),
			Transition(tokenSlideDur, Str("all"), Str("ease")),
			MarginRight(Rem(0.8)),
		),

		Rule(clsLinkText,
			Color(ColorPrimary),
			FontSize(Em(1.1)),
		),

		Rule(clsPanel,
			Display(None),
		),

		Rule(clsPanelActive,
			Display(Block),
			Top(Zero),
			Left(Zero),
			Margin(Zero),
			ZIndex(Str("1")),
			Width(tokenContentWidth),
			Height(Vh(100)),
			Position(Absolute),
			Background(ColorSecondary),
			Animation(Str("pdSlideInTop"), tokenSlideDur, Str("ease-in-out")),
		),

		Rule(clsMsgDesktop,
			Width(Pct(100)),
			JustifyContent(Center),
			Display(Flex_),
			AlignItems(Center),
		),

		Rule(clsArea,
			TextTransform(Uppercase),
			Color(ColorOnSecondary),
			TextAlign(Center),
			Display(Flex_),
			AlignItems(Center),
		),

		Rule(clsUserBlock,
			Display(Flex_),
			AlignItems(Center),
		),

		Rule(clsHeaderRight,
			Display(Flex_),
			AlignItems(Center),
			JustifyContent(Str("flex-end")),
			RawRule("gap: .6rem"),
		),

		// Neutralize a fixed-positioned action (e.g. the theme toggle) so it sits
		// inline in the header instead of floating over the nav rail. !important
		// because the component's own `position: fixed` has equal specificity.
		Rule(Selector("."+string(clsHeaderRight)+" > *"),
			RawRule("position: static !important; top: auto; right: auto"),
		),

		// A header action rendered as a button (the theme toggle) shows as a bare
		// icon here — strip the component's floating-button chrome (circle/fill)
		// so it isn't an out-of-place blob next to the area name.
		Rule(Selector("."+string(clsHeaderRight)+" button"),
			Background(Str("transparent")),
			Border(None),
			RawRule("box-shadow: none"),
			RawRule("width: auto; height: auto"),
			Padding(Zero),
			FontSize(Rem(1.4)),
			Cursor(Pointer),
		),
		// Keep it a bare icon on hover too — the component's :hover re-adds a
		// filled circle + scale; a plain opacity change is the only affordance.
		Rule(Selector("."+string(clsHeaderRight)+" button:hover"),
			Background(Str("transparent")),
			RawRule("transform: none; box-shadow: none"),
			Opacity(0.7),
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
			Color(ColorOnSecondary),
			TouchAction(None),
			Background(ColorSecondary),
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
			TextShadow(Em(0.1), Em(0.1), Em(0.1), ColorOnSecondary),
			Animation(Str("pdFadeIn"), tokenTransitionWait),
		),

		// Message variants
		Rule(Selector(".pd-msg-info"), Color(ColorMuted)),
		Rule(Selector(".pd-msg-error"), Color(ColorError)),
		Rule(Selector(".pd-msg-success"), Color(ColorSecondary)),
		Rule(Selector(".pd-msg-warning"), Color(ColorSecondary)), // Selection was mapped to ColorSecondary

		Keyframes("pdFadeOut",
			At("from", Opacity(1.0), Visibility(Visible)),
			At("to", Opacity(0.7), Visibility(Hidden)),
		),

		Keyframes("pdFadeIn",
			At("from", Opacity(0.0), Visibility(Hidden)),
			At("to", Opacity(1.0), Visibility(Visible)),
		),

		Keyframes("pdSlideInTop",
			At("from", Top(Vh(-100))),
			At("to", Top(Zero)),
		),

		Keyframes("pdSlideInLeft",
			At("from", Left(Pct(-100))),
			At("to", Left(Zero)),
		),

		Keyframes("pdDrawerOpen",
			At("from", Transform(Str("translateX(100%)"))),
			At("to", Transform(Str("translateX(0)"))),
		),

		/************ DESKTOP ************/
		MediaDesktop(
			Root(
				// Fixed header tall enough for the user name + area + theme toggle
				// (a 3vh strip was too short for the 2.4rem toggle). Content fills
				// the remainder.
				Declare(tokenHeaderHeight, "2.8rem"),
				Declare(tokenContentHeight, "calc(100vh - 2.8rem)"),
				// Fixed em rail: 4vw collapsed below the icon width (2.5em) at
				// narrow widths, so icons overflowed and overlapped the content.
				// An em value always fits the icons regardless of viewport.
				Declare(tokenMenuSize, "3.6em"),
				Declare(tokenContentWidth, "96vw"),
			),

			Rule(clsRoot,
				Overflow(Hidden),
				Background(ColorSurface),
				FontSize(Px(16)),
				Height(Vh(100)),
				// Reset mobile flex centering — critical so grid items stretch to full height
				AlignItems(Str("normal")),
				JustifyContent(Str("normal")),
				Display(Grid),
				GridTemplate(Str(`"header         header" `+tokenHeaderHeight.Var()+`
					"module-content menu-container" `+tokenContentHeight.Var()+` /
					1fr `+tokenMenuSize.Var())),
			),

			// Hide hamburger and overlay on desktop
			Rule(clsHamburger, Display(None)),
			Rule(clsNavOverlay, Display(None)),

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
				Background(ColorSurface),
			),

			Rule(clsMenu,
				Position(Fixed),
				GridArea(Str("menu-container")),
				Top(tokenHeaderHeight),
				Right(Zero),
				Width(tokenMenuSize),
				Height(tokenContentHeight),
				MarginLeft(Auto),
				Background(ColorSurface),
				Transform(Str("none")),
				Animation(Str("none")),
				Transition(Str("0.3s"), Str("width"), Str("ease")),
			),

			Rule(clsMenu.Hover(),
				Width(Em(10)),
			),

			Rule(Selector("."+string(clsMenu)+":hover ."+string(clsLinkText)),
				Display(Block),
			),

			// When the rail expands on hover, left-align the link and give the icon
			// breathing room so the label is not flush against it.
			Rule(Selector("."+string(clsMenu)+":hover ."+string(clsNavLink)),
				JustifyContent(Str("flex-start")),
			),
			Rule(Selector("."+string(clsMenu)+":hover ."+string(ClsNavIcon)),
				MarginRight(Rem(0.8)),
			),

			Rule(clsNavbar,
				FlexDirection(Column),
				JustifyContent(SpaceAround),
				AlignItems(Str("stretch")),
			),

			// Each nav item takes equal share of the rail height — larger hit targets
			Rule(clsNavItem,
				Height(Auto),
				Flex(Str("1")),
			),

			Rule(clsNavLink,
				FlexDirection(Row),
				// Collapsed rail shows the icon only — keep it centered in the
				// narrow rail; hover switches to flex-start (see rule above).
				JustifyContent(Center),
				Height(Pct(100)),
				Padding(Zero, Em(0.5)),
			),

			Rule(Selector("."+string(clsNavLink)+":hover"),
				Background(ColorHover),
			),

			Rule(Selector("."+string(clsNavActive)+", ."+string(clsNavActive)+" > ."+string(clsLinkText)),
				Color(Str("#ffffff")),
			),

			Rule(clsLinkText,
				Color(ColorPrimary),
				Display(None),
			),

			Rule(ClsNavIcon,
				MaxWidth(Em(2.5)),
				MinWidth(Em(2.5)),
				MarginRight(Zero),
			),

			Rule(clsStage,
				GridArea(Str("module-content")),
				Position(Str("relative")),
				Height(Pct(100)),
				Overflow(Hidden),
			),

			Rule(clsPanel,
				Display(None),
			),

			Rule(clsPanelActive,
				Display(Block),
				Position(Absolute),
				Top(Zero),
				Left(Zero),
				Width(Pct(100)),
				Height(Pct(100)),
				Background(ColorSurface),
				// Square, not rounded: this panel's 4 corners meet the platform
				// frame (header/nav rail) directly — a curve here shows as a
				// mismatched notch against the module content's own rounded
				// corners (crudview's clsModuleContent), not a clean fit.
				BorderRadius(Zero),
				Animation(Str("pdSlideInLeft"), tokenSlideDur, Str("ease-in-out")),
			),

			Rule(clsMsg,
				Padding(Zero),
				Animation(Str("pdFadeIn"), tokenTransitionWait),
				Visibility(Visible),
				Width(Pct(100)),
			),

			Rule(clsUserBlock,
				FontSize(Rem(0.95)),
				FontWeight(Str("600")),
				Color(Str("var(--color-on-surface, #1c1c1e)")),
				MarginLeft(Rem(0.8)),
			),

			Rule(clsArea,
				FontSize(Rem(0.9)),
				Color(Str("var(--color-on-surface, #1c1c1e)")),
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

		Rule(Selector("."+string(clsOrientationWarn)+":empty"),
			Display(None),
		),

		// Hide msg-mobile when empty (no notifications) — prevents blank overlay blocking content
		Rule(Selector("."+string(clsMsgMobile)+":empty"),
			Display(None),
		),
	)
}
