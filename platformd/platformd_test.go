//go:build !wasm

package platformd

import (
	"testing"
	"time"

	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	. "github.com/tinywasm/html"
	"github.com/tinywasm/svg"
)

type mockModule struct {
	id    string
	label string
	icon  svg.Icon
	view  Component
}

func (m *mockModule) ModelName() string { return m.id }
func (m *mockModule) Label() string     { return m.label }
func (m *mockModule) Icon() svg.Icon    { return m.icon }
func (m *mockModule) View() Component   { return m.view }

func TestPlatform_Render(t *testing.T) {
	p := &Platform{
		AppName: "Test App",
		Modules: []UIModule{
			&mockModule{id: "mod1", label: "Module 1"},
			&mockModule{id: "mod2", label: "Module 2"},
		},
	}
	p.Init(NilCtx())

	el := p.Render()
	html := el.String()

	if html == "" {
		t.Fatal("String() returned empty string")
	}

	expected := []string{
		"pd",
		"pd__header",
		"pd__menu",
		"pd__stage",
		"pd__panel",
		"mod1",
		"mod2",
	}

	for _, s := range expected {
		if !contains(html, s) {
			t.Errorf("expected HTML to contain %q", s)
		}
	}
}

func TestPlatform_Render_DefaultModule(t *testing.T) {
	p := &Platform{
		Element:   *Div(),
		DefaultID: "mod2",
		Modules: []UIModule{
			&mockModule{id: "mod1", label: "Mod 1"},
			&mockModule{id: "mod2", label: "Mod 2"},
		},
	}
	p.Init(NilCtx()) // Should activate mod2

	html := p.Render().String()
	// tinywasm/dom uses single quotes for attributes and boolean attributes are empty keys
	if !contains(html, "id='mod2' class='pd__panel' data-id='mod2' data-current=''") {
		t.Errorf("expected mod2 to be active, got HTML: %s", html)
	}
}

func TestPlatform_Activate(t *testing.T) {
	p := &Platform{
		Element: *Div(),
		Modules: []UIModule{
			&mockModule{id: "mod1", label: "Mod 1"},
			&mockModule{id: "mod2", label: "Mod 2"},
		},
	}
	p.Init(NilCtx())
	p.Activate("mod2")

	html := p.Render().String()
	if !contains(html, "id='mod2' class='pd__panel' data-id='mod2' data-current=''") {
		t.Errorf("expected mod2 to be active after Activate('mod2'), got HTML: %s", html)
	}
	if contains(html, "id='mod1' class='pd__panel' data-id='mod1' data-current=''") {
		t.Errorf("expected mod1 to NOT be active")
	}
}

func TestPlatform_NewUIModule(t *testing.T) {
	m := NewUIModule("test", "Label", svg.Icon("icon"), Div())
	if m.ModelName() != "test" {
		t.Errorf("expected test, got %s", m.ModelName())
	}
	if m.Label() != "Label" {
		t.Errorf("expected Label, got %s", m.Label())
	}
	if m.Icon() != svg.Icon("icon") {
		t.Errorf("expected icon, got %s", m.Icon())
	}
	if m.View() == nil {
		t.Error("expected view to be non-nil")
	}
}

func TestPlatform_CanView(t *testing.T) {
	p := &Platform{
		Modules: []UIModule{
			&mockModule{id: "mod1", label: "Mod 1"},
			&mockModule{id: "mod2", label: "Mod 2"},
		},
		CanView: func(id string) bool {
			return id == "mod2"
		},
	}
	p.Init(NilCtx())

	html := p.Render().String()

	// Nav rail should only have mod2
	if contains(html, "data-id='mod1'") {
		t.Error("expected mod1 link NOT to be rendered")
	}
	if !contains(html, "data-id='mod2'") {
		t.Error("expected mod2 link to be rendered")
	}

	// Stage should only have mod2 panel
	if contains(html, "id='mod1'") {
		t.Error("expected mod1 panel NOT to be rendered")
	}
	if !contains(html, "id='mod2'") {
		t.Error("expected mod2 panel to be rendered")
	}

	// mod2 should be active (fallback)
	if !contains(html, "id='mod2' class='pd__panel' data-id='mod2' data-current=''") {
		t.Error("expected mod2 to be active")
	}

	// Try activating mod1
	p.Activate("mod1")
	if p.active.Get() == "mod1" {
		t.Error("Activate should not set active to non-viewable module")
	}
}

func TestPlatform_Notify_Renders(t *testing.T) {
	p := &Platform{Element: *Div()}
	p.Init(NilCtx())
	p.Notify(Msg.Error, "boom", 0)

	html := p.Render().String()
	t.Logf("HTML: %s", html)

	// Unified slot
	if !contains(html, "id='pd-msg-slot'") {
		t.Error("expected pd-msg-slot")
	}
	if !contains(html, "pd__msg-error") {
		t.Error("expected pd__msg-error")
	}
	if !contains(html, "boom") {
		t.Error("expected boom")
	}
}

func TestPlatform_Notify_Dismiss(t *testing.T) {
	p := &Platform{Element: *Div()}
	p.Init(NilCtx())
	p.Notify(Msg.Info, "hi", 10) // 10ms

	if p.notificationCount() != 1 {
		t.Fatalf("expected 1 notification, got %d", p.notificationCount())
	}

	// Wait for dismissal
	time.Sleep(100 * time.Millisecond)

	if p.notificationCount() != 0 {
		t.Errorf("expected 0 notifications after dismissal, got %d", p.notificationCount())
	}
}

func TestRenderCSS_NonEmpty(t *testing.T) {
	p := &Platform{}
	css := p.RenderCSS().String()
	if css == "" {
		t.Fatal("RenderCSS() returned empty string")
	}
	if !contains(css, ".pd") {
		t.Errorf("expected CSS to contain .pd")
	}
}

type mockCtx struct{}

func (m *mockCtx) OnCleanup(fn func()) {}

func NilCtx() Ctx {
	return &mockCtx{}
}

func contains(s, substr string) bool {
	// Simple manual contains to avoid strings package as per project rules
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
