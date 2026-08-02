//go:build !wasm

package crudview

import (
	"testing"

	. "github.com/tinywasm/fmt"
	. "github.com/tinywasm/html"
	"github.com/tinywasm/model"
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
	caller := &conformance.FakeCaller{}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} })

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
