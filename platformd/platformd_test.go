//go:build !wasm

package platformd

import (
	"testing"
	"time"

	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	. "github.com/tinywasm/html"
)

func TestPlatform_Render(t *testing.T) {
	p := &Platform{
		AppName: "Test App",
		Modules: []Module{
			{ID: "mod1", Label: "Module 1"},
			{ID: "mod2", Label: "Module 2"},
		},
	}
	p.Init(NilCtx())

	el := p.Render()
	html := el.String()

	if html == "" {
		t.Fatal("String() returned empty string")
	}

	expected := []string{
		"pd-root",
		"pd-header",
		"pd-menu",
		"pd-stage",
		"pd-panel",
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
		Element: *Div(),
		Modules: []Module{
			{ID: "mod1", Label: "Mod 1"},
			{ID: "mod2", Label: "Mod 2", Default: true},
		},
	}
	p.Init(NilCtx()) // Should activate mod2

	html := p.Render().String()
	// tinywasm/dom uses single quotes for attributes
	if !contains(html, "id='mod2' class='pd-panel pd-panel-active'") {
		t.Errorf("expected mod2 to be active, got HTML: %s", html)
	}
}

func TestPlatform_Activate(t *testing.T) {
	p := &Platform{
		Element: *Div(),
		Modules: []Module{
			{ID: "mod1", Label: "Mod 1"},
			{ID: "mod2", Label: "Mod 2"},
		},
	}
	p.Init(NilCtx())
	p.Activate("mod2")

	html := p.Render().String()
	if !contains(html, "id='mod2' class='pd-panel pd-panel-active'") {
		t.Errorf("expected mod2 to be active after Activate('mod2'), got HTML: %s", html)
	}
	if contains(html, "id='mod1' class='pd-panel pd-panel-active'") {
		t.Errorf("expected mod1 to NOT be active")
	}
}

func TestPlatform_Notify_Renders(t *testing.T) {
	p := &Platform{Element: *Div()}
	p.Init(NilCtx())
	p.Notify(Msg.Error, "boom", 0)

	html := p.Render().String()
	t.Logf("HTML: %s", html)

	// Desktop slot
	if !contains(html, "id='pd-msg-desktop'") {
		t.Error("expected pd-msg-desktop")
	}
	if !contains(html, "pd-msg-error") {
		t.Error("expected pd-msg-error")
	}
	if !contains(html, "boom") {
		t.Error("expected boom")
	}

	// Mobile slot
	if !contains(html, "id='pd-msg-mobile'") {
		t.Error("expected pd-msg-mobile")
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
	if !contains(css, ".pd-root") {
		t.Errorf("expected CSS to contain .pd-root")
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
