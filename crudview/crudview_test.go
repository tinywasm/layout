//go:build !wasm

package crudview

import (
	"testing"

	. "github.com/tinywasm/fmt"
	"github.com/tinywasm/fmt/lang"
	. "github.com/tinywasm/html"
	"github.com/tinywasm/view"
	"github.com/tinywasm/view/conformance"
)

func TestCrudView_Render_Basic(t *testing.T) {
	v := &CrudView{
		Title: "Test CRUD",
		Form:  Div().Text("The Form"),
	}
	v.Init(&mockCtx{})

	html := v.Render().String()

	if !Contains(html, "Test CRUD") {
		t.Error("expected title")
	}
	if !Contains(html, "The Form") {
		t.Error("expected form")
	}
	// No presenter → no source → rightpanel omits the aside band entirely.
	if Contains(html, "rp__aside") {
		t.Error("did not expect aside without source")
	}
}

func TestCrudView_Render_WithSource(t *testing.T) {
	fb := &conformance.FakeBackend{}
	p := view.New(fb, &Device{})

	v := &CrudView{
		Title:     "CRUD with List",
		Presenter: p,
		OnNew:     func() {},
	}
	v.Init(&mockCtx{})

	html := v.Render().String()

	if !Contains(html, "class='rp'") {
		t.Error("expected the rightpanel skeleton")
	}
	if !Contains(html, "rp__aside") {
		t.Error("expected the aside band with a source")
	}
	if !Contains(html, "name='cv-crudtoggle'") {
		t.Error("expected the single crud toggle button (+/↺)")
	}
	// Delete/Edit no longer render as buttons here — they live in each
	// targetlist row's ⋮ menu (Editar/Eliminar).
	if Contains(html, "name='btn_cruddel'") {
		t.Error("did not expect a separate delete button")
	}
	// There is no back button: on a phone a sliver of the list stays visible
	// at the trailing edge of the swipe strip, and cancelling returns there
	// on its own (undoAction).
	if Contains(html, "name='cv-back'") {
		t.Error("did not expect a back-to-list button")
	}
	// The aside carries the scroll-snap target id rightpanel stamps — the id
	// the delegated ShowAside() resolves on a phone.
	if !Contains(html, ".aside'") {
		t.Error("expected the aside to carry the scroll-snap target id")
	}
}

type mockCtx struct{}

func (m *mockCtx) OnCleanup(fn func()) {}

// TestCrudView_DeleteConfirm_Language: the delete-confirmation dialog is
// framework-owned chrome, so its text is a lang.Translate call — resolved
// from the English canonical keys, never hardcoded. The dictionary is the
// consumer's concern (registered here, in the test — never in production
// code), mirroring tinywasm/input's lazy-resolution test.
func TestCrudView_DeleteConfirm_Language(t *testing.T) {
	openDialog := func(v *CrudView) string {
		v.confirmDelete.Init(&mockCtx{})
		v.deleteLabel.Set("Device One")
		v.confirmDelete.Open()
		// Rendered once per instance: the dialog's content element can only
		// be attached once (one element, one parent).
		return v.confirmDelete.Render().String()
	}

	// English is the canonical dictionary key and the default output
	// language: the dialog renders in EN until a consumer registers a
	// dictionary and activates another language via lang.OutLang.
	v := &CrudView{Title: "Test CRUD"}
	v.Init(&mockCtx{})
	html := openDialog(v)
	for _, want := range []string{
		"Confirm", "Cancel", "Delete",
		"Delete «Device One»? This action cannot be undone.",
	} {
		if !Contains(html, want) {
			t.Errorf("expected EN dialog text %q, got: %s", want, html)
		}
	}

	// With the ES dictionary registered and active BEFORE the view is built,
	// the same chrome resolves the Spanish text — title included. Every word
	// is its own entry, so words stay independent and reusable ("Delete" is
	// shared between the message and the confirm button).
	lang.RegisterWords([]lang.DictEntry{
		{EN: "Confirm", ES: "Confirmar"},
		{EN: "Cancel", ES: "Cancelar"},
		{EN: "Delete", ES: "Eliminar"},
		{EN: "This", ES: "Esta"},
		{EN: "action", ES: "acción"},
		{EN: "cannot", ES: "no"},
		{EN: "be", ES: "se"},
		{EN: "undone.", ES: "puede deshacer."},
	})
	defer lang.OutLang(lang.EN)
	lang.OutLang(lang.ES)

	v2 := &CrudView{Title: "Test CRUD"}
	v2.Init(&mockCtx{})
	html = openDialog(v2)
	for _, want := range []string{
		"Confirmar", "Cancelar", "Eliminar",
		"Eliminar «Device One»? Esta acción no se puede deshacer.",
	} {
		if !Contains(html, want) {
			t.Errorf("expected ES dialog text %q, got: %s", want, html)
		}
	}
}