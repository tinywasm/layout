package platformd

import (
	"testing"
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
