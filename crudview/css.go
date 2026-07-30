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
		// Fill() so the view takes the whole height of the platform panel that
		// hosts it; without it the module stops at its content and leaves dead
		// space under the stage.
		Root(
			style.Split(style.SplitTwoThirds, style.Space2),
			style.Fill(),
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
		// The <article> between the title and the fields carries no style of
		// its own, so without Fill() it stops at its content and the fields
		// card floats halfway up the blue panel.
		Part(widget.Part("article"),
			style.Stack(style.SpaceNone),
			style.Fill(),
		).
		Part(widget.Part("list"),
			style.As(style.Inset),
			style.Scroll(),
			style.Round(style.RadiusMd),
		).
		Part(widget.Part("actions"),
			style.Row(style.Space1),
			style.KeepSize(),
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
		// The primary action spans the column, mirroring the search bar above
		// the list rather than sitting as a stray square beside it.
		Part(widget.Part("action"),
			style.As(style.Primary),
			style.Round(style.RadiusMd),
			style.Pad(style.Space2),
			style.Width(style.Full),
		).
		Part(widget.Part("action-new"),
			style.As(style.Primary),
			style.IconBox(style.IconMd),
		).
		Part(widget.Part("action-cancel"),
			style.RevealedBy(widget.Open),
			style.IconBox(style.IconMd),
		).
		// The search bar is a card in its own right, the same height treatment
		// the action bar gets at the other end of the column.
		Part(widget.Part("search"),
			style.Row(style.Space1),
			style.As(style.Panel),
			style.Pad(style.Space2),
			style.Round(style.RadiusMd),
			style.KeepSize(),
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
