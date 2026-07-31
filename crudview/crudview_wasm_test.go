//go:build wasm

package crudview

import (
	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
	"github.com/tinywasm/view/conformance"
	"testing"
)

func TestCrudView_Wasm_Flow(t *testing.T) {
	caller := &conformance.FakeCaller{
		Reply: func(op string, into model.Decodable) {
			dl := into.(*DeviceList)
			d1 := dl.Append().(*Device)
			d1.Id = "1"
			d1.Name = "Item One"
			d1.Ip = "Desc One"

			d2 := dl.Append().(*Device)
			d2.Id = "2"
			d2.Name = "Item Two"
			d2.Ip = "Desc Two"
		},
	}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} },
		view.WithTitle("Wasm Test"),
		view.WithDeleteOp("device_delete"))

	v := &CrudView{
		Title:     "Wasm Test",
		Presenter: p,
	}
	v.Init(&mockCtxWasm{})
	_ = v.Reload()

	if len(v.Presenter.Items()) != 2 {
		t.Errorf("expected 2 items, got %d", len(v.Presenter.Items()))
	}

	// Test filtering
	v.search.Set("one")
	v.filter()
	if v.list.Count() != 1 {
		t.Errorf("expected 1 filtered item, got %d", v.list.Count())
	}

	// Test selection
	v.selectAction(view.Item{ID: "1"})
	if v.selected.Get() != "1" {
		t.Error("expected item 1 to be selected")
	}
	if !v.canDelete.Get() {
		t.Error("expected canDelete to be true")
	}

	v.selectAction(view.Item{ID: ""})
	if v.canDelete.Get() {
		t.Error("expected canDelete to be false")
	}
}

type mockCtxWasm struct{}

func (m *mockCtxWasm) OnCleanup(fn func()) {}
