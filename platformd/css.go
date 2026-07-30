//go:build !wasm

package platformd

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

// RenderSheet returns the style Sheet containing the rules for platformd.
func (p *Platform) RenderSheet() *style.Sheet {
	return style.For(p).
		// The outermost frame of the application: locks to the viewport, stacks
		// header over body. HideOverflow() keeps the frame itself from ever
		// scrolling — a tall module scrolls inside its own panel instead.
		Root(
			style.Cover(),
			style.HideOverflow(),
			style.As(style.Page),
		).
		Part(widget.Part("header"),
			style.Row(style.Space2),
			style.KeepSize(),
			style.As(style.Panel),
			style.Pad(style.Space2),
		).
		Part(widget.Part("user-block"),
			style.Row(style.Space1),
			style.FontSize(style.TextBase),
			style.FontWeight(style.WeightBold),
		).
		// Fill() here is what pushes header-right to the far edge: it grows to
		// take the free space between the two blocks.
		Part(widget.Part("msg-slot"),
			style.Row(style.Space1),
			style.Fill(),
		).
		Part(widget.Part("msg"),
			style.Pad(style.Space1),
			style.Round(style.RadiusSm),
		).
		Part(widget.Part("msg-info"), style.As(style.Subtle)).
		Part(widget.Part("msg-success"), style.As(style.Success)).
		Part(widget.Part("msg-warning"), style.As(style.Highlight)).
		Part(widget.Part("msg-error"), style.As(style.Danger)).
		Part(widget.Part("header-right"),
			style.Row(style.Space2),
			style.KeepSize(),
		).
		Part(widget.Part("area"),
			style.FontSize(style.TextBase),
			style.As(style.Subtle),
		).
		// The rail sits at the inline-end edge; the stage takes everything else.
		// Below the stage's minimum width the two reflow into one column with no
		// media query — that is Sidebar's own behaviour, not something to add.
		Part(widget.Part("body"),
			style.Sidebar(style.SideEnd, style.RailNarrow, style.SpaceNone),
			style.Fill(),
		).
		Part(widget.Part("stage"),
			style.Fill(),
			style.HideOverflow(),
		).
		// Scroll(), not Fill(): the stage clips, so a module taller than the
		// viewport has to scroll inside its own panel or its overflow is
		// unreachable. Scroll() is Fill() plus overflow-y.
		Part(widget.Part("panel"),
			style.Stack(style.SpaceNone),
			style.Scroll(),
			style.RevealedBy(widget.Current),
		).
		Part(widget.Part("menu"),
			style.Stack(style.Space1),
			style.As(style.Panel),
			style.Fill(),
		).
		Part(widget.Part("navbar"),
			style.Stack(style.SpaceNone),
			style.Fill(),
		).
		Part(widget.Part("nav-item"),
			style.Row(style.SpaceNone),
			style.Fill(),
		).
		Part(widget.Part("nav-link"),
			style.Row(style.Space1),
			style.Pad(style.Space2),
			style.Fill(),
			style.Animate(style.MotionFast),
		).
		// A bare <svg> with no box falls back to 300x150 and wrecks the rail.
		Part(widget.Part("nav-icon"),
			style.IconBox(style.IconLg),
		).
		// The rail is icon-only: at RailNarrow a label does not fit, and forcing
		// it in is what widened the rail past its token. The mobile drawer is
		// two thirds of the viewport, so there the label rides along.
		// Row() is not decoration here: OnlyOn hides the part outside the device
		// and only a flow puts a display back on it inside, so a rule carrying
		// nothing but FontSize would stay hidden everywhere.
		OnlyOn(css.Mobile, widget.Part("link-text"),
			style.Row(style.SpaceNone),
			style.FontSize(style.TextBase),
		).
		// The active route reads as "current", the same vocabulary the rail and
		// crudview's list rows share. It is a STATE, never a class.
		When(widget.Current, widget.Part("nav-link"),
			style.As(style.Highlight),
		).
		Cue(widget.Hover, widget.Part("nav-link"),
			style.As(style.Panel),
		).
		// ── mobile-only chrome ────────────────────────────────────────────────
		OnlyOn(css.Mobile, widget.Part("hamburger"),
			style.Row(style.Space1),
			style.As(style.Primary),
			style.Pad(style.Space2),
			style.Round(style.RadiusSm),
			style.Width(style.Content),
		).
		OnlyOn(css.Mobile, widget.Part("nav-overlay"),
			style.Backdrop(style.Viewport),
			style.Veil(),
			style.RevealedBy(widget.Open),
		).
		// On a phone the rail stops being a column and becomes a panel that
		// slides in from the edge, gated by the same Open state as the overlay.
		On(css.Mobile, widget.Part("menu"),
			style.Drawer(style.SideEnd, style.TwoThirds),
			style.Stack(style.Space1),
			style.As(style.Panel),
			style.RevealedBy(widget.Open),
		)
}

// RenderCSS implements the visual contract for platformd using the style DSL.
func (p *Platform) RenderCSS() *css.Stylesheet {
	return p.RenderSheet().Stylesheet()
}
