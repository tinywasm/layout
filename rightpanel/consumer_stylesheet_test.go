//go:build !wasm

package rightpanel

import (
	"strings"
	"testing"

	"github.com/tinywasm/fmt"
	"github.com/tinywasm/html"
)

type mockModule struct {
	id string
}

func (m *mockModule) ModelName() string { return m.id }

// ruleBlock returns the declaration block of the LAST rule whose selector
// contains want, or "" when absent. The primitives layer is emitted before the
// widgets layer and groups edge cases (e.g. ".rp, .rp__header, .rp__title {
// margin: 0; border-radius: 0; }"), so the first match can be a legitimate
// collective primitive — the intent lives in the widgets-layer rule.
func ruleBlock(cssStr, want string) string {
	i := strings.LastIndex(cssStr, want)
	if i == -1 {
		return ""
	}
	start := strings.Index(cssStr[i:], "{")
	if start == -1 {
		return ""
	}
	body := cssStr[i+start:]
	end := strings.Index(body, "}")
	if end == -1 {
		return ""
	}
	return body[:end]
}

// TestRightPanel_EdgeAsserts guards PLAN v0.2.0 item 5: the panel is welded to
// the application frame and must be square, while interior parts keep their
// radius. The aside-header is the deliberate exception: it sheds every
// treatment (the consumer's control brings its own), so it carries no radius.
func TestRightPanel_EdgeAsserts(t *testing.T) {
	r := &RightPanel{
		Module:        &mockModule{id: "test-module"},
		Title:         "Test Title",
		Head:          html.Div(),
		HeadControls:  html.Div(),
		Article:       html.Div(),
		AsideControls: html.Div(),
		Aside:         html.Div(),
	}

	cssStr := r.RenderCSS().String()

	// The mobile strip re-declares the root and the title inside a media query
	// that comes after the desktop rules, so LastIndex would land there. The
	// frame's square corners and the title's radius are desktop contracts;
	// assert them on everything before the first @media block.
	desktop := cssStr
	if i := strings.Index(cssStr, "@media"); i != -1 {
		desktop = cssStr[:i]
	}

	for _, sel := range []string{".rp {"} {
		if b := ruleBlock(desktop, sel); fmt.Contains(b, "border-radius") {
			t.Errorf("%s must be squared at the frame, block:\n%s", sel, b)
		}
	}
	if b := ruleBlock(desktop, ".rp__title {"); !fmt.Contains(b, "border-radius") {
		t.Errorf("interior title must keep its radius, block:\n%s", b)
	}
}

func TestRightPanel_StylesheetAsserts(t *testing.T) {
	r := &RightPanel{
		Module:        &mockModule{id: "test-module"},
		Title:         "Test Title",
		Head:          html.Div(),
		HeadControls:  html.Div(),
		Article:       html.Div(),
		AsideControls: html.Div(),
		Aside:         html.Div(),
		AsideFooter:   html.Div(),
	}

	sheet := r.RenderSheet()
	cssStr := r.RenderCSS().String()

	// 1. Check "!important"
	if fmt.Contains(cssStr, "!important") {
		t.Error("stylesheet contains forbidden !important directive")
	}

	// 2. Check @layer order
	expectedLayers := "@layer tokens, primitives, widgets, states;"
	if !fmt.Contains(cssStr, expectedLayers) {
		t.Errorf("expected stylesheet to declare layers in order %q", expectedLayers)
	}

	// 3. Extract markup and match classes starting with "rp"
	allHTML := r.Render().String()

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
					if strings.HasPrefix(cls, "rp") {
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
		// We search for class selectors: e.g. .rp or .rp__something
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
				if strings.HasPrefix(cls, "rp") {
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
		want := kv.Key + "='" + kv.Value + "'"
		if !fmt.Contains(allHTML, want) {
			t.Errorf("stylesheet selects on %q but no element in the markup ever writes it", want)
		}
	}
}

func TestRightPanel_FlowIsSplitRootStackedMain(t *testing.T) {
	r := &RightPanel{
		Module:        &mockModule{id: "test-module"},
		Title:         "Test Title",
		Article:       html.Div(),
		AsideControls: html.Div(),
		Aside:         html.Div(),
	}

	cssStr := r.RenderCSS().String()

	// The root splits the two columns; main stacks the title band above the body.
	// These were swapped, and the swap was invisible because no demo module
	// passed an Aside. ruleBlock's LastIndex would land on the mobile strip's
	// re-flow of the root, so the desktop flow is asserted on the exact rule
	// the style DSL emits for the Split/Stack pair. Main's stack is part of the
	// primitives-layer compound selector it shares with the aside bands.
	if !fmt.Contains(cssStr, ".rp {\n  display: flex;\n  flex-wrap: wrap") {
		t.Errorf(".rp must carry the Split flow")
	}
	if !fmt.Contains(cssStr, ".rp__aside, .rp__aside-content, .rp__main {\n  display: flex;\n  flex-direction: column") {
		t.Errorf(".rp__main must stack, not split")
	}
}
