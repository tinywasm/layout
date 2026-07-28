//go:build !wasm

package platformd

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

// RenderCSS implements the visual contract for platformd using the style DSL.
func (p *Platform) RenderCSS() *css.Stylesheet {
	return style.Of(NamePlatform).
		Root(
			style.On(style.Page),
			style.Stack(style.Space0),
			style.Fill(),
			style.Animate(style.MotionSlow),
		).
		Part(widget.Part("header"),
			style.Row(style.Space2),
			style.On(style.Panel),
			style.Pad(style.Space1),
		).
		Part(widget.Part("user-block"),
			style.Row(style.Space1),
			style.Text(style.TextBase),
			style.FontWeight(style.WeightBold),
		).
		Part(widget.Part("header-right"),
			style.Row(style.Space2),
		).
		Part(widget.Part("msg-desktop"),
			style.Row(style.Space1),
		).
		Part(widget.Part("area"),
			style.Text(style.TextBase),
			style.On(style.Muted),
		).
		Part(widget.Part("msg-mobile"),
			style.Stack(style.Space1),
		).
		Part(widget.Part("menu"),
			style.Stack(style.Space1),
			style.On(style.Panel),
			style.Animate(style.MotionSlow),
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
			style.Text(style.TextBase),
		).
		Part(widget.Part("nav-icon"),
			style.Width(style.Content),
		).
		Part(widget.Part("nav-active"),
			style.On(style.Selected),
		).
		Part(widget.Part("stage"),
			style.Cover(),
			style.Fill(),
		).
		Part(widget.Part("panel"),
			style.Hidden(),
		).
		Part(widget.Part("panel-active"),
			style.Shown(),
			style.Fill(),
		).
		Part(widget.Part("orientation-warn"),
			style.Hidden(),
		).
		Part(widget.Part("msg"),
			style.On(style.Panel),
			style.Pad(style.Space2),
			style.Round(style.RadiusMd),
		).
		Part(widget.Part("hamburger"),
			style.Row(style.Space1),
		).
		Part(widget.Part("nav-overlay"),
			style.Hidden(),
		).
		When(widget.Open, widget.Part("menu"),
			style.Shown(),
		).
		When(widget.Open, widget.Part("nav-overlay"),
			style.Shown(),
		).
		Stylesheet()
}
