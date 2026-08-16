package landing

import (
	"github.com/tinywasm/dom"
	"github.com/tinywasm/html"
	"github.com/tinywasm/image"
	"github.com/tinywasm/widget"

	"github.com/tinywasm/components/contentcard"
	"github.com/tinywasm/components/herobanner"
	"github.com/tinywasm/components/infobar"
	"github.com/tinywasm/components/sitenav"
	"github.com/tinywasm/components/statgrid"
)

const NameLanding = widget.Name("landing")

// The band (<header>/<section>/<footer>) owns the outer rhythm — padding and
// vertical stacking. The parts below it own only their internal arrangement, so
// no element ever carries the band class twice.
var (
	clsLanding = NameLanding.Root()

	partHeader   = widget.Part("header")
	partSection  = widget.Part("section")
	partFooter   = widget.Part("footer")
	partSplit    = widget.Part("split")
	partCards    = widget.Part("cards")
	partForm     = widget.Part("form")
	partHours    = widget.Part("hours")
	partMap      = widget.Part("map")
	partTitle    = widget.Part("title")
	partSubtitle = widget.Part("subtitle")
	partBody     = widget.Part("body")
	partMedia    = widget.Part("media")
	partGrid     = widget.Part("grid")
	partBadge    = widget.Part("badge")
	partBrand    = widget.Part("brand")

	clsHeader   = NameLanding.Class(partHeader)
	clsSection  = NameLanding.Class(partSection)
	clsFooter   = NameLanding.Class(partFooter)
	clsSplit    = NameLanding.Class(partSplit)
	clsCards    = NameLanding.Class(partCards)
	clsForm     = NameLanding.Class(partForm)
	clsHours    = NameLanding.Class(partHours)
	clsMap      = NameLanding.Class(partMap)
	clsTitle    = NameLanding.Class(partTitle)
	clsSubtitle = NameLanding.Class(partSubtitle)
	clsBody     = NameLanding.Class(partBody)
	clsMedia    = NameLanding.Class(partMedia)
	clsGrid     = NameLanding.Class(partGrid)
	clsBadge    = NameLanding.Class(partBadge)
	clsBrand    = NameLanding.Class(partBrand)
)

// Brand defines the site identity parameters for headers and footers.
type Brand struct {
	Name           string
	WideLogoSrc    string
	CompactLogoSrc string
	LogoAlt        string
	Href           string
}

// Link represents a site navigation link entry.
type Link struct {
	Label  string
	Href   string
	Active bool
}

// Contact holds business contact information.
type Contact struct {
	Phone   string
	Email   string
	Address string
	Hours   string
}

// Schedule holds operational time intervals for a range of days.
type Schedule struct {
	Days  string
	Hours string
}

// Card defines a content item used in card grid sections.
type Card struct {
	Title       string
	Description string
	Image       string
	Href        string
	Badge       string
}

// Stat represents a numerical metric and its descriptive label.
type Stat struct {
	Value string
	Label string
}

// Slide represents a hero gallery slide. It carries only the image because
// that is all herobanner renders today: a caption field here would be accepted
// and silently dropped. Captions are added when herobanner grows them upstream.
type Slide struct {
	Image string
}

// SubPage represents a detail page in a multi-page site.
type SubPage struct {
	Path      string
	Doc       html.DocumentOptions
	Sections  []*Section
	Component dom.Component
}

// sectionKind selects the band element and class a section renders into.
type sectionKind string

const (
	kindChrome  sectionKind = "chrome" // infobar and header: top strips, no band padding
	kindContent sectionKind = "content"
	kindFooter  sectionKind = "footer"
)

// Section is an explicit page section with optional navigation anchor ID.
// It is built by the section functions (Hero, Split, Cards, …); the only thing
// a consumer does with one is chain .At(id).
//
// A section holds a builder, not an element: header and footer sections are
// rendered once per page of a multi-page site, and a dom element has exactly
// one parent — sharing it across pages panics.
type Section struct {
	id    string
	kind  sectionKind
	build func(brand Brand) dom.Component
}

// At sets the navigation anchor ID explicitly for this section.
func (s *Section) At(id string) *Section {
	s.id = id
	return s
}

