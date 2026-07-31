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
		// The drawer's head. It exists on a phone, where there is no header for
		// identity chrome to live in, and on a wide screen only while the rail
		// is expanded — squeezed into a 57px rail it wraps into nonsense.
		OnlyOn(css.Mobile, widget.Part("drawer-head"),
			style.Stack(style.Space1),
			style.KeepSize(),
			style.Pad(style.Space2),
			style.As(style.Inset),
			style.Round(style.RadiusMd),
		).
		OnlyOn(css.Mobile, widget.Part("app-name"),
			style.Row(style.SpaceNone),
			style.FontSize(style.TextBase),
			style.FontWeight(style.WeightBold),
		).
		OnlyOn(css.Mobile, widget.Part("drawer-actions"),
			style.Row(style.Space1),
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
		// A deck, not a stack of hidden panels: RevealedBy toggles `display`,
		// which is discrete and cannot transition, so a route change jumped. The
		// movement comes from the scroller instead — Activate() calls
		// ScrollIntoView and the strip slides. Every module stays mounted.
		Part(widget.Part("stage"),
			style.Deck(style.SpaceNone),
			style.Fill(),
		).
		// Scroll(), not Fill(): a module taller than the viewport has to scroll
		// inside its own page of the deck. Scroll() is Fill() plus overflow-y.
		Part(widget.Part("panel"),
			style.Stack(style.SpaceNone),
			style.Scroll(),
		).
		// Grow() for its min-width:0 — without it the rail is sized by whatever
		// it contains, so revealing the labels on hover widened it and pushed
		// the stage. Its width is the Sidebar's rail token and nothing else.
		Part(widget.Part("menu"),
			style.Anchor(),
			style.Stack(style.Space1),
			style.As(style.Panel),
			style.Fill(),
			style.Grow(),
		).
		Part(widget.Part("drawer-panel"),
			style.Stack(style.Space1),
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
		// Glyph, not As: an item that is merely available is a coloured icon on
		// the rail's own surface. Only the current one is filled.
		Part(widget.Part("nav-link"),
			style.Row(style.Space1),
			style.Pad(style.Space2),
			style.Fill(),
			style.CenterContent(),
			style.Glyph(style.Primary),
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
		// The route you are on is the one filled block in the rail; its icon
		// rides the filled surface through currentColor.
		When(widget.Current, widget.Part("nav-link"),
			style.As(style.Primary),
		).
		// The whole control lights up, not just the glyph inside it: a hover is
		// about the target you are aiming at, and the target is the button.
		Cue(widget.Hover, widget.Part("nav-link"),
			style.As(style.Accent),
		).
		// Hovering the rail floats the whole panel out over the content at label
		// width. The panel leaves the flow, so the rail's box — already pinned
		// to the Sidebar's rail token by Grow's min-width:0 — cannot change, and
		// nothing beside it moves.
		CueWithin(widget.Hover, widget.Part("menu"), widget.Part("drawer-panel"),
			style.Docked(style.Parent, style.EdgeTop, style.SideEnd, style.SpaceNone),
			style.Width(style.Content),
			style.As(style.Panel),
			style.Raise(style.Floating),
			style.Pad(style.Space1),
			style.Round(style.RadiusMd),
			style.Stack(style.Space1),
		).
		// The labels only exist while the rail is expanded — or on a phone,
		// where the drawer is two thirds of the viewport and has room for them.
		CueWithin(widget.Hover, widget.Part("menu"), widget.Part("link-text"),
			style.Row(style.SpaceNone),
			style.FontSize(style.TextBase),
		).
		CueWithin(widget.Hover, widget.Part("menu"), widget.Part("nav-link"),
			style.Row(style.Space2),
			style.Pad(style.Space2),
		).
		// The head rides along with the expansion: the theme toggle and the user
		// are reachable on a wide screen the moment the rail opens.
		CueWithin(widget.Hover, widget.Part("menu"), widget.Part("drawer-head"),
			style.Stack(style.Space1),
			style.KeepSize(),
			style.Pad(style.Space2),
			style.As(style.Inset),
			style.Round(style.RadiusMd),
		).
		CueWithin(widget.Hover, widget.Part("menu"), widget.Part("app-name"),
			style.Row(style.SpaceNone),
			style.FontSize(style.TextBase),
			style.FontWeight(style.WeightBold),
		).
		CueWithin(widget.Hover, widget.Part("menu"), widget.Part("drawer-actions"),
			style.Row(style.Space1),
		).
		// ── mobile-only chrome ────────────────────────────────────────────────
		// No header on a phone: the module brings its own title and the chrome
		// floats over the content. The button pins to the screen so it is
		// reachable from either page of a swipe strip.
		On(css.Mobile, widget.Part("header"),
			style.Hide(),
		).
		OnlyOn(css.Mobile, widget.Part("hamburger"),
			style.Row(style.Space1),
			style.As(style.Primary),
			style.Pad(style.Space2),
			style.Round(style.RadiusSm),
			style.Raise(style.Floating),
			style.CenterContent(),
			style.Docked(style.Viewport, style.EdgeTop, style.SideEnd, style.Space4),
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
