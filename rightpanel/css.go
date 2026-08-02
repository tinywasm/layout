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
		// Pad is what turns the primary surface into a visible frame: without it
		// the content and the aside sit flush against the panel edges and the
		// module reads as stacked rectangles instead of one view.
		// One gutter everywhere: the frame's pad is the Split's gap, so the
		// margin touching the four sides measures the same as the seam between
		// the two columns.
		Root(
			style.Split(style.SplitTwoThirds, style.Space2),
			style.Fill(),
			style.As(style.Primary),
			style.Pad(style.Space2),
			style.EdgeToEdge(),
		).
		Part(widget.Part("main"),
			style.Stack(style.Space2),
			style.Fill(),
		).
		// The heading needs its own indent: it sits directly on the primary
		// surface with no card of its own to inset it. PadInline, not Pad —
		// the indent is horizontal; vertically the band answers to
		// --control-height, the same token every control measures by, so the
		// frame reads as one rhythm.
		Part(widget.Part("header"),
			style.Row(style.Space1),
			style.PadInline(style.Space2),
			style.ControlBox(),
			style.KeepSize(),
		).
		Part(widget.Part("title-row"),
			style.Row(style.Space2),
			style.Center(),
		).
		// Explicit now that the reset stops <h1> from carrying 2em of its own.
		Part(widget.Part("title"),
			style.As(style.Primary),
			style.FontSize(style.Text2xl),
			style.FontWeight(style.WeightBold),
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
			style.Stack(style.Space1),
			style.As(style.Panel),
			style.Pad(style.Space1),
			style.Fill(),
		).
		// The controls band carries no surface, padding or radius of its own:
		// whatever the consumer puts here — a search bar, a calendar, a select —
		// brings its own. A second frame around a control that already has one
		// reads as a box inside a box. All the band owes it is a refusal to be
		// squashed when the content below grows, which is what KeepSize buys.
		Part(widget.Part("aside-header"),
			style.KeepSize(),
		).
		Part(widget.Part("aside-content"),
			style.Fill(),
			style.Stack(style.SpaceNone),
		).
		Part(widget.Part("aside-footer"),
			style.Row(style.Space1),
			style.KeepSize(),
		).
		// On a phone the desktop Split becomes a horizontal scroll-snap strip:
		// the aside is what shows on arrival, and selecting an item slides the
		// main panel in from the left, leaving a sliver of the aside on the right
		// so it is obvious where you came from.
		// Pad(SpaceNone) is part of the contract, not a detail: the panels are
		// sized as a share of the scroll container, and any padding on it makes
		// each panel that much narrower than the window, so a strip of the
		// neighbour shows through at rest.
		On(css.Mobile, "",
			style.MasterDetail(style.Most),
			style.Pad(style.SpaceNone),
		).
		On(css.Mobile, widget.Part("title"),
			style.FontSize(style.TextBase),
			style.FontWeight(style.WeightBold),
		).
		// The title travels like the action button: fixed to the screen, so it is
		// there on both pages of the swipe strip. On a phone the platform has no
		// header, and this is the only thing naming the section.
		// Compact on a phone: it floats over the content, so it has to read as a
		// chip rather than a banner.
		On(css.Mobile, widget.Part("header"),
			style.Docked(style.Parent, style.EdgeTop, style.SideStart, style.Space4),
			style.Row(style.Space1),
			style.As(style.Primary),
			style.Round(style.RadiusMd),
			style.Pad(style.Space2),
			style.Raise(style.Floating),
			style.Width(style.Content),
		).
		// Reserve the band the floating header and the hamburger occupy, so the
		// content starts below them instead of underneath.
		On(css.Mobile, widget.Part("main"),
			style.PadEdge(style.EdgeTop, style.Space12),
		).
		On(css.Mobile, widget.Part("aside"),
			style.PadEdge(style.EdgeTop, style.Space12),
		)
}

// RenderCSS implements visual contract for rightpanel layout.
func (r *RightPanel) RenderCSS() *css.Stylesheet {
	return r.RenderSheet().Stylesheet()
}
