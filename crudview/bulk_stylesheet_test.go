//go:build !wasm

package crudview

import (
	"testing"

	"github.com/tinywasm/fmt"
	"github.com/tinywasm/html"
	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
	"github.com/tinywasm/view/conformance"
)

func TestActionGrowsInsteadOfFillingTheRow(t *testing.T) {
	caller := &conformance.FakeCaller{}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} })
	v := &CrudView{
		Title:     "CRUD",
		Presenter: p,
		Form:      html.Div(),
	}
	v.Init(&fakeCtx{})

	cssStr := v.RenderCSS().String()
	b := ruleBlock(cssStr, ".crudview__action {")

	if fmt.Contains(b, "width: 100%") || fmt.Contains(b, "width:100%") {
		t.Errorf(".crudview__action must grow instead of filling the row (width: 100%%), block:\n%s", b)
	}
	if !fmt.Contains(cssStr, "flex-grow") {
		t.Errorf("css must carry flex-grow (from Grow())")
	}
}

func TestFooterButtonsShareTheControlHeight(t *testing.T) {
	caller := &conformance.FakeCaller{}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} })
	v := &CrudView{
		Title:     "CRUD",
		Presenter: p,
		Form:      html.Div(),
	}
	v.Init(&fakeCtx{})

	cssStr := v.RenderCSS().String()

	for _, sel := range []string{".crudview__action {", ".crudview__action-delete {", ".crudview__action-edit {"} {
		b := ruleBlock(cssStr, sel)
		if b == "" {
			t.Fatalf("expected rule block for %s", sel)
		}
		if !fmt.Contains(b, "min-height: var(--control-height") {
			t.Errorf("footer button %s must be sized by --control-height, block:\n%s", sel, b)
		}
	}
}
