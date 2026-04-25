package rightpanel_test

import (
	"strings"
	"testing"

	"github.com/tinywasm/dom"
	"github.com/tinywasm/layout/rightpanel"
)

// stubModule implements Module for tests.
type stubModule struct{ name string }

func (s stubModule) ModelName() string { return s.name }

// stubComponent implements dom.Component for tests.
type stubComponent struct{ html string }

func (s *stubComponent) GetID() string             { return "stub" }
func (s *stubComponent) SetID(_ string)            {}
func (s *stubComponent) RenderHTML() string        { return s.html }
func (s *stubComponent) Children() []dom.Component { return nil }

func TestRightPanel_RenderHTML_WithAllSlots(t *testing.T) {
	panel := &rightpanel.RightPanel{
		Module:        stubModule{"users"},
		Title:         "Users",
		Head:          &stubComponent{"<span>badge</span>"},
		HeadControls:  &stubComponent{"<select></select>"},
		Article:       &stubComponent{"<table></table>"},
		AsideControls: &stubComponent{"<input type=search>"},
		Aside:         &stubComponent{"<ul></ul>"},
	}

	el := panel.Render()
	html := el.RenderHTML()

	checks := []struct {
		label, want string
	}{
		{"root id", "id='users'"},
		{"wrapper class", "class='rp-wrapper'"},
		{"main class", "class='rp-main'"},
		{"header class", "class='rp-header'"},
		{"title row", "class='rp-title-row'"},
		{"h1 title", "<h1>Users</h1>"},
		{"Head slot", "<span>badge</span>"},
		{"HeadControls slot", "<select></select>"},
		{"article class", "class='rp-article'"},
		{"Article slot", "<table></table>"},
		{"aside class", "class='rp-aside'"},
		{"aside header", "class='rp-aside-header'"},
		{"AsideControls slot", "<input type=search>"},
		{"aside content", "class='rp-aside-content'"},
		{"Aside slot", "<ul></ul>"},
	}

	for _, c := range checks {
		if !strings.Contains(html, c.want) {
			t.Errorf("[%s] expected %q in HTML:\n%s", c.label, c.want, html)
		}
	}
}

func TestRightPanel_RenderHTML_AsideOmittedWhenNil(t *testing.T) {
	panel := &rightpanel.RightPanel{
		Module:  stubModule{"orders"},
		Title:   "Orders",
		Article: &stubComponent{"<table></table>"},
		// No AsideControls, no Aside
	}

	html := panel.Render().RenderHTML()

	if strings.Contains(html, "rp-aside") {
		t.Error("expected rp-aside to be absent when both AsideControls and Aside are nil")
	}
}

func TestRightPanel_RenderHTML_NoModuleNoID(t *testing.T) {
	panel := &rightpanel.RightPanel{Title: "No ID"}
	html := panel.Render().RenderHTML()

	if strings.Contains(html, "id=") {
		t.Error("expected no id attribute when Module is nil")
	}
}
