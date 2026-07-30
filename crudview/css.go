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
		// Pad is what turns the primary surface into a visible frame: without it
		// the detail card and the aside sit flush against the panel edges and
		// the module reads as three stacked rectangles instead of one view.
		Root(
			style.Split(style.SplitTwoThirds, style.Space2),
			style.Fill(),
			style.As(style.Primary),
			style.Pad(style.Space3),
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
			style.Anchor(),
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
		// The heading needs its own indent: it sits directly on the primary
		// surface with no card of its own to inset it.
		Part(widget.Part("title"),
			style.Row(style.Space1),
			style.Pad(style.Space2),
			style.KeepSize(),
		).
		Part(widget.Part("title-text"),
			style.As(style.Primary),
		).
		// Mobile-only: on a phone the two panels are a swipe strip, so the row
		// that was tapped needs a way back to the list.
		OnlyOn(css.Mobile, widget.Part("back"),
			style.Row(style.SpaceNone),
			style.As(style.Primary),
			style.Round(style.RadiusSm),
			style.Pad(style.Space1),
			style.Width(style.Content),
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
			style.Pad(style.Space3),
			style.Width(style.Full),
			style.CenterContent(),
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
			style.Pad(style.Space3),
			style.Round(style.RadiusMd),
			style.KeepSize(),
		).
		// The input takes what the icon leaves, so the field is the whole card
		// instead of a short box with dead space beside it.
		Part(widget.Part("search-input"),
			style.As(style.Inset),
			style.Round(style.RadiusSm),
			style.Pad(style.Space1),
			style.Grow(),
		).
		// On a phone the desktop Split becomes a horizontal scroll-snap strip:
		// the list is what shows on arrival, and tapping a row slides the detail
		// in from the left, leaving a sliver of the list on the right so it is
		// obvious where you came from. crudview.go already drives the snap with
		// ScrollIntoView on select and on the back button.
		// Pad(SpaceNone) is part of the contract, not a detail: the panels are
		// sized as a share of the scroll container, and any padding on it makes
		// each panel that much narrower than the window, so a strip of the
		// neighbour shows through at rest.
		On(css.Mobile, "",
			style.MasterDetail(style.Most),
			style.Pad(style.SpaceNone),
		).
		// On a phone the action is a floating square pinned to the list's corner
		// instead of a bar across the bottom: the list keeps the whole panel and
		// the button matches the hamburger it shares the screen with.
		On(css.Mobile, widget.Part("action"),
			style.Docked(style.SideEnd, style.Space4),
			style.Width(style.Content),
			style.Pad(style.Space2),
			style.Round(style.RadiusMd),
			style.Raise(style.Floating),
			style.CenterContent(),
		).
		On(css.Mobile, widget.Part("action-new"), style.IconBox(style.IconLg)).
		On(css.Mobile, widget.Part("action-cancel"), style.IconBox(style.IconLg)).
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
