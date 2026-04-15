package rightpanel

import "github.com/tinywasm/dom"

// Module is the interface the consumer must satisfy to provide the layout ID.
// Any struct with a ModelName() string method qualifies (e.g. ORM model structs).
type Module interface {
	ModelName() string
}

// RightPanel is a two-column layout skeleton:
//   - Left: main content area with header (title + controls) and article.
//   - Right: aside panel with its own header (controls) and content.
//
// All slots are optional. A nil slot is simply not rendered.
// The layout does not define what the slots contain — that is the consumer's job.
//
// Usage:
//
//	panel := &rightpanel.RightPanel{
//	    Module:        myModel,          // implements ModelName() string
//	    Title:         "Users",
//	    HeadControls:  mySelectSearch,
//	    Article:       myTable,
//	    AsideControls: myFilterBar,
//	    Aside:         myDetailPanel,
//	}
//	panel.Render()
type RightPanel struct {
	*dom.Element

	// Module provides the ID for the root wrapper element.
	Module Module

	// Title is rendered as <h1> in the header.
	Title string

	// Head is rendered beside the <h1> (e.g. status badge, icon).
	Head dom.Component

	// HeadControls is rendered below the title row (e.g. select with search).
	HeadControls dom.Component

	// Article is the main content area.
	Article dom.Component

	// AsideControls is rendered at the top of the aside panel (e.g. search + filter).
	AsideControls dom.Component

	// Aside is the content area of the aside panel (e.g. detail view, info card).
	Aside dom.Component
}

// Render builds the layout element tree.
// Implements dom.ViewRenderer.
func (r *RightPanel) Render() *dom.Element {
	if r.Element == nil {
		r.Element = &dom.Element{}
	}

	// ── root wrapper ─────────────────────────────────────────────────────────
	id := ""
	if r.Module != nil {
		id = r.Module.ModelName()
	}

	wrapper := dom.Div().Class("rp-wrapper")
	if id != "" {
		wrapper.ID(id)
	}

	// ── main section ─────────────────────────────────────────────────────────
	main := dom.Section().Class("rp-main")

	// header row: title + Head slot + HeadControls slot
	header := dom.Div().Class("rp-header")

	titleRow := dom.Div().Class("rp-title-row")
	if r.Title != "" {
		titleRow.Add(dom.H1().Text(r.Title))
	}
	if r.Head != nil {
		titleRow.Add(r.Head)
	}
	header.Add(titleRow)

	if r.HeadControls != nil {
		header.Add(dom.Div().Class("rp-head-controls").Add(r.HeadControls))
	}
	main.Add(header)

	// article
	if r.Article != nil {
		main.Add(dom.Article().Class("rp-article").Add(r.Article))
	} else {
		main.Add(dom.Article().Class("rp-article"))
	}

	wrapper.Add(main)

	// ── aside panel ──────────────────────────────────────────────────────────
	if r.AsideControls != nil || r.Aside != nil {
		aside := dom.Aside().Class("rp-aside")

		if r.AsideControls != nil {
			aside.Add(dom.Div().Class("rp-aside-header").Add(r.AsideControls))
		}
		if r.Aside != nil {
			aside.Add(dom.Div().Class("rp-aside-content").Add(r.Aside))
		}

		wrapper.Add(aside)
	}

	return wrapper
}
