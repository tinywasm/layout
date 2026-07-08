//go:build !wasm

package crudview

import (
	. "github.com/tinywasm/css"
	"github.com/tinywasm/svg"
)

func GenerateCSS() *Stylesheet {
	s := NewStylesheet(
		Rule(clsModuleContent,
			Display(Grid),
			Border(Str(".1vw solid var(--color-outline-variant)")),
			BorderRadius(Str("0 .4em .4em 0")),
			RawRule("grid-template: none / auto 1fr"),
		),

		Rule(clsArticleContend,
			Flex(None),
			Display(Grid),
			RawRule("grid-template: 'title' var(--cv-title-height, 8vh) 'article' 80vh 'controls' var(--cv-controls-height, 9vh) / 66vw"),
		),

		Rule(clsArticleContendFullPage,
			Display(Grid),
			RawRule("grid-template: 'title' var(--cv-title-height, 8vh) 'article' 89vh / 96vw"),
		),

		Rule(clsAsideContend,
			Flex(None),
			MarginLeft(Str("var(--cv-mag-pri, .5rem)")),
			Display(Grid),
			RawRule("grid-template: 'aside-list' var(--cv-title-height, 8vh) 'aside-list' 80vh 'aside-search' var(--cv-controls-height, 9vh) / 29vw"),
		),

		Rule(clsTitleContainer,
			GridArea(Str("title")),
			Display(Flex_),
			AlignItems(Center),
			Background(Str("var(--color-primary)")),
		),

		Rule("."+string(clsTitle)+" h1",
			MarginLeft(Str(".7em")),
			Color(Str("var(--color-on-primary)")),
			FontSize(Str("1.5rem")),
		),

		Rule("article",
			GridArea(Str("article")),
			Background(Str("var(--color-surface)")),
			BorderRadius(Str(".4em .4em 0 0")),
			Display(Flex_),
			FlexDirection(Column),
			MarginLeft(Str("var(--cv-mag-cua, .2rem)")),
		),

		Rule(clsBoxContent,
			Decl{Prop: "flex-grow", Val: "1"}, // TODO(tinywasm/css): swap for FlexGrow(...) once css/docs/PLAN.md ships
			Display(Flex_),
			MinHeight(Str("0")),
			Background(Str("var(--color-surface-variant)")),
			BorderRadius(Str(".4em")),
			Padding(Str("var(--cv-mag-pri, .5rem) 0 var(--cv-mag-pri, .5rem) var(--cv-mag-pri, .5rem)")),
		),

		Rule(clsControls,
			GridArea(Str("controls")),
			Decl{Prop: "margin-bottom", Val: "var(--cv-mag-pri, .5rem)"}, // TODO(tinywasm/css): swap for MarginBottom(...) once css/docs/PLAN.md ships
		),

		Rule(clsContebuton,
			Display(Flex_),
			Height(Str("100%")),
			Background(Str("var(--color-surface)")),
			JustifyContent(Str("space-between")),
			Padding(Str(".1em")),
			BorderRadius(Str("0 0 .4em .4em")),
		),

		Rule(clsBtnCrud,
			Flex(Auto),
			Margin(Str(".25em")),
			Padding(Str(".4rem")),
			Background(Str("var(--color-primary)")),
			Color(Str("var(--color-on-primary)")),
			BorderRadius(Str(".4em")),
			Border(Str("none")),
			Cursor(Pointer),
			Display(Flex_),
			AlignItems(Center),
			JustifyContent(Center),
		),

		Rule("."+string(clsBtnCrud)+":disabled",
			Background(Str("var(--color-outline-variant)")),
			Color(Str("var(--color-on-surface-variant)")),
			Cursor(Str("not-allowed")),
			Opacity(0.6),
		),

		Rule(clsAsideList,
			GridArea(Str("aside-list")),
			Decl{Prop: "margin-top", Val: "var(--cv-mag-pri, .5rem)"}, // TODO(tinywasm/css): swap for MarginTop(...) once css/docs/PLAN.md ships
			Background(Str("var(--color-surface)")),
			BorderRadius(Str(".4em .4em 0 0")),
			Display(Flex_),
			FlexDirection(Column),
			Decl{Prop: "padding-top", Val: "var(--cv-mag-sec, .2rem)"}, // TODO(tinywasm/css): swap for PaddingTop(...) once css/docs/PLAN.md ships
		),

		Rule(clsListaBox,
			Display(Flex_),
			Height(Str("100%")),
			FlexDirection(Column),
			Background(Str("var(--color-surface-variant)")),
			BorderRadius(Str(".4em")),
			Padding(Str(".2rem")),
			Overflow(Hidden),
		),

		Rule(clsLista,
			Decl{Prop: "flex-grow", Val: "1"}, // TODO(tinywasm/css): swap for FlexGrow(...) once css/docs/PLAN.md ships
			Overflow(Auto),
			MinHeight(Str("0")),
			Padding(Str(".2em")),
			ListStyle(None),
		),

		Rule(clsTargetLi,
			Decl{Prop: "position", Val: "relative"}, // TODO(tinywasm/css): swap for Position(Relative) once css/docs/PLAN.md ships
			Display(Flex_),
			AlignItems(Center),
			MinHeight(Str("60px")),
			MaxHeight(Str("60px")),
			Width(Str("100%")),
			Padding(Str(".3em")),
			Decl{Prop: "margin-bottom", Val: "1.2em"}, // TODO(tinywasm/css): swap for MarginBottom(...) once css/docs/PLAN.md ships
			Cursor(Pointer),
			Border(Str("2px solid var(--color-outline-variant)")),
			BorderRadius(Str(".3em")),
			Background(Str("var(--color-surface)")),
			Transition(Str(".3s all ease")),
		),

		Rule(clsTargetLiOn,
			BorderColor(Str("var(--color-primary)")),
			Background(Str("var(--color-primary-container)")),
		),

		Rule(clsDescriptionTarget,
			Decl{Prop: "position", Val: "absolute"}, // TODO(tinywasm/css): swap for Position(Absolute) once css/docs/PLAN.md ships
			Right(Px(5)),
			Bottom(Px(-10)),
			Background(Str("var(--color-secondary-container)")),
			Color(Str("var(--color-on-secondary-container)")),
			FontSize(Pct(80)),
			BorderRadius(Str(".4em")),
			Padding(Str(".3em")),
		),

		Rule(clsAsideSearch,
			GridArea(Str("aside-search")),
			Display(Flex_),
			Background(Str("var(--color-surface)")),
			Padding(Str(".1em")),
			BorderRadius(Str("0 0 .4em .4em")),
			Decl{Prop: "margin-bottom", Val: "var(--cv-mag-pri, .5rem)"}, // TODO(tinywasm/css): swap for MarginBottom(...) once css/docs/PLAN.md ships
			AlignItems(Center),
		),

		Rule("."+string(clsAsideSearch)+" label",
			Background(Str("var(--color-primary)")),
			Color(Str("var(--color-on-primary)")),
			Padding(Str(".3em")),
			BorderRadius(Str(".3em 0 0 .3em")),
			Display(Flex_),
		),

		Rule("."+string(clsAsideSearch)+" input",
			Decl{Prop: "flex-grow", Val: "1"}, // TODO(tinywasm/css): swap for FlexGrow(...) once css/docs/PLAN.md ships
			Border(Str("1px solid var(--color-outline-variant)")),
			Padding(Str(".3em")),
			BorderRadius(Str("0 .3em .3em 0")),
		),
	)

	return s
}

func IconSvg() *svg.Sprite {
	return svg.NewSprite(
		svg.Define("icon-crud-new", "0 0 16 16", svg.Path("M8 1v14M1 8h14")),
		svg.Define("icon-crud-del", "0 0 16 16", svg.Path("M1 8h14")),
		svg.Define("icon-crud-cancel", "0 0 16 16", svg.Path("M4 4h-4v-4M0 4c2-4 8-4 10-2s4 8 2 10-8 4-10 2")),
		svg.Define("icon-crud-save", "0 0 16 16", svg.Path("M2 1h10l3 3v11h-13zM10 1v4h-7v-4M4 15v-5h8v5")),
		svg.Define("icon-crud-search", "0 0 16 16", svg.Path("M11 11l4 4M1 7a6 6 0 1 0 12 0a6 6 0 1 0-12 0")),
	)
}
