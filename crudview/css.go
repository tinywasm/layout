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
		// crudview paints no frame: rightpanel owns the root, the columns and
		// the mobile strip. What remains here are the widgets this controller
		// puts INTO those slots.
		// Fill: the fields are the article's whole content now — rp__article is
		// a plain block (page surface, pad, scroll), not a flex column, so a
		// content-height child would collapse to the form's natural height and
		// leave the column floating at the top of the card.
		Part(widget.Part("fields"),
			style.As(style.Inset),
			style.Pad(style.Space2),
			style.Scroll(),
			style.Round(style.RadiusMd),
			style.Fill(),
		).
		Part(widget.Part("list"),
			style.As(style.Inset),
			style.Scroll(),
			style.Round(style.RadiusMd),
		).
		// The primary action spans the column, mirroring the search bar above
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
		Part(widget.Part("action-new"),
			style.As(style.Primary),
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
