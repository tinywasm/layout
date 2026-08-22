//go:build !wasm

package login

import (
	"strings"
	"testing"
)

// RenderCSS vive en css.go, que es //go:build !wasm: este test tiene que
// llevar la misma etiqueta o el build de wasm no encuentra el metodo.
func TestLogin_RootUsesPageNotPrimary(t *testing.T) {
	css := (&Login{Title: "App"}).RenderCSS().String()

	if strings.Contains(css, "--color-primary") {
		t.Errorf("login root must not paint the brand color as its backdrop, got:\n%s", css)
	}
	if !strings.Contains(css, "--color-background") {
		t.Errorf("login root must use the neutral page background, got:\n%s", css)
	}
}
