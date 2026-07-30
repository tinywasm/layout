//go:build !wasm

package crudview

import (
	"strings"
	"testing"

	"github.com/tinywasm/fmt"
	"github.com/tinywasm/html"
	"github.com/tinywasm/view"
	"github.com/tinywasm/view/conformance"
	"github.com/tinywasm/model"
)

func TestConsumer_StylesheetAsserts(t *testing.T) {
	caller := &conformance.FakeCaller{}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} })
	v := &CrudView{
		Title:     "CRUD",
		Presenter: p,
		Form:      html.Div(),
	}
	v.Init(&fakeCtx{})
	_ = v.Reload()
	v.composing.Set(true) // Ensure data-open renders as true

	cssStr := v.RenderCSS().String()

	// 1. Check "!important"
	if fmt.Contains(cssStr, "!important") {
		t.Error("stylesheet contains forbidden !important directive")
	}

	// 2. Check @layer order
	expectedLayers := "@layer tokens, primitives, widgets, states;"
	if !fmt.Contains(cssStr, expectedLayers) {
		t.Errorf("expected stylesheet to declare layers in order %q", expectedLayers)
	}

	// 3. Extract markup and match classes starting with "crudview"
	html1 := v.Render().String()

	vFull := &CrudView{Title: "Full", Form: html.Div()}
	vFull.Init(&fakeCtx{})
	html2 := vFull.Render().String()

	// Manually render delete confirmation content because ModalDialog doesn't render it in hidden/initial state
	confirmHTML := v.renderDeleteConfirm().String()

	allHTML := html1 + " " + html2 + " " + confirmHTML

	// Extract classes from HTML
	importMatches := func() map[string]bool {
		classes := make(map[string]bool)
		// We look for class='...' attributes
		// HTML strings use single quotes for attributes in tinywasm/dom.
		for i := 0; i < len(allHTML); i++ {
			if strings.HasPrefix(allHTML[i:], "class='") {
				start := i + len("class='")
				end := start
				for end < len(allHTML) && allHTML[end] != '\'' {
					end++
				}
				clsGroup := allHTML[start:end]
				// split by space in case of multiple classes
				for _, cls := range strings.Fields(clsGroup) {
					if strings.HasPrefix(cls, "crudview") {
						classes[cls] = true
					}
				}
			}
		}
		return classes
	}()

	// Extract classes from CSS
	cssClasses := func() map[string]bool {
		classes := make(map[string]bool)
		// We search for class selectors: e.g. .crudview or .crudview__something
		for i := 0; i < len(cssStr); i++ {
			if cssStr[i] == '.' {
				start := i + 1
				end := start
				for end < len(cssStr) && ((cssStr[end] >= 'a' && cssStr[end] <= 'z') ||
					(cssStr[end] >= 'A' && cssStr[end] <= 'Z') ||
					(cssStr[end] >= '0' && cssStr[end] <= '9') ||
					cssStr[end] == '-' || cssStr[end] == '_') {
					end++
				}
				cls := cssStr[start:end]
				if strings.HasPrefix(cls, "crudview") {
					classes[cls] = true
				}
			}
		}
		return classes
	}()

	// Verify all markup classes exist in CSS
	for c := range importMatches {
		if !cssClasses[c] {
			t.Errorf("class %q present in markup but missing in stylesheet", c)
		}
	}

	// Verify all CSS classes exist in markup
	for c := range cssClasses {
		if !importMatches[c] {
			t.Errorf("class %q present in stylesheet but missing in markup", c)
		}
	}

	// 4. State-attribute parity
	sheet := v.RenderSheet()
	for _, kv := range sheet.StateAttrs() {
		want := kv.Key + "='" + kv.Value + "'"
		if !fmt.Contains(allHTML, want) {
			t.Errorf("stylesheet selects on %q but no element in the markup ever writes it", want)
		}
	}
}
