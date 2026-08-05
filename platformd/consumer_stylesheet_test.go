//go:build !wasm

package platformd

import (
	"strings"
	"testing"

	"github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	"github.com/tinywasm/html"
	"github.com/tinywasm/svg"
)

// testIdentity is the smallest thing that satisfies the Identity contract.
type testIdentity struct{}

func (testIdentity) UserName() string    { return "Tester" }
func (testIdentity) UserAvatar() string  { return "" }
func (testIdentity) UserRoles() []string { return []string{"QA", "Soporte"} }

// testBrand is the smallest thing that satisfies the Brand contract.
type testBrand struct {
	name string
	mark string
}

func (b testBrand) BrandName() string { return b.name }
func (b testBrand) BrandMark() string { return b.mark }

// ruleBlock returns the declaration block of the first rule whose selector
// contains want, or "" when absent.
func ruleBlock(cssStr, want string) string {
	i := strings.Index(cssStr, want)
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

// TestHoverOnANavLinkDoesNotResizeIt is the net that keeps the rail from
// flickering: the hover state of a nav item must paint an outline over the
// same pixel, never a border — a border adds 2px of layout and the item grows
// past --rail-narrow, leaves the pointer, loses :hover, and loops.
func TestHoverOnANavLinkDoesNotResizeIt(t *testing.T) {
	cssStr := (&Platform{}).RenderCSS().String()

	b := ruleBlock(cssStr, ".pd__nav-link:hover")
	if b == "" {
		t.Fatalf("expected a rule for .pd__nav-link:hover")
	}
	if Contains(b, "border:") {
		t.Errorf("hover must not paint a border (would resize the box), block:\n%s", b)
	}
	if !Contains(b, "outline:") {
		t.Errorf("hover must paint an outline, block:\n%s", b)
	}
	if !Contains(b, "outline-offset: -1px;") {
		t.Errorf("hover must use outline-offset: -1px, block:\n%s", b)
	}
}

func TestPlatform_StylesheetAsserts(t *testing.T) {
	// Identity, brand and the actions slot are all supplied: the class-parity
	// assertion below is two-directional, so anything left nil would read as a
	// stylesheet rule with no markup behind it.
	p := &Platform{
		AppName:     "Test App",
		Brand:       testBrand{name: "Acme", mark: "https://example.com/logo.svg"},
		User:        testIdentity{},
		UserActions: func() dom.Component { return html.Div().Text("actions") },
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

	// 0. Chrome-correctness assertions (PLAN v0.2.0)
	// Header and rail are welded to the frame: the part rules must not carry a
	// radius, and EdgeToEdge's border-radius: 0 must actually win over the
	// surface's default (the layer-ordering bug that left crudview at 4px).
	if b := ruleBlock(cssStr, ".pd__header {"); Contains(b, "border-radius") {
		t.Errorf("header must not carry a radius (EdgeToEdge squares it), block:\n%s", b)
	}
	if !Contains(cssStr, "border-radius: 0;") {
		t.Error("expected EdgeToEdge to emit border-radius: 0")
	}
	// The message block is centred, and the variants are tinted text with no
	// background: severity stays legible without a slab breaking the header.
	if b := ruleBlock(cssStr, ".pd__msg-slot {"); !Contains(b, "justify-content: center") {
		t.Errorf("msg-slot must centre its content, block:\n%s", b)
	}
	for part, wantColor := range map[string]string{
		".pd__msg-info {":    "--color-muted",
		".pd__msg-success {": "--color-success",
		".pd__msg-warning {": "--color-accent",
		".pd__msg-error {":   "--color-danger",
	} {
		b := ruleBlock(cssStr, part)
		if b == "" {
			t.Errorf("expected a rule for %s", part)
			continue
		}
		if Contains(b, "background-color") {
			t.Errorf("%s must not paint a background, block:\n%s", part, b)
		}
		if !Contains(b, "fill: currentColor") {
			t.Errorf("%s must be a Glyph (fill: currentColor), block:\n%s", part, b)
		}
		if !Contains(b, wantColor) {
			t.Errorf("%s should use %s, block:\n%s", part, wantColor, b)
		}
	}
	// The active nav item wears Primary (filled icon on dark bg) matching the
	// mobile hamburger; hover wears a tonal Inset shift so the two are never
	// confused.
	if b := ruleBlock(cssStr, `.pd__nav-link[data-current="true"] {`); !Contains(b, "--color-primary") {
		t.Errorf("current nav item must use Primary, block:\n%s", b)
	}
	if b := ruleBlock(cssStr, ".pd__nav-link:hover {"); !Contains(b, "--color-surface-sunken") {
		t.Errorf("nav hover must be the tonal Inset shift, block:\n%s", b)
	}
	if b := ruleBlock(cssStr, ".pd__nav-link:hover {"); Contains(b, "--color-primary") {
		t.Errorf("nav hover must not use Primary (indistinguishable from current), block:\n%s", b)
	}

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
		want := kv.Key + "='" + kv.Value + "'"
		if !Contains(allHTML, want) {
			t.Errorf("stylesheet selects on %q but no element in the markup ever writes it", want)
		}
	}
}
