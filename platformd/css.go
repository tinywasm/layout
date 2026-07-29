//go:build !wasm

package platformd

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

// RenderCSS implements the visual contract for platformd using the style DSL.
func (p *Platform) RenderCSS() *css.Stylesheet {
	return style.For(p).
		Root(
			style.As(style.Page),
			style.Stack(style.SpaceNone),
			style.Fill(),
			style.Animate(style.MotionSlow),
		).
		Part(widget.Part("header"),
			style.Row(style.Space2),
			style.As(style.Panel),
			style.Pad(style.Space1),
		).
		Part(widget.Part("user-block"),
			style.Row(style.Space1),
			style.FontSize(style.TextBase),
			style.FontWeight(style.WeightBold),
		).
		Part(widget.Part("header-right"),
			style.Row(style.Space2),
		).
		Part(widget.Part("msg-desktop"),
			style.Row(style.Space1),
		).
		Part(widget.Part("area"),
			style.FontSize(style.TextBase),
			style.As(style.Subtle),
		).
		Part(widget.Part("msg-mobile"),
			style.Stack(style.Space1),
		).
		Part(widget.Part("menu"),
			style.Stack(style.Space1),
			style.As(style.Panel),
			style.Animate(style.MotionSlow),
			style.RevealedBy(widget.Open),
		).
		Part(widget.Part("navbar"),
			style.Stack(style.Space1),
		).
		Part(widget.Part("nav-item"),
			style.Row(style.Space1),
		).
		Part(widget.Part("nav-link"),
			style.Row(style.Space1),
			style.Pad(style.Space2),
		).
		Part(widget.Part("link-text"),
			style.FontSize(style.TextBase),
		).
		Part(widget.Part("nav-icon"),
			style.Width(style.Content),
		).
		Part(widget.Part("nav-active"),
			style.As(style.Highlight),
		).
		Part(widget.Part("stage"),
			style.FillCentered(),
			style.Fill(),
		).
		Part(widget.Part("panel"),
			style.RevealedBy(widget.Open),
		).
		Part(widget.Part("panel-active"),
			style.Stack(style.SpaceNone),
			style.Fill(),
		).
		Part(widget.Part("orientation-warn"),
			style.RevealedBy(widget.Open),
		).
		Part(widget.Part("msg"),
			style.As(style.Panel),
			style.Pad(style.Space2),
			style.Round(style.RadiusMd),
		).
		Part(widget.Part("hamburger"),
			style.Row(style.Space1),
		).
		Part(widget.Part("nav-overlay"),
			style.RevealedBy(widget.Open),
		).
		Stylesheet()
}
