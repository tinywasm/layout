package login

import (
	"strings"
	"testing"

	. "github.com/tinywasm/html"
)

func TestLogin_RendersTitleAndForm(t *testing.T) {
	form := Div().Attr("id", "the-form")
	html := (&Login{Title: "MJosefa CMS", Form: form}).Render().String()

	for _, want := range []string{"login", "login__card", "login__title", "MJosefa CMS", "the-form"} {
		if !strings.Contains(html, want) {
			t.Errorf("markup missing %q\n%s", want, html)
		}
	}
}

func TestLogin_SubtitleIsOptional(t *testing.T) {
	html := (&Login{Title: "App"}).Render().String()
	if strings.Contains(html, "login__subtitle") {
		t.Errorf("expected no subtitle line when Subtitle is empty\n%s", html)
	}

	html = (&Login{Title: "App", Subtitle: "Ingrese sus credenciales"}).Render().String()
	if !strings.Contains(html, "login__subtitle") || !strings.Contains(html, "Ingrese sus credenciales") {
		t.Errorf("expected the subtitle to render when set\n%s", html)
	}
}

func TestLogin_MarkIsOptional(t *testing.T) {
	html := (&Login{Title: "App"}).Render().String()
	if strings.Contains(html, "login__mark") {
		t.Errorf("expected no mark when LogoMark is empty\n%s", html)
	}

	html = (&Login{Title: "App", LogoMark: "data:image/svg+xml,x"}).Render().String()
	if !strings.Contains(html, "login__mark") {
		t.Errorf("expected a mark when LogoMark is set\n%s", html)
	}
	if !strings.Contains(html, "src='data:image/svg+xml,x'") {
		t.Errorf("expected the mark's src to carry LogoMark, got\n%s", html)
	}
}
