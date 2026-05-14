package platformd

import (
	"testing"
	"time"

	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
)

func TestPlatform_Render(t *testing.T) {
	p := &Platform{
		AppName: "Test App",
		Modules: []Module{
			{ID: "mod1", Label: "Module 1"},
			{ID: "mod2", Label: "Module 2"},
		},
	}

	el := p.Render()
	html := el.RenderHTML()

	if html == "" {
		t.Fatal("RenderHTML() returned empty string")
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
		Element: Div(),
		Modules: []Module{
			{ID: "mod1", Label: "Mod 1"},
			{ID: "mod2", Label: "Mod 2", Default: true},
		},
	}
	p.OnMount() // Should activate mod2

	html := p.Render().RenderHTML()
	// tinywasm/dom uses single quotes for attributes
	if !contains(html, "id='mod2' class='pd-panel pd-panel-active'") {
		t.Errorf("expected mod2 to be active, got HTML: %s", html)
	}
}

func TestPlatform_Activate(t *testing.T) {
	p := &Platform{
		Element: Div(),
		Modules: []Module{
			{ID: "mod1", Label: "Mod 1"},
			{ID: "mod2", Label: "Mod 2"},
		},
	}
	p.Activate("mod2")

	html := p.Render().RenderHTML()
	if !contains(html, "id='mod2' class='pd-panel pd-panel-active'") {
		t.Errorf("expected mod2 to be active after Activate('mod2'), got HTML: %s", html)
	}
	if contains(html, "id='mod1' class='pd-panel pd-panel-active'") {
		t.Errorf("expected mod1 to NOT be active")
	}
}

func TestPlatform_Notify_Renders(t *testing.T) {
	p := &Platform{Element: Div()}
	p.Notify(Msg.Error, "boom", 0)

	html := p.Render().RenderHTML()

	// Desktop slot
	if !contains(html, "id='pd-msg-desktop'") || !contains(html, "pd-msg-error") || !contains(html, "boom") {
		t.Error("expected notification in desktop slot")
	}

	// Mobile slot
	if !contains(html, "id='pd-msg-mobile'") || !contains(html, "pd-msg-error") || !contains(html, "boom") {
		t.Error("expected notification in mobile slot")
	}
}

func TestPlatform_Notify_Dismiss(t *testing.T) {
	p := &Platform{Element: Div()}
	p.Notify(Msg.Info, "hi", 10) // 10ms

	if len(p.notifications) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(p.notifications))
	}

	// Wait for dismissal
	time.Sleep(100 * time.Millisecond)

	if len(p.notifications) != 0 {
		t.Errorf("expected 0 notifications after dismissal, got %d", len(p.notifications))
	}
}

func TestRenderCSS_NonEmpty(t *testing.T) {
	p := SSRInstance()
	css := p.RenderCSS().String()
	if css == "" {
		t.Fatal("RenderCSS() returned empty string")
	}
	if !contains(css, ".pd-root") {
		t.Errorf("expected CSS to contain .pd-root")
	}
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
