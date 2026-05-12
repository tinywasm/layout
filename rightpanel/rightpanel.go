package rightpanel

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/dom"
)

var (
	ClsWrapper      css.Class = "rp-wrapper"
	ClsMain         css.Class = "rp-main"
	ClsHeader       css.Class = "rp-header"
	ClsTitleRow     css.Class = "rp-title-row"
	ClsHeadControls css.Class = "rp-head-controls"
	ClsArticle      css.Class = "rp-article"
	ClsAside        css.Class = "rp-aside"
	ClsAsideHeader  css.Class = "rp-aside-header"
	ClsAsideContent css.Class = "rp-aside-content"
)

var (
	TokenTitleHeight    = css.Token{Name: "--rp-title-height", Fallback: "8vh"}
	TokenContentHeight  = css.Token{Name: "--rp-content-height", Fallback: "89vh"}
	TokenControlsHeight = css.Token{Name: "--rp-controls-height", Fallback: "3vh"}
	TokenMainWidth      = css.Token{Name: "--rp-main-width", Fallback: "66vw"}
	TokenAsideWidth     = css.Token{Name: "--rp-aside-width", Fallback: "30vw"}
	TokenGap            = css.Token{Name: "--rp-gap", Fallback: "var(--space-2)"}
	TokenBorderColor    = css.Token{Name: "--rp-border-color", Fallback: "var(--color-muted)"}
	TokenBg             = css.Token{Name: "--rp-bg", Fallback: "var(--color-surface)"}
	TokenAsideBg        = css.Token{Name: "--rp-aside-bg", Fallback: "var(--color-on-surface)"}
	TokenTitleColor     = css.Token{Name: "--rp-title-color", Fallback: "var(--color-secondary)"}
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
// IMPORTANT: All dom.Component implementors passed as slots MUST embed dom.Element as a value,
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

	wrapper := dom.Div(dom.Class(ClsWrapper))
	if id != "" {
		wrapper.ID(id)
	}

	// ── main section ─────────────────────────────────────────────────────────
	main := dom.Section(dom.Class(ClsMain))

	// header row: title + Head slot + HeadControls slot
	header := dom.Div(dom.Class(ClsHeader))

	titleRow := dom.Div(dom.Class(ClsTitleRow))
	if r.Title != "" {
		titleRow.Add(dom.H1().Text(r.Title))
	}
	if r.Head != nil {
		titleRow.Add(r.Head)
	}
	header.Add(titleRow)

	if r.HeadControls != nil {
		header.Add(dom.Div(dom.Class(ClsHeadControls)).Add(r.HeadControls))
	}
	main.Add(header)

	// article
	if r.Article != nil {
		main.Add(dom.Article(dom.Class(ClsArticle)).Add(r.Article))
	} else {
		main.Add(dom.Article(dom.Class(ClsArticle)))
	}

	wrapper.Add(main)

	// ── aside panel ──────────────────────────────────────────────────────────
	if r.AsideControls != nil || r.Aside != nil {
		aside := dom.Aside(dom.Class(ClsAside))

		if r.AsideControls != nil {
			aside.Add(dom.Div(dom.Class(ClsAsideHeader)).Add(r.AsideControls))
		}
		if r.Aside != nil {
			aside.Add(dom.Div(dom.Class(ClsAsideContent)).Add(r.Aside))
		}

		wrapper.Add(aside)
	}

	return wrapper
}
