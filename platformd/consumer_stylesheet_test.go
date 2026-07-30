//go:build !wasm

package platformd

import (
	"strings"
	"testing"

	"github.com/tinywasm/svg"
	. "github.com/tinywasm/fmt"
)

func TestPlatform_StylesheetAsserts(t *testing.T) {
	p := &Platform{
		AppName: "Test App",
		Modules: []UIModule{
			&mockModule{id: "mod1", label: "Module 1", icon: svg.Icon("home")},
			&mockModule{id: "mod2", label: "Module 2", icon: svg.Icon("info")},
		},
	}
	p.Init(NilCtx())

	// Queue notifications of all types to ensure variant classes render
	p.Notify(Msg.Info, "info msg", 0)
	p.Notify(Msg.Success, "success msg", 0)
	p.Notify(Msg.Warning, "warning msg", 0)
	p.Notify(Msg.Error, "error msg", 0)

	// Set menuOpen to true so that data-open attribute renders in markup
	p.menuOpen.Set(true)

	sheet := p.RenderSheet()
	cssStr := p.RenderCSS().String()

	// 1. Check "!important"
	if Contains(cssStr, "!important") {
		t.Error("stylesheet contains forbidden !important directive")
	}

	// 2. Check @layer order
	expectedLayers := "@layer tokens, primitives, widgets, states;"
	if !Contains(cssStr, expectedLayers) {
		t.Errorf("expected stylesheet to declare layers in order %q", expectedLayers)
	}

	// 3. Extract markup and match classes starting with "pd"
	allHTML := p.Render().String()

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
					if strings.HasPrefix(cls, "pd") {
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
		// We search for class selectors: e.g. .pd or .pd__something
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
				if strings.HasPrefix(cls, "pd") {
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
	for _, kv := range sheet.StateAttrs() {
		attrKey := kv.Key
		if !Contains(allHTML, attrKey) {
			t.Errorf("stylesheet selects on %q but no element in the markup ever writes it", attrKey)
		}
	}
}
