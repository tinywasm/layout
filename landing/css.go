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
		Part(partSection,
			style.Stack(style.Space4),
			style.Pad(style.Space6),
		).
		Part(partSplit,
			style.Grid(style.ColumnMedium, style.Space4),
			style.Pad(style.Space6),
		).
		Part(partCards,
			style.Stack(style.Space4),
			style.Pad(style.Space6),
		).
		Part(partStats,
			style.Stack(style.Space4),
			style.Pad(style.Space6),
		).
		Part(partForm,
			style.Stack(style.Space4),
			style.Pad(style.Space6),
		).
		Part(partHours,
			style.Stack(style.Space4),
			style.Pad(style.Space6),
		).
		Part(partMap,
			style.Stack(style.Space4),
			style.Pad(style.Space6),
		).
		Part(partFooter,
			style.Stack(style.Space2),
			style.Pad(style.Space6),
			style.As(style.Primary),
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
		)
}

// RenderCSS implements visual contract for landing layout.
func (p *Page) RenderCSS() *css.Stylesheet {
	return p.RenderSheet().Stylesheet()
}
