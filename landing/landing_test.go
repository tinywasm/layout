package landing_test

import (
	"strings"
	"testing"

	"github.com/tinywasm/dom"
	"github.com/tinywasm/html"
	"github.com/tinywasm/layout/landing"
)

type dummyComponent struct {
	html string
}

func (d *dummyComponent) GetID() string             { return "dummy" }
func (d *dummyComponent) SetID(_ string)            {}
func (d *dummyComponent) String() string            { return d.html }
func (d *dummyComponent) Children() []dom.Component { return nil }

func TestLandingFullCompositionAndSectionOrdering(t *testing.T) {
	marca := landing.Brand{
		Name:        "Clínica San José",
		WideLogoSrc: "/logo-wide.svg",
		LogoAlt:     "Logo San José",
	}

	contacto := landing.Contact{
		Phone:   "+56912345678",
		Email:   "contacto@sanjose.cl",
		Address: "Av. Libertad 123",
		Hours:   "Lunes a Viernes 8:00 - 18:00",
	}

	menu := []landing.Link{
		{Label: "Inicio", Href: "#inicio"},
		{Label: "Nosotros", Href: "#nosotros"},
		{Label: "Especialidades", Href: "#especialidades"},
		{Label: "Compromiso", Href: "#compromiso"},
		{Label: "Voluntariado", Href: "#voluntariado"},
		{Label: "Contacto", Href: "#contacto"},
		{Label: "Ubicación", Href: "#ubicacion"},
	}

	tarjetas := []landing.Card{
		{Title: "Oftalmología", Description: "Cuidado visual integral", Href: "/especialidades/oftalmologia/"},
		{Title: "Pediatría", Description: "Atención infantil", Href: "/especialidades/pediatria/"},
	}

	cifras := []landing.Stat{
		{Value: "+10.000", Label: "Pacientes atendidos"},
		{Value: "15", Label: "Años de experiencia"},
	}

	horarios := []landing.Schedule{
		{Days: "Lunes a Viernes", Hours: "08:00 - 18:00"},
		{Days: "Sábado", Hours: "09:00 - 13:00"},
	}

	galeria := []landing.Slide{
		{Image: "/slide1.jpg", Title: "Bienvenidos", Subtitle: "Atención médica humana"},
	}

	cta := landing.Link{Label: "Reservar hora", Href: "#contacto"}
	formComp := &dummyComponent{html: "<form id='voluntariado-form'></form>"}

	page := landing.New(marca,
		landing.InfoBar(contacto),
		landing.Header(menu...),
		landing.Hero("Atención Médica Integral", "Cuidamos tu salud y la de tu familia", cta, galeria...).At("inicio"),
		landing.Split("Nuestra Historia", "/historia.jpg", "Fundada en 2010...", "Comprometidos con la comunidad...").At("nosotros"),
		landing.Cards("Especialidades", tarjetas...).At("especialidades"),
		landing.Stats(cifras...).At("compromiso"),
		landing.Form("Voluntariado", "Únete a nuestro equipo", formComp).At("voluntariado"),
		landing.Hours("Contáctanos", contacto, horarios...).At("contacto"),
		landing.MapEmbed("Ubicación", "https://maps.google.com/embed").At("ubicacion"),
		landing.Footer(menu...),
	)

	out := page.String()

	// Verify section anchor IDs and rendering order
	anchors := []string{
		"id='inicio'",
		"id='nosotros'",
		"id='especialidades'",
		"id='compromiso'",
		"id='voluntariado'",
		"id='contacto'",
		"id='ubicacion'",
	}

	lastIdx := -1
	for _, anchor := range anchors {
		idx := strings.Index(out, anchor)
		if idx == -1 {
			t.Errorf("expected markup to contain anchor %s, got:\n%s", anchor, out)
		} else if idx <= lastIdx {
			t.Errorf("anchor %s appears out of order in rendered markup", anchor)
		}
		lastIdx = idx
	}

	// Verify menu links match anchors
	for _, m := range menu {
		if !strings.Contains(out, "href='"+m.Href+"'") && !strings.Contains(out, "href=\""+m.Href+"\"") {
			t.Errorf("expected menu link %s to be present in markup", m.Href)
		}
	}
}

