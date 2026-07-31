//go:build !wasm

package rightpanel

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

// RenderSheet returns the style Sheet containing the rules for rightpanel.
func (r *RightPanel) RenderSheet() *style.Sheet {
	return style.For(r).
		// EdgeToEdge: the panel is welded to the application frame on its right
		// and bottom; the curve As(Panel) would add leaves corner slivers of
		// the background behind it.
		Root(
			style.Stack(style.SpaceNone),
			style.Fill(),
			style.As(style.Panel),
			style.EdgeToEdge(),
		).
		Part(widget.Part("main"),
			style.Split(style.SplitTwoThirds, style.Space2),
			style.Fill(),
		).
		// EdgeToEdge: the header runs the full width of the panel and touches
		// the frame's right edge; the radius As(Panel) would add opens a sliver
		// against the window border.
		Part(widget.Part("header"),
			style.Stack(style.Space1),
			style.Pad(style.Space2),
			style.As(style.Panel),
			style.EdgeToEdge(),
		).
		Part(widget.Part("title-row"),
			style.Row(style.Space2),
			style.Center(),
		).
		// EdgeToEdge: the title chip rides against the header's right edge,
		// where As(Secondary)'s default 4px corner shows the header behind it.
		Part(widget.Part("title"),
			style.FontSize(style.TextXl),
			style.As(style.Secondary),
			style.EdgeToEdge(),
		).
		Part(widget.Part("controls"),
			style.Row(style.Space1),
		).
		Part(widget.Part("article"),
			style.As(style.Page),
			style.Pad(style.Space2),
			style.Scroll(),
			style.Fill(),
		).
		Part(widget.Part("aside"),
			style.As(style.Panel),
			style.Stack(style.SpaceNone),
			style.Fill(),
		).
		Part(widget.Part("aside-header"),
			style.Row(style.Space2),
			style.Pad(style.Space1),
			style.As(style.Panel),
		).
		Part(widget.Part("aside-content"),
			style.As(style.Panel),
			style.Pad(style.Space2),
			style.Scroll(),
			style.Fill(),
		)
}

// RenderCSS implements visual contract for rightpanel layout.
func (r *RightPanel) RenderCSS() *css.Stylesheet {
	return r.RenderSheet().Stylesheet()
}
