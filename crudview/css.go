//go:build !wasm

package crudview


import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

// cardInset is the gap between a card's own border and its content. fields
// and list are the two panels crudview paints into rightpanel's article and
// aside — one named constant instead of two independently-chosen literals is
// what keeps them from drifting apart the way they already had (fields
// padded, list flush against its own border).
const cardInset = style.Space1

// RenderSheet returns the style Sheet containing the rules for crudview.
func (v *CrudView) RenderSheet() *style.Sheet {
	return style.For(v).
		// crudview paints no frame: rightpanel owns the root, the columns and
		// the mobile strip. What remains here are the widgets this controller
		// puts INTO those slots.
		// Fill: the fields are the article's whole content now — rp__article is
		// a plain block (page surface, pad, scroll), not a flex column, so a
		// content-height child would collapse to the form's natural height and
		// leave the column floating at the top of the card.
		Part(widget.Part("fields"),
			style.As(style.Inset),
			style.Pad(cardInset),
			style.Scroll(),
			style.Round(style.RadiusMd),
			style.Fill(),
		).
		Part(widget.Part("list"),
			style.As(style.Inset),
			style.Pad(cardInset),
			style.Scroll(),
			style.Round(style.RadiusMd),
			style.Fill(),
		).		// The primary action spans the column, mirroring the search bar above
		// the list rather than sitting as a stray square beside it.
		// ControlBox: the action is a control, and every control answers to
		// --control-height — the same token the search bar now measures by, so
		// the two ends of the column agree by construction instead of by
		// hand-tuned padding.
		Part(widget.Part("action"),
			style.As(style.Primary),
			style.Round(style.RadiusMd),
			style.Pad(style.Space3),
			style.Width(style.Full),
			style.ControlBox(),
			style.CenterContent(),
		).
		// No As(): action-new is the icon INSIDE the Primary "action" button, not
		// a surface of its own. As(Primary) gave it a second Primary box —
		// background + border-radius + (under css.SetGradient) its own gradient
		// slice — a visible square over the button. It inherits the button's
		// white currentColor, exactly like action-cancel below.
		Part(widget.Part("action-new"),
			style.IconBox(style.IconMd),
		).
		Part(widget.Part("action-cancel"),
			style.RevealedBy(widget.Open),
			style.IconBox(style.IconMd),
		).
		// On a phone the action is a floating square instead of a bar across the
		// bottom: the list keeps the whole panel and the button matches the
		// hamburger it shares the screen with. Viewport scope, not Parent — it
		// has to stay reachable on the detail panel too, where its job is to
		// cancel. It still disappears with the module: a fixed descendant of a
		// display:none panel is not rendered, so no other route sees it.
		// ControlBox's min-height stays on: the floating chip grows to the
		// control token rather than hugging the IconLg glyph, which is a bigger
		// target, not a fight.
		//
		// Docked gap and Pad are Space4/Space2, matching platformd's hamburger
		// (OnlyOn(Mobile, "hamburger") in platformd/css.go) exactly, not
		// cardInset: these two are the screen's floating chrome, a different
		// pairing from fields/list's in-card content inset, and pinning both
		// floating buttons to the same corner offset and padding is what makes
		// them read as one design language instead of two unrelated badges.
		On(css.Mobile, widget.Part("action"),
			style.Docked(style.Parent, style.EdgeBottom, style.SideEnd, style.Space4),
			style.Width(style.Content),
			style.Pad(style.Space2),
			style.Round(style.RadiusMd),
			style.Raise(style.Floating),
			style.CenterContent(),
		).
		On(css.Mobile, widget.Part("action-new"), style.IconBox(style.IconLg)).
		On(css.Mobile, widget.Part("action-cancel"), style.IconBox(style.IconLg)).
		// Bare, not Inset, on a phone: the grey card + border of fields/list
		// adds nothing over the panel's own page surface, and the list is
		// where the user first wants whitespace to breathe. Bare strips
		// background and border without touching the text color — Subtle
		// would also fix the grey but would silently mute inherited text,
		// which is a different statement than "this is not a card".
		// The rows (targetlist, white Page) and the inputs (fieldset, white
		// Page) each declare their own surfaces, so nothing loses contrast.
		On(css.Mobile, widget.Part("fields"),
			style.As(style.Bare),
		).
		On(css.Mobile, widget.Part("list"),
			style.As(style.Bare),
		).
		// cardInset even on mobile: fields and list keep the same pad between
		// the panel border and the content on every breakpoint, because the
		// list no longer has to buy space for the ⋮ menu in the sliver
		// MasterDetail(Most) leaves visible of it — the ⋮ is unreachable from
		// the sliver now (see targetlist/css.go's PartList comment), so the
		// indent budget is simply the form's: rightpanel 4 + cardInset 4 +
		// targetlist's own 8 = 16px on both columns. The old flush was a
		// deliberate mobile exception to cardInset's promise; it is retired
		// with the premise that justified it.
		//
		// FloatingChrome(EdgeBottom, IconLg, Space4): the action button above
		// is Docked to THIS panel's corner, floating over content this Part
		// contains — the icon (IconLg, matching action-new/action-cancel's own
		// mobile size two rules up) plus the same Space4 the button is offset
		// from the edge by, doubled, is its measured 56px-tall + 16px-offset =
		// 72px footprint. This Part is not itself what scrolls — targetlist's
		// own Fill()+Scroll() two levels in is — but --floating-bottom is an
		// inherited custom property, so it reaches that descendant without
		// either package knowing the other's class name. Whoever changes the
		// button's icon size or its Docked gap must update this call to match;
		// nothing enforces the two staying in sync beyond that they are three
		// lines apart.
		On(css.Mobile, widget.Part("list"),
			style.Pad(cardInset),
			style.FloatingChrome(style.EdgeBottom, style.IconLg, style.Space4),
		).
		// Amber while open, matching the selection language the list already
		// uses: the button IS what is currently "selected" in the sense that
		// its own panel is the one on screen. Cancelling (closing) drops it
		// back to Primary blue, since at that point nothing is active anymore.
		When(widget.Open, widget.Part("action"),
			style.As(style.AccentInverse),
		).
		// The delete-confirmation modal's holder must not cost the frame a flex
		// share: as a plain in-flow child of rightpanel's Split it would take a
		// third of the width (the skeleton's `.rp > *` grow beats this sheet's
		// KeepSize across sheets). Pinning it fixed to the viewport corner takes
		// it out of the flow entirely; it is 0x0 while the modal is closed, and
		// the modal's own Backdrop(Viewport) covers the screen when it opens.
		Part(widget.Part("delconfirm-mount"),
			style.Docked(style.Viewport, style.EdgeTop, style.SideStart, style.SpaceNone),
		).
		// Space2: the same step modaldialog's panel puts between its header and
		// its body, so title→question and question→buttons read as one rhythm.
		Part(widget.Part("delconfirm-body"),
			style.Stack(style.Space2),
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