func TestRenderPagesMultiPageAndUniqueMetadata(t *testing.T) {
	marca := landing.Brand{Name: "Clínica San José"}

	subPages := []landing.SubPage{
		{
			Path: "/especialidades/oftalmologia/",
			Doc: html.DocumentOptions{
				Title:       "Oftalmología — Clínica San José",
				Description: "Servicio especializado en cuidado visual y cirugías oculares.",
				Canonical:   "https://sanjose.cl/especialidades/oftalmologia/",
			},
			Sections: []*landing.Section{
				landing.Split("Oftalmología", "/oftalmo.jpg", "Diagnóstico y tratamiento..."),
			},
		},
		{
			Path: "/especialidades/pediatria/",
			Doc: html.DocumentOptions{
				Title:       "Pediatría — Clínica San José",
				Description: "Atención pediátrica integral para bebés, niños y adolescentes.",
				Canonical:   "https://sanjose.cl/especialidades/pediatria/",
			},
			Sections: []*landing.Section{
				landing.Split("Pediatría", "/pediatria.jpg", "Cuidado especializado infantil..."),
			},
		},
	}

	page := landing.New(marca,
		landing.Hero("Portada", "Bajada", landing.Link{}),
	).WithSEO(html.DocumentOptions{
		Title:       "Portada — Clínica San José",
		Description: "Centro médico de salud de referencia.",
		Canonical:   "https://sanjose.cl/",
	}).WithSubPages(subPages...)

	pages := page.RenderPages()

	if len(pages) != 3 {
		t.Fatalf("expected 3 pages emitted (1 home + 2 subpages), got %d", len(pages))
	}

	// Verify paths and trailing slashes
	expectedPaths := []string{"/", "/especialidades/oftalmologia/", "/especialidades/pediatria/"}
	for i, expected := range expectedPaths {
		if pages[i].Path != expected {
			t.Errorf("page[%d] Path = %q, expected %q", i, pages[i].Path, expected)
		}
	}

	// Verify titles and descriptions are distinct
	seenTitles := make(map[string]bool)
	seenDescs := make(map[string]bool)
	for _, p := range pages {
		if seenTitles[p.Doc.Title] {
			t.Errorf("duplicate page Title detected: %q", p.Doc.Title)
		}
		seenTitles[p.Doc.Title] = true

		if seenDescs[p.Doc.Description] {
			t.Errorf("duplicate page Description detected: %q", p.Doc.Description)
		}
		seenDescs[p.Doc.Description] = true
	}
}

func TestRenderPagesDuplicateTitlePanic(t *testing.T) {
	page := landing.New(landing.Brand{Name: "Test"},
		landing.Hero("H", "S", landing.Link{}),
	).WithSEO(html.DocumentOptions{
		Title:       "Duplicate Title",
		Description: "Desc 1",
	}).WithSubPages(landing.SubPage{
		Path: "/subpage/",
		Doc: html.DocumentOptions{
			Title:       "Duplicate Title",
			Description: "Desc 2",
		},
	})

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected RenderPages to panic on duplicate Title, but it did not")
		} else if !strings.Contains(r.(string), "duplicate page Title") {
			t.Errorf("expected duplicate Title panic message, got: %v", r)
		}
	}()

	_ = page.RenderPages()
}

func TestRenderPagesDuplicateDescriptionPanic(t *testing.T) {
	page := landing.New(landing.Brand{Name: "Test"},
		landing.Hero("H", "S", landing.Link{}),
	).WithSEO(html.DocumentOptions{
		Title:       "Title 1",
		Description: "Shared Description",
	}).WithSubPages(landing.SubPage{
		Path: "/subpage/",
		Doc: html.DocumentOptions{
			Title:       "Title 2",
			Description: "Shared Description",
		},
	})

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected RenderPages to panic on duplicate Description, but it did not")
		} else if !strings.Contains(r.(string), "duplicate page Description") {
			t.Errorf("expected duplicate Description panic message, got: %v", r)
		}
	}()

	_ = page.RenderPages()
}
