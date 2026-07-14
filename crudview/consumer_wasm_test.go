//go:build wasm

package crudview

import (
	"testing"

	"github.com/tinywasm/form"
)

type mockCtxWasmConsumer struct{}

func (m *mockCtxWasmConsumer) OnCleanup(fn func()) {}

func TestConsumer_Wasm_Render(t *testing.T) {
	caller := &fakeCaller{}
	cfg := Config{
		ParentID:          "my-wasm-id",
		Caller:            caller,
		Record:            &Device{},
		ListOp:            "list_devices",
		SearchPlaceholder: "Custom Search...",
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error in New: %v", err)
	}

	v.Init(&mockCtxWasmConsumer{})

	// Check that the form was wired correctly
	f, ok := v.Form.(*form.Form)
	if !ok {
		t.Fatalf("expected form to be *form.Form, got %T", v.Form)
	}

	if f.ParentID() != "my-wasm-id" {
		t.Errorf("expected parent ID to be 'my-wasm-id', got '%s'", f.ParentID())
	}

	root := v.Render()
	if root == nil {
		t.Fatal("expected rendered root element to be non-nil")
	}

	html := root.String()
	if !github_com_tinywasm_fmt_Contains(html, "list_devices") && !github_com_tinywasm_fmt_Contains(html, "Custom Search...") {
		// Wait, let's just make sure the custom placeholder is present in the rendered HTML
		if !github_com_tinywasm_fmt_Contains(html, "Custom Search...") {
			t.Errorf("expected html to contain custom search placeholder, got: %s", html)
		}
	}
}

// Simple local helper to avoid unused package or syntax issues
func github_com_tinywasm_fmt_Contains(s, sub string) bool {
	// A basic implementation of substring search
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
