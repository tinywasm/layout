//go:build !wasm

package crudview

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

// RenderCSS implements the visual contract for crudview using the style DSL.
func (v *CrudView) RenderCSS() *css.Stylesheet {
	return style.Of(NameCrudView).
		Root(
			style.Split(style.RatioTwoThirds, style.Space2),
			style.On(style.Accent),
			style.Flush(),
		).
		Part(widget.Part("detail"),
			style.Stack(style.Space2),
			style.Fill(),
		).
		Part(widget.Part("detail-full"),
			style.Stack(style.Space2),
			style.Fill(),
		).
		Part(widget.Part("fields"),
			style.On(style.Sunken),
			style.Pad(style.Space2),
			style.Scrolls(),
			style.Round(style.RadiusMd),
		).
		Part(widget.Part("aside"),
			style.Stack(style.Space1),
			style.On(style.Panel),
			style.Pad(style.Space1),
			style.Fill(),
		).
		Part(widget.Part("aside-content"),
			style.Fill(),
			style.Stack(style.Space0),
		).
		Part(widget.Part("list"),
			style.On(style.Sunken),
			style.Scrolls(),
			style.Round(style.RadiusMd),
		).
		Part(widget.Part("actions"),
			style.Row(style.Space1),
		).
		Part(widget.Part("title"),
			style.Row(style.Space1),
			style.Fixed(),
		).
		Part(widget.Part("title-text"),
			style.On(style.Accent),
		).
		Part(widget.Part("back"),
			style.Hidden(),
		).
		Part(widget.Part("icon"),
			style.Width(style.Content),
		).
		Part(widget.Part("action"),
			style.On(style.Accent),
			style.Round(style.RadiusMd),
		).
		Part(widget.Part("action-new"),
			style.On(style.Accent),
		).
		Part(widget.Part("action-cancel"),
			style.Hidden(),
		).
		Part(widget.Part("search"),
			style.Row(style.Space1),
		).
		Part(widget.Part("delconfirm-mount"),
			style.Fixed(),
		).
		Part(widget.Part("delconfirm-actions"),
			style.Row(style.Space1),
		).
		Part(widget.Part("delconfirm-btn"),
			style.On(style.Panel),
			style.Round(style.RadiusSm),
			style.Pad(style.Space1),
		).
		Part(widget.Part("delconfirm-btn-danger"),
			style.On(style.Danger),
			style.Round(style.RadiusSm),
			style.Pad(style.Space1),
		).
		When(widget.Open, widget.Part("action-new"),
			style.Hidden(),
		).
		When(widget.Open, widget.Part("action-cancel"),
			style.Shown(),
		).
		Stylesheet()
}
