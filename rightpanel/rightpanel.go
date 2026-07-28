package rightpanel

import (
	"github.com/tinywasm/layout"
	"github.com/tinywasm/widget"

	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/html"
)

const NameRightPanel widget.Name = "rp"

var (
	clsWrapper      = NameRightPanel.Root()
	clsMain         = NameRightPanel.Class("main")
	clsHeader       = NameRightPanel.Class("header")
	clsTitleRow     = NameRightPanel.Class("title-row")
	clsTitle        = NameRightPanel.Class("title")
	clsHeadControls = NameRightPanel.Class("controls")
	clsArticle      = NameRightPanel.Class("article")
	clsAside        = NameRightPanel.Class("aside")
	clsAsideHeader  = NameRightPanel.Class("aside-header")
	clsAsideContent = NameRightPanel.Class("aside-content")
)

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
	Module layout.Module

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

	wrapper := Div().Set(clsWrapper.AsAttr())
	if id != "" {
		wrapper.ID(id)
	}

	// ── main section ─────────────────────────────────────────────────────────
	main := Section().Set(clsMain.AsAttr())

	// header row: title + Head slot + HeadControls slot
	header := Div().Set(clsHeader.AsAttr())

	titleRow := Div().Set(clsTitleRow.AsAttr())
	if r.Title != "" {
		titleRow.Child(H1().Set(clsTitle.AsAttr()).Text(r.Title))
	}
	if r.Head != nil {
		titleRow.Child(r.Head)
	}
	header.Child(titleRow)

	if r.HeadControls != nil {
		header.Child(Div().Set(clsHeadControls.AsAttr()).Child(r.HeadControls))
	}
	main.Child(header)

	// article
	article := Article().Set(clsArticle.AsAttr())
	if r.Article != nil {
		article.Child(r.Article)
	}
	main.Child(article)

	wrapper.Child(main)

	// ── aside panel ──────────────────────────────────────────────────────────
	if r.AsideControls != nil || r.Aside != nil {
		aside := Aside().Set(clsAside.AsAttr())

		if r.AsideControls != nil {
			aside.Child(Div().Set(clsAsideHeader.AsAttr()).Child(r.AsideControls))
		}
		if r.Aside != nil {
			aside.Child(Div().Set(clsAsideContent.AsAttr()).Child(r.Aside))
		}

		wrapper.Child(aside)
	}

	return wrapper
}