// InfoBar wraps infobar.InfoBar component into a page section.
func InfoBar(contact Contact) *Section {
	var items []infobar.InfoItem
	if contact.Phone != "" {
		items = append(items, infobar.InfoItem{
			Icon: infobar.IconPhone,
			Text: contact.Phone,
			Href: "tel:" + contact.Phone,
		})
	}
	if contact.Email != "" {
		items = append(items, infobar.InfoItem{
			Icon: infobar.IconMail,
			Text: contact.Email,
			Href: "mailto:" + contact.Email,
		})
	}
	if contact.Address != "" {
		items = append(items, infobar.InfoItem{
			Icon: infobar.IconPin,
			Text: contact.Address,
		})
	}
	if contact.Hours != "" {
		items = append(items, infobar.InfoItem{
			Icon: infobar.IconClock,
			Text: contact.Hours,
		})
	}

	// Components are embedded through Render(): a component handed to Child()
	// as itself serializes from its zero embedded Element — an empty <></>.
	return &Section{
		kind: kindChrome,
		build: func(Brand) dom.Component {
			return (&infobar.InfoBar{Items: items}).Render()
		},
	}
}

// Header wraps sitenav.SiteNav component into a header navigation section.
func Header(menu ...Link) *Section {
	var navItems []sitenav.NavItem
	for _, m := range menu {
		navItems = append(navItems, sitenav.NavItem{
			Label:  m.Label,
			Href:   m.Href,
			Active: m.Active,
		})
	}

	return &Section{
		kind: kindChrome,
		build: func(brand Brand) dom.Component {
			return (&sitenav.SiteNav{
				Links:          navItems,
				WideLogoSrc:    brand.WideLogoSrc,
				CompactLogoSrc: brand.CompactLogoSrc,
				LogoAlt:        brand.LogoAlt,
			}).Render()
		},
	}
}

// Hero wraps herobanner.HeroBanner into a hero section.
func Hero(title, subtitle string, cta Link, slides ...Slide) *Section {
	var images []string
	for _, slide := range slides {
		if slide.Image != "" {
			images = append(images, slide.Image)
		}
	}

	return &Section{
		kind: kindContent,
		build: func(Brand) dom.Component {
			var actions []dom.Component
			if cta.Label != "" || cta.Href != "" {
				actions = append(actions, html.A(cta.Href).Text(cta.Label))
			}
			return (&herobanner.HeroBanner{
				Title:    title,
				Subtitle: subtitle,
				Images:   images,
				Actions:  actions,
			}).Render()
		},
	}
}

// Split creates a two-column story/content section with title, image, and paragraphs.
func Split(title string, imageSrc string, paragraphs ...string) *Section {
	return &Section{
		kind: kindContent,
		build: func(Brand) dom.Component {
			container := html.Div().Set(clsSplit.AsAttr())

			content := html.Div().Set(clsBody.AsAttr())
			if title != "" {
				content.Child(html.H2().Set(clsTitle.AsAttr()).Text(title))
			}
			for _, p := range paragraphs {
				if p != "" {
					content.Child(html.P().Text(p))
				}
			}
			container.Child(content)

			if imageSrc != "" {
				media := html.Div().Set(clsMedia.AsAttr())
				media.Child(image.Img(imageSrc, title).Lazy().AsElement())
				container.Child(media)
			}
			return container
		},
	}
}

// Cards wraps contentcard.ContentCard instances in a responsive card grid section.
func Cards(title string, cards ...Card) *Section {
	return &Section{
		kind: kindContent,
		build: func(Brand) dom.Component {
			container := html.Div().Set(clsCards.AsAttr())
			if title != "" {
				container.Child(html.H2().Set(clsTitle.AsAttr()).Text(title))
			}

			grid := html.Div().Set(clsGrid.AsAttr())
			for _, c := range cards {
				var head dom.Component
				if c.Image != "" {
					head = image.Img(c.Image, c.Title).Lazy().AsElement()
				}

				body := html.Div()
				if c.Badge != "" {
					body.Child(html.Span().Set(clsBadge.AsAttr()).Text(c.Badge))
				}
				if c.Title != "" {
					body.Child(html.H3().Text(c.Title))
				}
				if c.Description != "" {
					body.Child(html.P().Text(c.Description))
				}

				var foot dom.Component
				if c.Href != "" {
					foot = html.A(c.Href).Text(c.Title)
				}

				grid.Child((&contentcard.ContentCard{
					Header: head,
					Body:   body,
					Footer: foot,
				}).Render())
			}
			container.Child(grid)
			return container
		},
	}
}

