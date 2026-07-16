//go:build wasm

package crudview

import (
	"testing"

	"github.com/tinywasm/form"
	"github.com/tinywasm/fmt"
)

type mockCtxWasmConsumer struct{}

func (m *mockCtxWasmConsumer) OnCleanup(fn func()) {}

func TestConsumer_Wasm_Render(t *testing.T) {
	p := &fakePresenter{
		title: "Custom Search Title",
		record: &Device{},
		searchPlaceholder: "Custom Search...",
	}
	cfg := Config{
		ParentID:          "my-wasm-id",
		Presenter:         p,
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
	if !fmt.Contains(html, "Custom Search...") {
		t.Errorf("expected html to contain custom search placeholder, got: %s", html)
	}
}
