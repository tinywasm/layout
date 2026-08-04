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
	clsAsideFooter  = NameRightPanel.Class("aside-footer")
)

func (r *RightPanel) WidgetName() widget.Name { return NameRightPanel }
func (r *RightPanel) WidgetKind() widget.Kind { return widget.Region }

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
//	    AsideFooter:   myActionButton,
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

	// AsideFooter is rendered at the bottom of the aside panel, below the
	// content (e.g. a primary action button). It keeps its size while the
	// content between it and AsideControls takes the slack.
	AsideFooter Component
}

// Render builds the layout element tree.
// Implements ViewRenderer.
func (r *RightPanel) Render() *Element {
	// ── root wrapper ─────────────────────────────────────────────────────────
	wrapper := Div().Set(clsWrapper.AsAttr()).ID(r.panelID())

	// ── main section ─────────────────────────────────────────────────────────
	main := Section().Set(clsMain.AsAttr()).ID(r.MainPanelID())

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
	if r.AsideControls != nil || r.Aside != nil || r.AsideFooter != nil {
		aside := Aside().Set(clsAside.AsAttr()).ID(r.AsidePanelID())

		if r.AsideControls != nil {
			aside.Child(Div().Set(clsAsideHeader.AsAttr()).Child(r.AsideControls))
		}
		if r.Aside != nil {
			aside.Child(Div().Set(clsAsideContent.AsAttr()).Child(r.Aside))
		}
		if r.AsideFooter != nil {
			aside.Child(Div().Set(clsAsideFooter.AsAttr()).Child(r.AsideFooter))
		}

		wrapper.Child(aside)
	}

	return wrapper
}

// MainPanelID and AsidePanelID identify the two scroll-snap targets of the
// mobile strip. A host that drives the snap (see ShowMain/ShowAside) does not
// need them; they are exported because a host may want to link to a panel.
func (r *RightPanel) MainPanelID() string  { return r.panelID() + ".main" }
func (r *RightPanel) AsidePanelID() string { return r.panelID() + ".aside" }

// ShowMain brings the main panel into view; ShowAside brings the aside back.
//
// Both are no-ops on a wide screen, and that guard is the whole point:
// ScrollIntoView walks EVERY scrollable ancestor, not just the nearest one the
// caller had in mind. Side by side there is nothing to scroll here, so an
// unguarded call reached the platform's module deck instead and slid the whole
// application to the next module.
func (r *RightPanel) ShowMain()  { r.showPanel(r.MainPanelID()) }
func (r *RightPanel) ShowAside() { r.showPanel(r.AsidePanelID()) }

func (r *RightPanel) showPanel(id string) {
	strip, ok := Get(r.panelID())
	if !ok || !strip.ScrollsX() {
		return
	}
	if el, ok := Get(id); ok {
		el.ScrollIntoView()
	}
}

// panelID is the id stamped on the wrapper element: the module's name when one
// is set, the element's own generated id otherwise.
func (r *RightPanel) panelID() string {
	if r.Module != nil {
		return r.Module.ModelName()
	}
	return r.GetID()
}