// Stats wraps statgrid.StatGrid component into a metric grid section.
func Stats(stats ...Stat) *Section {
	var items []statgrid.StatItem
	for _, s := range stats {
		items = append(items, statgrid.StatItem{
			Value: s.Value,
			Label: s.Label,
		})
	}

	return &Section{
		kind: kindContent,
		build: func(Brand) dom.Component {
			return (&statgrid.StatGrid{Items: items}).Render()
		},
	}
}

// Form creates a form section embedding intro text and form component. The form
// itself is a single element the consumer built, so this section belongs to one
// page — a form placed on two pages is the same element with two parents.
func Form(title string, intro string, formComp dom.Component) *Section {
	return &Section{
		kind: kindContent,
		build: func(Brand) dom.Component {
			container := html.Div().Set(clsForm.AsAttr())
			if title != "" {
				container.Child(html.H2().Set(clsTitle.AsAttr()).Text(title))
			}
			if intro != "" {
				container.Child(html.P().Set(clsSubtitle.AsAttr()).Text(intro))
			}
			if formComp != nil {
				container.Child(formComp)
			}
			return container
		},
	}
}

// MapEmbed creates a map iframe embed section.
func MapEmbed(title string, mapURL string) *Section {
	return &Section{
		kind: kindContent,
		build: func(Brand) dom.Component {
			container := html.Div().Set(clsMap.AsAttr())
			if title != "" {
				container.Child(html.H2().Set(clsTitle.AsAttr()).Text(title))
			}
			if mapURL != "" {
				container.Child(dom.NewElement("iframe").
					Attr("src", mapURL).
					Attr("loading", "lazy"))
			}
			return container
		},
	}
}

// Hours creates an operational hours and contact section.
func Hours(title string, contact Contact, schedules ...Schedule) *Section {
	return &Section{
		kind: kindContent,
		build: func(Brand) dom.Component {
			container := html.Div().Set(clsHours.AsAttr())
			if title != "" {
				container.Child(html.H2().Set(clsTitle.AsAttr()).Text(title))
			}

			grid := html.Div().Set(clsGrid.AsAttr())

			contactCol := html.Div().Set(clsBody.AsAttr())
			if contact.Address != "" {
				contactCol.Child(html.P().Text(contact.Address))
			}
			if contact.Phone != "" {
				contactCol.Child(html.P().Child(html.A("tel:" + contact.Phone).Text(contact.Phone)))
			}
			if contact.Email != "" {
				contactCol.Child(html.P().Child(html.A("mailto:" + contact.Email).Text(contact.Email)))
			}
			grid.Child(contactCol)

			if len(schedules) > 0 {
				schedCol := html.Div().Set(clsBody.AsAttr())
				ul := html.Ul()
				for _, s := range schedules {
					ul.Child(html.Li().Text(s.Days + ": " + s.Hours))
				}
				schedCol.Child(ul)
				grid.Child(schedCol)
			}

			container.Child(grid)
			return container
		},
	}
}

// Footer creates a site footer section with navigation menu links. The brand
// line is added by the page, which is the only place that knows the Brand.
func Footer(menu ...Link) *Section {
	return &Section{
		kind: kindFooter,
		build: func(Brand) dom.Component {
			nav := html.Nav()
			for _, item := range menu {
				a := html.A(item.Href).Text(item.Label)
				if item.Active {
					a.Attr("aria-current", "page")
				}
				nav.Child(a)
			}
			return nav
		},
	}
}

// Page is the landing layout root structure.
type Page struct {
	dom.Element
	Brand    Brand
	Sections []*Section
	Doc      html.DocumentOptions
	SubPages []SubPage
}

// New constructs a landing Page root.
func New(brand Brand, sections ...*Section) *Page {
	return &Page{
		Brand:    brand,
		Sections: sections,
		Doc: html.DocumentOptions{
			Lang: "en",
		},
	}
}

// WidgetName implements widget.Widget.
func (p *Page) WidgetName() widget.Name { return NameLanding }

// WidgetKind implements widget.Widget.
func (p *Page) WidgetKind() widget.Kind { return widget.Region }

// WithSEO configures the root page head metadata.
func (p *Page) WithSEO(doc html.DocumentOptions) *Page {
	p.Doc = doc
	return p
}

