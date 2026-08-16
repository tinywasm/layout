//go:build !wasm

package landing

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/widget/style"
)

// RenderSheet returns the style Sheet containing the rules for landing.
func (p *Page) RenderSheet() *style.Sheet {
	return style.For(p).
		Root(
			style.Stack(style.SpaceNone),
			style.Fill(),
			style.As(style.Secondary),
		).
		Part(partHeader,
			style.Stack(style.SpaceNone),
			style.KeepSize(),
		).
		// The band: one padded, stacked strip per section. Everything below it
		// arranges content inside that strip and never re-pads it.
		Part(partSection,
			style.Stack(style.Space4),
			style.Pad(style.Space6),
		).
		Part(partFooter,
			style.Stack(style.Space2),
			style.Pad(style.Space6),
			style.As(style.Primary),
		).
		Part(partSplit,
			style.Grid(style.ColumnMedium, style.Space4),
		).
		Part(partCards,
			style.Stack(style.Space4),
		).
		Part(partForm,
			style.Stack(style.Space4),
		).
		Part(partHours,
			style.Stack(style.Space4),
		).
		Part(partMap,
			style.Stack(style.Space4),
		).
		Part(partTitle,
			style.FontSize(style.Text2xl),
			style.FontWeight(style.WeightBold),
		).
		Part(partSubtitle,
			style.FontSize(style.TextLg),
			style.Glyph(style.Subtle),
		).
		Part(partBody,
			style.Stack(style.Space2),
		).
		Part(partMedia,
			style.MediaBox(style.Aspect16x9),
		).
		Part(partGrid,
			style.Grid(style.ColumnMedium, style.Space4),
		).
		Part(partBadge,
			style.FontSize(style.TextSm),
			style.Glyph(style.Subtle),
		).
		Part(partBrand,
			style.FontWeight(style.WeightBold),
		)
}

// RenderCSS implements visual contract for landing layout.
func (p *Page) RenderCSS() *css.Stylesheet {
	return p.RenderSheet().Stylesheet()
}
