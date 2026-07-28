//go:build !wasm

package rightpanel

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

// RenderCSS implements visual contract for rightpanel layout.
func (r *RightPanel) RenderCSS() *css.Stylesheet {
	return style.Of(NameRightPanel).
		Root(
			style.Stack(style.Space0),
			style.Fill(),
			style.On(style.Panel),
		).
		Part(widget.Part("main"),
			style.Split(style.RatioTwoThirds, style.Space2),
			style.Fill(),
		).
		Part(widget.Part("header"),
			style.Stack(style.Space1),
			style.Pad(style.Space2),
			style.On(style.Panel),
		).
		Part(widget.Part("title-row"),
			style.Row(style.Space2),
			style.Center(),
		).
		Part(widget.Part("title"),
			style.Text(style.TextXl),
			style.On(style.Secondary),
		).
		Part(widget.Part("controls"),
			style.Row(style.Space1),
		).
		Part(widget.Part("article"),
			style.On(style.Page),
			style.Pad(style.Space2),
			style.Scrolls(),
			style.Fill(),
		).
		Part(widget.Part("aside"),
			style.On(style.Panel),
			style.Stack(style.Space0),
			style.Fill(),
		).
		Part(widget.Part("aside-header"),
			style.Row(style.Space2),
			style.Pad(style.Space1),
			style.On(style.Panel),
		).
		Part(widget.Part("aside-content"),
			style.On(style.Panel),
			style.Pad(style.Space2),
			style.Scrolls(),
			style.Fill(),
		).
		Stylesheet()
}