// WithSubPages attaches additional pages for multi-page site output.
func (p *Page) WithSubPages(subpages ...SubPage) *Page {
	p.SubPages = append(p.SubPages, subpages...)
	return p
}

// AddSubPage adds a single detail subpage.
func (p *Page) AddSubPage(subpage SubPage) *Page {
	p.SubPages = append(p.SubPages, subpage)
	return p
}

// brandMark renders the site identity shown in the footer: the name, linked to
// Brand.Href when the site declares a home URL. Returns nil when unnamed.
func (p *Page) brandMark() *dom.Element {
	if p.Brand.Name == "" {
		return nil
	}
	mark := html.Div().Set(clsBrand.AsAttr())
	if p.Brand.Href != "" {
		return mark.Child(html.A(p.Brand.Href).Text(p.Brand.Name))
	}
	return mark.Child(html.Span().Text(p.Brand.Name))
}

// Render compiles the Page DOM structure.
func (p *Page) Render() *dom.Element {
	root := html.Div().Set(clsLanding.AsAttr())

	for _, s := range p.Sections {
		if s == nil {
			continue
		}

		var content dom.Component
		if s.build != nil {
			content = s.build(p.Brand)
		}

		// The band owns the rhythm; the section content owns its own class, so
		// the same class never lands twice on nested elements. Chrome bands
		// stay a div because infobar and sitenav bring their own <header>.
		cls, tag := clsSection, "section"
		switch s.kind {
		case kindChrome:
			cls, tag = clsHeader, "div"
		case kindFooter:
			cls, tag = clsFooter, "footer"
		}

		secEl := dom.NewElement(tag).Set(cls.AsAttr())
		if s.id != "" {
			secEl.Attr("id", s.id)
		}

		if s.kind == kindFooter {
			if brand := p.brandMark(); brand != nil {
				secEl.Child(brand)
			}
		}

		if content != nil {
			secEl.Child(content)
		}
		root.Child(secEl)
	}

	return root
}

// String serializes the Page to HTML string.
func (p *Page) String() string {
	return p.Render().String()
}

// RenderPages implements html.PagesProvider for multi-page static site emission.
func (p *Page) RenderPages() []html.Page {
	var pages []html.Page

	homeDoc := p.Doc
	if homeDoc.Title == "" {
		homeDoc.Title = p.Brand.Name
	}

	homePage := html.Page{
		Path: "/",
		Doc:  homeDoc,
		Body: p.String(),
	}
	pages = append(pages, homePage)

	for _, sp := range p.SubPages {
		subPath := sp.Path
		if subPath == "" {
			subPath = "/"
		}
		body := p.renderSubPage(sp)
		pages = append(pages, html.Page{
			Path: subPath,
			Doc:  sp.Doc,
			Body: body,
		})
	}

	// Two pages sharing a Title or a Description is the most common way to lose
	// per-page ranking, so it fails loudly at emission time. Compared pairwise
	// on the slice: a map would ship map machinery into the wasm binary for a
	// handful of pages.
	for i, page := range pages {
		for _, prev := range pages[:i] {
			if page.Doc.Title != "" && page.Doc.Title == prev.Doc.Title {
				panic("landing: duplicate page Title '" + page.Doc.Title + "' found on paths '" + prev.Path + "' and '" + page.Path + "'")
			}
			if page.Doc.Description != "" && page.Doc.Description == prev.Doc.Description {
				panic("landing: duplicate page Description '" + page.Doc.Description + "' found on paths '" + prev.Path + "' and '" + page.Path + "'")
			}
		}
	}

	return pages
}

// renderSubPage wraps the detail page content in the same chrome as the home
// page — a detail page that loses the header and footer is a different site.
func (p *Page) renderSubPage(sp SubPage) string {
	var body []*Section
	body = append(body, sp.Sections...)
	if sp.Component != nil {
		body = append(body, &Section{
			kind:  kindContent,
			build: func(Brand) dom.Component { return sp.Component },
		})
	}
	if len(body) == 0 {
		return ""
	}

	var subSections []*Section
	for _, s := range p.Sections {
		if s != nil && s.kind == kindChrome {
			subSections = append(subSections, s)
		}
	}
	subSections = append(subSections, body...)
	for _, s := range p.Sections {
		if s != nil && s.kind == kindFooter {
			subSections = append(subSections, s)
		}
	}

	return New(p.Brand, subSections...).String()
}
