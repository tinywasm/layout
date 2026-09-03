//go:build !wasm

package crudview

import (
	"strings"
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

// The delete button leans red with the rows it will act on, and only there:
// the rule selects Within the footer's Open state, which holds exactly while
// deleting — normal mode stays Primary blue.
func TestDeleteButtonTurnsRedOnlyWhileDeleting(t *testing.T) {
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

	sel := `.crudview__footer[data-open="true"] .crudview__action-delete {`
	i := strings.Index(cssStr, sel)
	if i == -1 {
		t.Fatalf("expected the footer-tone rule %s", sel)
	}
	body := cssStr[i:]
	if end := strings.Index(body, "}"); end != -1 {
		body = body[:end]
	}
	if !fmt.Contains(body, "--color-danger-wash") {
		t.Errorf("the delete button must use the DangerWash surface while deleting, block:\n%s", body)
	}
}
