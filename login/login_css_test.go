//go:build !wasm

package login

import (
	"strings"
	"testing"
)

// RenderCSS vive en css.go, que es //go:build !wasm: este test tiene que
// llevar la misma etiqueta o el build de wasm no encuentra el metodo.
//
// The check is scoped to the ROOT rule (`.login {`), not the whole stylesheet:
// a part may legitimately use the brand colour (PartSubtitle is Glyph(Primary),
// brand-tinted text), and what this guards is only that the full-bleed backdrop
// stays the neutral Page surface — not a wash of --color-primary.
func TestLogin_RootUsesPageNotPrimary(t *testing.T) {
	sheet := (&Login{Title: "App"}).RenderCSS().String()

	rootBlocks := loginRuleBlocks(sheet, ".login {")
	if len(rootBlocks) == 0 {
		t.Fatalf("expected a `.login {` root rule, got:\n%s", sheet)
	}
	joined := strings.Join(rootBlocks, "\n---\n")
	if strings.Contains(joined, "--color-primary") {
		t.Errorf("login root must not paint the brand color as its backdrop, root rule:\n%s", joined)
	}
	if !strings.Contains(joined, "--color-background") {
		t.Errorf("login root must use the neutral page background, root rule:\n%s", joined)
	}
}

// loginRuleBlocks returns the declaration body of every rule whose selector
// line is exactly `sel` (standalone, not grouped or a __part), across layers.
func loginRuleBlocks(cssStr, sel string) []string {
	want := "\n" + sel
	var out []string
	for i := 0; ; {
		j := strings.Index(cssStr[i:], want)
		if j == -1 {
			return out
		}
		start := i + j + len(want)
		end := strings.Index(cssStr[start:], "}")
		if end == -1 {
			return out
		}
		out = append(out, cssStr[start:start+end])
		i = start + end
	}
}
