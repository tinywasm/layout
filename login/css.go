//go:build !wasm

package login

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/widget/style"
)

// RenderCSS defines the login screen's visual contract using the style DSL.
func (l *Login) RenderCSS() *css.Stylesheet {
	return style.For(l).
		// Cover, not Fill+FillCentered: Fill's height:100% only resolves
		// against an ancestor with an explicit height, and nothing between
		// this root and <body> declares one (a host's own SPA root included)
		// — the card rendered shrink-wrapped in the top-left corner instead
		// of centered, height:100% resolving to 0. Cover is "the outermost
		// frame of an application shell" for exactly this reason: 100dvh is
		// definite regardless of ancestor chain. CenterContent is a separate
		// field, not another flowType, so it layers on top of Cover's own
		// flex-column without the conflict Stack+MediaBox or Fill+FillCentered
		// would have. Anchor makes this the positioned ancestor PartMark
		// docks against, so the mark stays pinned to THIS screen's corner
		// even if a host renders Login somewhere other than the document root.
		//
		// Primary is the whole point of the screen: see the type's own doc.
		// It also does the work no shadow can — a card only reads as elevated
		// against something it is clearly ON, and Page-on-Page left the form
		// looking like loose markup rather than a front door.
		Root(
			style.Cover(),
			style.CenterContent(),
			style.Anchor(),
			style.As(style.Primary),
			style.Pad(style.Space4),
		).
		// Inset, not Panel or Page: fieldset paints its input As(Page) — the
		// whitest surface the theme has — on the stated assumption that the
		// form sits in a SUNKEN card, which is what makes that white read as
		// "write here" without a border announcing it. A Panel card is three
		// percent off the input's own white and the fields dissolve into it;
		// Inset is the surface crudview already puts its own field stack on,
		// so a login form and a module form come out of the same box.
		//
		// Space6 twice over: the gap between header and form is the same air
		// as the card's own inset, which is what keeps a card this small from
		// reading as cramped. Compact caps it at a single column of controls
		// — Readable's 65ch is a measure for prose and made the card wider
		// than most of the screens it fronts.
		Part(PartCard,
			style.Stack(style.Space6),
			style.Width(style.Compact),
			style.As(style.Inset),
			style.Round(style.RadiusLg),
			style.Raise(style.Floating),
			style.Pad(style.Space6),
		).
		// Space1, against the card's Space6: title and subtitle are one block
		// that happens to be set in two sizes, and spacing them like siblings
		// of the form would read as three unrelated things stacked up.
		Part(PartHeader,
			style.Stack(style.Space1),
			style.CenterContent(),
		).
		// TextXl, not Text2xl: 2rem is a page banner, and set inside a card
		// this narrow it wrapped the app's own name. The title leads the card;
		// it does not have to dominate the viewport to do that.
		Part(PartTitle,
			style.FontSize(style.TextXl),
			style.FontWeight(style.WeightBold),
		).
		// Glyph(Subtle), not a second surface: the subtitle is muted TEXT on
		// the card it already sits on — painting it a fill would make it a
		// banner competing with the title it is supposed to serve.
		Part(PartSubtitle,
			style.FontSize(style.TextSm),
			style.Glyph(style.Subtle),
		).
		// MediaBox, not IconBox: IconBox sizes an inline <svg> glyph off the
		// font metrics it recolors with currentColor, wrong for an <img> —
		// an arbitrary crest or seal keeps its own colors and needs its own
		// box. Docked pins it to the viewport corner independent of the
		// card's own height, matching the legacy reference this screen
		// productionizes: the crest sits bottom-left regardless of how
		// tall the form above grows.
		Part(PartMark,
			style.MediaBox(style.AspectSquare),
			style.Docked(style.Viewport, style.EdgeBottom, style.SideStart, style.Space4),
		).
		Stylesheet()
}
