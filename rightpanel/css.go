//go:build !wasm

package rightpanel

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

// RenderCSS implements visual contract for rightpanel layout.
func (r *RightPanel) RenderCSS() *css.Stylesheet {
	return style.For(r).
		Root(
			style.Stack(style.SpaceNone),
			style.Fill(),
			style.As(style.Panel),
		).
		Part(widget.Part("main"),
			style.Split(style.SplitTwoThirds, style.Space2),
			style.Fill(),
		).
		Part(widget.Part("header"),
			style.Stack(style.Space1),
			style.Pad(style.Space2),
			style.As(style.Panel),
		).
		Part(widget.Part("title-row"),
			style.Row(style.Space2),
			style.Center(),
		).
		Part(widget.Part("title"),
			style.FontSize(style.TextXl),
			style.As(style.Secondary),
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
		).
		Stylesheet()
}
