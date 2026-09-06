package landing_test

import (
	"fmt"
	"strings"
	"testing"

	"webtyp.com/dom"
	"webtyp.com/html"
	"webtyp.com/image"
	"webtyp.com/layout/landing"
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
		Name:        "Clínica Demo",
		WideLogoSrc: "/logo-wide.svg",
		LogoAlt:     "Logo Demo",
		Href:        "/",
	}

	contacto := landing.Contact{
		Phone:   "+56 9 0000 0000",
		Email:   "contacto@example.com",
		Address: "Av. Siempre Viva 123",
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
		{Title: "Oftalmología", Description: "Cuidado visual integral", Href: "/especialidades/oftalmologia/", Image: "/oftalmo.jpg", Badge: "Nuevo"},
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
		{Image: "/slide1.jpg"},
	}

	cta := landing.Link{Label: "Reservar hora", Href: "#contacto"}
	formComp := &dummyComponent{html: "<form id='voluntariado-form'></form>"}

	page := landing.New(marca,
		landing.InfoBar(contacto),
		landing.Header(menu...),
		landing.Hero("Atención Médica Integral", "Cuidamos tu salud y la de tu familia", []landing.Link{cta}, galeria...).At("inicio"),
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

	// Every menu entry must land on an anchor that actually exists: a link to a
	// missing id is a menu that silently scrolls nowhere.
	for _, m := range menu {
		if !strings.Contains(out, "href='"+m.Href+"'") && !strings.Contains(out, "href=\""+m.Href+"\"") {
			t.Errorf("expected menu link %s to be present in markup", m.Href)
		}
		if strings.HasPrefix(m.Href, "#") {
			if !strings.Contains(out, "id='"+strings.TrimPrefix(m.Href, "#")+"'") {
				t.Errorf("menu links to %s but no section carries that anchor id", m.Href)
			}
		}
	}

	// The band carries the section class; the content inside carries its own.
	// The same class twice on nested elements means padding applied twice.
	for _, cls := range []string{"landing__section", "landing__split", "landing__cards", "landing__form", "landing__hours", "landing__map", "landing__footer"} {
		if strings.Contains(out, "class='"+cls+"'><div class='"+cls+"'") {
			t.Errorf("class %s is nested inside itself: the band and its content must not share a class", cls)
		}
	}

	// <img> is a void element; a closing tag is invalid markup. Scoped to the
	// images this layout emits (all lazy) — herobanner's own eager images are
	// an upstream concern.
	if strings.Contains(out, "loading='lazy'></img>") {
		t.Errorf("images must render as void elements (image.Img), got a closing </img> tag:\n%s", out)
	}

	// Every wrapped component must reach the markup. A component handed to
	// Child() as itself serializes as an empty <></>, which no id/href
	// assertion above would notice.
	for _, marker := range []string{"infobar", "sitenav", "herobanner", "contentcard", "statgrid"} {
		if !strings.Contains(out, "class='"+marker) {
			t.Errorf("component %q does not appear in the rendered markup: it was embedded but never rendered:\n%s", marker, out)
		}
	}
	if strings.Contains(out, "<></>") {
		t.Errorf("markup contains an empty <></> node — a component was embedded without Render():\n%s", out)
	}

	// A link renders its label, not just its href.
	for _, label := range []string{"Reservar hora", "Oftalmología", "contacto@example.com"} {
		if !strings.Contains(out, ">"+label+"<") {
			t.Errorf("link label %q is missing from the markup: html.A takes the href, the label is the text", label)
		}
	}

	// Fields that are accepted but never rendered are silent failures.
	if !strings.Contains(out, "Nuevo") {
		t.Errorf("Card.Badge is accepted but not rendered")
	}
	if !strings.Contains(out, "class='landing__brand'") || !strings.Contains(out, marca.Name) {
		t.Errorf("Brand.Name/Href are accepted but the footer renders no brand mark")
	}
}

func TestRenderPagesZeroValueProbeReturnsNoPages(t *testing.T) {
	// sitec's extraction discovers RenderPages by calling it on a bare
	// &landing.Page{} to check the method exists — that probe must not itself
	// register a homepage at "/", or it collides with the real one.
	probe := &landing.Page{}

	if pages := probe.RenderPages(); pages != nil {
		t.Fatalf("expected nil pages for the zero-value probe, got %d", len(pages))
	}
}

