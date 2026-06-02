package rightpanel

import (
	. "github.com/tinywasm/css"
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/html"
)

var (
	clsWrapper      Class = "rp-wrapper"
	clsMain         Class = "rp-main"
	clsHeader       Class = "rp-header"
	clsTitleRow     Class = "rp-title-row"
	clsHeadControls Class = "rp-head-controls"
	clsArticle      Class = "rp-article"
	clsAside        Class = "rp-aside"
	clsAsideHeader  Class = "rp-aside-header"
	clsAsideContent Class = "rp-aside-content"
)

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
// IMPORTANT: All Component implementors passed as slots MUST embed Element as a value,
// not as a pointer. See tinywasm/dom interface.dom.go for details.
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
	Element

	// Module provides the ID for the root wrapper element.
	Module Module

	// Title is rendered as <h1> in the header.
	Title string

	// Head is rendered beside the <h1> (e.g. status badge, icon).
	Head Component

	// HeadControls is rendered below the title row (e.g. select with search).
	HeadControls Component

	// Article is the main content area.
	Article Component

	// AsideControls is rendered at the top of the aside panel (e.g. search + filter).
	AsideControls Component

	// Aside is the content area of the aside panel (e.g. detail view, info card).
	Aside Component
}

// Render builds the layout element tree.
// Implements ViewRenderer.
func (r *RightPanel) Render() *Element {
	// ── root wrapper ─────────────────────────────────────────────────────────
	id := ""
	if r.Module != nil {
		id = r.Module.ModelName()
	}

	wrapper := Div(clsWrapper.AsAttr())
	if id != "" {
		wrapper.ID(id)
	}

	// ── main section ─────────────────────────────────────────────────────────
	main := Section(clsMain.AsAttr())

	// header row: title + Head slot + HeadControls slot
	header := Div(clsHeader.AsAttr())

	titleRow := Div(clsTitleRow.AsAttr())
	if r.Title != "" {
		titleRow.Add(H1().Text(r.Title))
	}
	if r.Head != nil {
		titleRow.Add(r.Head)
	}
	header.Add(titleRow)

	if r.HeadControls != nil {
		header.Add(Div(clsHeadControls.AsAttr()).Add(r.HeadControls))
	}
	main.Add(header)

	// article
	if r.Article != nil {
		main.Add(Article(clsArticle.AsAttr()).Add(r.Article))
	} else {
		main.Add(Article(clsArticle.AsAttr()))
	}

	wrapper.Add(main)

	// ── aside panel ──────────────────────────────────────────────────────────
	if r.AsideControls != nil || r.Aside != nil {
		aside := Aside(clsAside.AsAttr())

		if r.AsideControls != nil {
			aside.Add(Div(clsAsideHeader.AsAttr()).Add(r.AsideControls))
		}
		if r.Aside != nil {
			aside.Add(Div(clsAsideContent.AsAttr()).Add(r.Aside))
		}

		wrapper.Add(aside)
	}

	return wrapper
}
