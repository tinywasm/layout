//go:build !wasm

package platformd

import (
	"strings"
	"testing"

	"github.com/tinywasm/css"
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
// flickering: the hover state of a nav item must never add a border, which
// would grow the box past --rail-narrow, push it out from under the pointer,
// drop :hover, and loop. AccentHover (see platformd/css.go) carries no border
// or outline at all — only background-color and color — so there is nothing
// left that could resize the box. The outline-instead-of-border requirement
// this test used to enforce was specific to Inset, which AccentHover replaced
// as the nav-link hover treatment; a surface with no border property in the
// first place satisfies the same invariant more simply than working around
// one.
func TestHoverOnANavLinkDoesNotResizeIt(t *testing.T) {
	cssStr := (&Platform{}).RenderCSS().String()

	b := ruleBlock(cssStr, ".pd__nav-link:hover")
	if b == "" {
		t.Fatalf("expected a rule for .pd__nav-link:hover")
	}
	if Contains(b, "border:") || Contains(b, "outline:") {
		t.Errorf("hover must not paint a border or outline (would resize the box), block:\n%s", b)
	}
}

// TestDrawerPanelFloatIsFinePointerScoped is the net that keeps the double menu
// off a phone: the floating drawer-panel must live inside
// `@media (hover: hover)` and never in the plain states layer. A touch tap
// fires :hover but never the fine-pointer capability, so a float emitted
// without that gate activates on tap and duplicates inside the Drawer.
func TestDrawerPanelFloatIsFinePointerScoped(t *testing.T) {
	cssStr := (&Platform{}).RenderCSS().String()

	mediaIdx := strings.Index(cssStr, "@media (hover: hover)")
	if mediaIdx < 0 {
		t.Fatalf("expected the fine-pointer gate, got:\n%s", cssStr)
	}

	block := ruleBlock(cssStr[mediaIdx:], ".pd__menu:hover .pd__drawer-panel")
	if block == "" {
		t.Fatalf("expected the floating drawer-panel rule inside the hover media query, got:\n%s", cssStr[mediaIdx:])
	}
	if !Contains(block, "position: absolute;") || !Contains(block, "box-shadow:") {
		t.Errorf("expected the float (Docked + Raise) inside the gate, block:\n%s", block)
	}

	if b := ruleBlock(cssStr[:mediaIdx], ".pd__menu:hover .pd__drawer-panel"); b != "" {
		t.Errorf("the float must not escape into the plain states layer, block:\n%s", b)
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

	// Queue notifications of all types so the class-parity check below sees
	// every toast shape that Notify can produce. Persistent: no timers
	// armed, nothing can fire mid-test.
	p.Notify(Msg.Info, "info msg", Persistent())
	p.Notify(Msg.Success, "success msg", Persistent())
	p.Notify(Msg.Warning, "warning msg", Persistent())
	p.Notify(Msg.Error, "error msg", Persistent())

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
	// The message block is centred.
	if b := ruleBlock(cssStr, ".pd__msg-slot {"); !Contains(b, "justify-content: center") {
		t.Errorf("msg-slot must centre its content, block:\n%s", b)
	}
	// The active nav item wears AccentInverse (white icon on the fully
	// committed amber fill), matching the mobile hamburger and crudview's
	// open action -- one "current" language across the whole chassis
	// instead of the rail alone staying on Primary blue. Hover wears
	// AccentHover: the same white icon on a weaker (70%) amber, so the two
	// states are told apart by intensity rather than by icon color --
	// AccentWash's 85%-faded tint was tried and rejected here specifically
	// because it fades too close to the page background for white to read.
	if b := ruleBlock(cssStr, `.pd__nav-link[data-current="true"] {`); !Contains(b, css.ColorAccent.LightValue()) || !Contains(b, css.ColorOnPrimary.LightValue()) {
		t.Errorf("current nav item must use Accent bg + white (ColorOnPrimary) icon, block:\n%s", b)
	}
	if b := ruleBlock(cssStr, ".pd__nav-link:hover {"); !Contains(b, "color-mix(") || !Contains(b, css.ColorAccentHover.LightValue()) {
		t.Errorf("nav hover must be the weaker AccentHover tint, block:\n%s", b)
	}
	if b := ruleBlock(cssStr, ".pd__nav-link:hover {"); !Contains(b, css.ColorOnPrimary.LightValue()) {
		t.Errorf("nav hover must keep the same white icon as current, block:\n%s", b)
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

	// pd__app-name is Brand's mutually-exclusive fallback (see Platform.Render):
	// with a Brand set, as p has above, AppName never reaches markup, but its
	// rule is still in cssStr — RenderSheet() is one fixed stylesheet for the
	// widget type, not tailored to this instance's fields. A second render
	// with no Brand is what actually exercises it, needed for the
	// stylesheet-to-markup direction of the class-parity check below.
	pNoBrand := &Platform{
		AppName: "Test App",
		User:    testIdentity{},
		Modules: p.Modules,
	}
	pNoBrand.Init(NilCtx())
	allHTML += pNoBrand.Render().String()

	// Extract classes from HTML
	extractClasses := func(html string) map[string]bool {
		classes := make(map[string]bool)
		// We look for class='...' attributes
		// HTML strings use single quotes for attributes in tinywasm/dom.
		for i := 0; i < len(html); i++ {
			if strings.HasPrefix(html[i:], "class='") {
				start := i + len("class='")
				end := start
				for end < len(html) && html[end] != '\'' {
					end++
				}
				clsGroup := html[start:end]
				// split by space in case of multiple classes
				for _, cls := range strings.Fields(clsGroup) {
					if strings.HasPrefix(cls, "pd") {
						classes[cls] = true
					}
				}
			}
		}
		return classes
	}
	importMatches := extractClasses(allHTML)

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

// TestMessageColorHasOnlyOneSource is the regression net for a reported bug:
// a success toast read green on desktop and black on mobile, and the delete-
// confirmation dialog was a third color again, because each surface declared
// its own. The mobile msg rule (On(css.Mobile, "msg", As(...))) and four
// severity variants (Part("msg-info"/"-success"/"-warning"/"-error"),
// Glyph(...)) had equal CSS specificity; since @media rules are grouped and
// emitted after the base rules regardless of where they are declared in Go,
// the mobile color always won — every severity read as the same near-black
// on a phone while desktop still showed its tint.
//
// The fix collapsed the four variants into one identity (the system's own
// Primary blue) instead of reconciling their specificity: a plain-text toast
// with no icon had no second channel carrying severity to sighted users
// anyway, so color was the entire signal, and it was already silently broken
// for every phone. role="status"/"alert" (toastNodes in platformd.go) is what
// actually carries severity to assistive tech, unaffected by any of this.
//
// One identity does not mean one literal CSS declaration, though: desktop
// keeps text-only blue (Glyph(Primary), no background — a slab would compete
// with the header it sits on) while mobile fills the chip solid with white
// text (As(Primary) — the content around it went white in the same pass that
// caused this report, so a tint-only toast would have blended straight into
// it). Both are Primary; what this test pins down is that BOTH read the exact
// same token pair for their role, so a future edit to one cannot quietly
// leave the other behind the way ColorOnSurface/ColorSuccess/etc. did before.
func TestMessageColorHasOnlyOneSource(t *testing.T) {
	p := &Platform{}
	p.Init(NilCtx())
	p.Notify(Msg.Info, "i", Persistent())
	p.Notify(Msg.Success, "s", Persistent())
	p.Notify(Msg.Warning, "w", Persistent())
	p.Notify(Msg.Error, "e", Persistent())

	cssStr := p.RenderCSS().String()

	// Nothing left to disagree: no per-severity selector exists at all.
	for _, sel := range []string{
		".pd__msg-info {", ".pd__msg-success {", ".pd__msg-warning {", ".pd__msg-error {",
	} {
		if Contains(cssStr, sel) {
			t.Errorf("found %q — severity-specific message styling must not exist; "+
				"style .pd__msg once and let every device reinterpret the same Primary "+
				"identity for its own background", sel)
		}
	}

	// Desktop (the first, unconditional ".pd__msg {" rule — Glyph(Primary)
	// alone, no As(), so nothing here can also set a background): text-only
	// blue, no fill.
	desktop := ruleBlock(cssStr, ".pd__msg {")
	if desktop == "" {
		t.Fatal("expected a base rule for .pd__msg")
	}
	if !Contains(desktop, css.ColorPrimary.LightValue()) {
		t.Errorf(".pd__msg (desktop) must use ColorPrimary (%s) as its text color, block:\n%s",
			css.ColorPrimary.LightValue(), desktop)
	}
	if Contains(desktop, "background-color") {
		t.Errorf(".pd__msg (desktop) must stay text-only, no background, block:\n%s", desktop)
	}

	// Mobile: the filled counterpart — blue background, white text.
	maxWIdx := strings.Index(cssStr, "@media (max-width")
	if maxWIdx == -1 {
		t.Fatal("expected a mobile (max-width) media query")
	}
	mobileRegion := cssStr[maxWIdx:]
	if next := strings.Index(mobileRegion[1:], "@media"); next != -1 {
		mobileRegion = mobileRegion[:next+1]
	}
	mobile := ruleBlock(mobileRegion, ".pd__msg {")
	if mobile == "" {
		t.Fatal("expected a mobile rule for .pd__msg")
	}
	if !Contains(mobile, "background-color: "+css.ColorPrimary.LightValue()) {
		t.Errorf(".pd__msg (mobile) must fill with ColorPrimary (%s), block:\n%s",
			css.ColorPrimary.LightValue(), mobile)
	}
	if !Contains(mobile, css.ColorOnPrimary.LightValue()) {
		t.Errorf(".pd__msg (mobile) must pair the fill with ColorOnPrimary (%s) text, block:\n%s",
			css.ColorOnPrimary.LightValue(), mobile)
	}

	// Neither rule may reach for a severity tint or the plain body ink —
	// both are exactly the bugs this test exists to keep out.
	for _, b := range []string{desktop, mobile} {
		for _, tok := range []css.Token{css.ColorSuccess, css.ColorDanger, css.ColorAccent, css.ColorMuted, css.ColorOnSurface} {
			if Contains(b, tok.LightValue()) {
				t.Errorf("message rule must not carry %s (%s), block:\n%s", tok.Name, tok.LightValue(), b)
			}
		}
	}
}
