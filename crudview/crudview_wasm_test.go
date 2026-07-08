//go:build wasm

package crudview

import (
	"testing"
	"github.com/tinywasm/model"
)

func TestCrudView_Wasm_Flow(t *testing.T) {
	mock := &mockCaller{}
	v := &CrudView{
		Title: "Wasm Test",
		Source: Source{
			Caller: mock,
			ListOp: "list",
			Decode: func(raw []byte) ([]Item, error) {
				return []Item{
					{ID: "1", Label: "Item One", Description: "Desc One"},
					{ID: "2", Label: "Item Two", Description: "Desc Two"},
				}, nil
			},
		},
	}
	v.Init(&mockCtxWasm{})

	if len(v.allItems) != 2 {
		t.Errorf("expected 2 items, got %d", len(v.allItems))
	}

	// Test filtering
	v.search.Set("one")
	v.filter()
	if len(v.items.Get()) != 1 {
		t.Errorf("expected 1 filtered item, got %d", len(v.items.Get()))
	}

	// Test selection
	v.Select("1")
	if v.selected.Get() != "1" {
		t.Error("expected item 1 to be selected")
	}
	if !v.canDelete.Get() {
		t.Error("expected canDelete to be true")
	}

	v.Select("")
	if v.canDelete.Get() {
		t.Error("expected canDelete to be false")
	}
}

type mockCaller struct {
	calls int
}

func (c *mockCaller) Call(op string, args model.Encodable) ([]byte, error) {
	c.calls++
	return nil, nil
}

type mockCtxWasm struct{}
func (m *mockCtxWasm) OnCleanup(fn func()) {}