func TestRenderPagesMultiPageAndUniqueMetadata(t *testing.T) {
	marca := landing.Brand{Name: "Clínica Demo"}

	subPages := []landing.SubPage{
		{
			Path: "/especialidades/oftalmologia/",
			Doc: html.DocumentOptions{
				Title:       "Oftalmología — Clínica Demo",
				Description: "Servicio especializado en cuidado visual y cirugías oculares.",
				Canonical:   "https://example.com/especialidades/oftalmologia/",
			},
			Sections: []*landing.Section{
				landing.Split("Oftalmología", "/oftalmo.jpg", "Diagnóstico y tratamiento..."),
			},
		},
		{
			Path: "/especialidades/pediatria/",
			Doc: html.DocumentOptions{
				Title:       "Pediatría — Clínica Demo",
				Description: "Atención pediátrica integral para bebés, niños y adolescentes.",
				Canonical:   "https://example.com/especialidades/pediatria/",
			},
			Sections: []*landing.Section{
				landing.Split("Pediatría", "/pediatria.jpg", "Cuidado especializado infantil..."),
			},
		},
	}

	page := landing.New(marca,
		landing.Header(landing.Link{Label: "Inicio", Href: "/"}),
		landing.Hero("Portada", "Bajada", nil),
		landing.Footer(landing.Link{Label: "Inicio", Href: "/"}),
	).WithSEO(html.DocumentOptions{
		Title:       "Portada — Clínica Demo",
		Description: "Centro médico de salud de referencia.",
		Canonical:   "https://example.com/",
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

	// A detail page keeps the site chrome: without header and footer it is a
	// different site, not another page of this one.
	for _, sub := range pages[1:] {
		if !strings.Contains(sub.Body, "class='sitenav") {
			t.Errorf("page %s renders no header: detail pages keep the site chrome", sub.Path)
		}
		if !strings.Contains(sub.Body, "class='landing__footer'") {
			t.Errorf("page %s renders no footer: detail pages keep the site chrome", sub.Path)
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
		landing.Hero("H", "S", nil),
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

func TestLandingResponsiveImages(t *testing.T) {
	pageWithImages := landing.New(
		landing.Brand{Name: "Test"},
		landing.Split("Historia", "/img/historia.jpg", "Parrafo 1"),
		landing.Cards("Servicios", landing.Card{Title: "Card 1", Image: "/img/card1.jpg"}),
	)
	out := pageWithImages.String()

	// 1. Split con imagen: el <img> trae srcset con las tres variantes y src predeterminado
	expectedSplitSrcSet := fmt.Sprintf(
		"srcset='/img/historia.S.jpg %dw, /img/historia.M.jpg %dw, /img/historia.L.jpg %dw'",
		image.VariantS.Width(), image.VariantM.Width(), image.VariantL.Width(),
	)
	if !strings.Contains(out, expectedSplitSrcSet) && !strings.Contains(out, strings.ReplaceAll(expectedSplitSrcSet, "'", "\"")) {
		t.Errorf("expected Split image to contain srcset %q, got output:\n%s", expectedSplitSrcSet, out)
	}

	// 2. Split: sizes igual a SizesSplit
	expectedSplitSizes := "sizes='" + landing.SizesSplit + "'"
	if !strings.Contains(out, expectedSplitSizes) && !strings.Contains(out, strings.ReplaceAll(expectedSplitSizes, "'", "\"")) {
		t.Errorf("expected Split image sizes %q, got output:\n%s", expectedSplitSizes, out)
	}

	// 3. Cards con imagen: srcset con las tres variantes
	expectedCardSrcSet := fmt.Sprintf(
		"srcset='/img/card1.S.jpg %dw, /img/card1.M.jpg %dw, /img/card1.L.jpg %dw'",
		image.VariantS.Width(), image.VariantM.Width(), image.VariantL.Width(),
	)
	if !strings.Contains(out, expectedCardSrcSet) && !strings.Contains(out, strings.ReplaceAll(expectedCardSrcSet, "'", "\"")) {
		t.Errorf("expected Card image to contain srcset %q, got output:\n%s", expectedCardSrcSet, out)
	}

	// 4. Cards: sizes igual a SizesCard
	expectedCardSizes := "sizes='" + landing.SizesCard + "'"
	if !strings.Contains(out, expectedCardSizes) && !strings.Contains(out, strings.ReplaceAll(expectedCardSizes, "'", "\"")) {
		t.Errorf("expected Card image sizes %q, got output:\n%s", expectedCardSizes, out)
	}

	// 5. Split y Cards: siguen con loading="lazy"
	if strings.Count(out, "loading='lazy'") < 2 && strings.Count(out, "loading=\"lazy\"") < 2 {
		t.Errorf("expected both Split and Card images to have loading='lazy', got output:\n%s", out)
	}

	// 6. Split sin imagen: no emite el media, sin pánico
	pageNoImages := landing.New(
		landing.Brand{Name: "Test"},
		landing.Split("Historia Sin Imagen", "", "Parrafo"),
		landing.Cards("Servicios Sin Imagen", landing.Card{Title: "Card sin foto"}),
	)
	outNoImages := pageNoImages.String()

	if strings.Contains(outNoImages, "landing__media") {
		t.Errorf("Split without image should not render media container, got:\n%s", outNoImages)
	}

	// 7. Card sin imagen: no emite la cabecera, sin pánico
	if strings.Contains(outNoImages, "<img") {
		t.Errorf("Cards without image should not render any <img tag, got:\n%s", outNoImages)
	}
}

func TestRenderPagesDuplicateDescriptionPanic(t *testing.T) {
	page := landing.New(landing.Brand{Name: "Test"},
		landing.Hero("H", "S", nil),
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
