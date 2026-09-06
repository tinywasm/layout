//go:build !wasm

package landing_test

import (
	"strings"
	"testing"

	"webtyp.com/layout/landing"
)

func TestRenderCSS(t *testing.T) {
	page := landing.New(landing.Brand{Name: "Test"})
	stylesheet := page.RenderCSS()

	cssStr := stylesheet.String()

	if !strings.Contains(cssStr, "@layer tokens, primitives, widgets, states;") {
		t.Errorf("expected stylesheet to declare layer order @layer tokens, primitives, widgets, states;")
	}

	if strings.Contains(cssStr, "!important") {
		t.Errorf("stylesheet must not contain !important")
	}
}
