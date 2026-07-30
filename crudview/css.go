//go:build !wasm

package crudview

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

// RenderSheet returns the style Sheet containing the rules for crudview.
func (v *CrudView) RenderSheet() *style.Sheet {
	return style.For(v).
		Root(
			style.Split(style.SplitTwoThirds, style.Space2),
			style.As(style.Primary),
			style.EdgeToEdge(),
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
			style.As(style.Inset),
			style.Pad(style.Space2),
			style.Scroll(),
			style.Round(style.RadiusMd),
		).
		Part(widget.Part("aside"),
			style.Stack(style.Space1),
			style.As(style.Panel),
			style.Pad(style.Space1),
			style.Fill(),
		).
		Part(widget.Part("aside-content"),
			style.Fill(),
			style.Stack(style.SpaceNone),
		).
		Part(widget.Part("list"),
			style.As(style.Inset),
			style.Scroll(),
			style.Round(style.RadiusMd),
		).
		Part(widget.Part("actions"),
			style.Row(style.Space1),
		).
		Part(widget.Part("title"),
			style.Row(style.Space1),
			style.KeepSize(),
		).
		Part(widget.Part("title-text"),
			style.As(style.Primary),
		).
		Part(widget.Part("back"),
			style.RevealedBy(widget.Open),
		).
		// A bare <svg> with no box falls back to 300x150; IconBox pins it.
		Part(widget.Part("icon"),
			style.IconBox(style.IconMd),
		).
		Part(widget.Part("action"),
			style.As(style.Primary),
			style.Round(style.RadiusMd),
		).
		Part(widget.Part("action-new"),
			style.As(style.Primary),
		).
		Part(widget.Part("action-cancel"),
			style.RevealedBy(widget.Open),
		).
		Part(widget.Part("search"),
			style.Row(style.Space1),
		).
		Part(widget.Part("delconfirm-mount"),
			style.KeepSize(),
		).
		Part(widget.Part("delconfirm-actions"),
			style.Row(style.Space1),
		).
		Part(widget.Part("delconfirm-btn"),
			style.As(style.Panel),
			style.Round(style.RadiusSm),
			style.Pad(style.Space1),
		).
		Part(widget.Part("delconfirm-btn-danger"),
			style.As(style.Danger),
			style.Round(style.RadiusSm),
			style.Pad(style.Space1),
		).
		When(widget.Open, widget.Part("action-new"),
			style.RevealedBy(widget.Open),
		)
}

// RenderCSS implements the visual contract for crudview using the style DSL.
func (v *CrudView) RenderCSS() *css.Stylesheet {
	return v.RenderSheet().Stylesheet()
}
