package landing

import (
	"github.com/tinywasm/dom"
	"github.com/tinywasm/html"
	"github.com/tinywasm/widget"

	"github.com/tinywasm/components/contentcard"
	"github.com/tinywasm/components/herobanner"
	"github.com/tinywasm/components/infobar"
	"github.com/tinywasm/components/sitenav"
	"github.com/tinywasm/components/statgrid"
)

const NameLanding = widget.Name("landing")

var (
	clsLanding = NameLanding.Root()

	partHeader   = widget.Part("header")
	partSection  = widget.Part("section")
	partSplit    = widget.Part("split")
	partCards    = widget.Part("cards")
	partStats    = widget.Part("stats")
	partForm     = widget.Part("form")
	partHours    = widget.Part("hours")
	partMap      = widget.Part("map")
	partFooter   = widget.Part("footer")
	partTitle    = widget.Part("title")
	partSubtitle = widget.Part("subtitle")
	partBody     = widget.Part("body")
	partMedia    = widget.Part("media")
	partGrid     = widget.Part("grid")

	clsHeader   = NameLanding.Class(partHeader)
	clsSection  = NameLanding.Class(partSection)
	clsSplit    = NameLanding.Class(partSplit)
	clsCards    = NameLanding.Class(partCards)
	clsStats    = NameLanding.Class(partStats)
	clsForm     = NameLanding.Class(partForm)
	clsHours    = NameLanding.Class(partHours)
	clsMap      = NameLanding.Class(partMap)
	clsFooter   = NameLanding.Class(partFooter)
	clsTitle    = NameLanding.Class(partTitle)
	clsSubtitle = NameLanding.Class(partSubtitle)
	clsBody     = NameLanding.Class(partBody)
	clsMedia    = NameLanding.Class(partMedia)
	clsGrid     = NameLanding.Class(partGrid)
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

// Slide represents a hero gallery slide.
type Slide struct {
	Image    string
	Title    string
	Subtitle string
}

// SubPage represents a detail page in a multi-page site.
type SubPage struct {
	Path      string
	Doc       html.DocumentOptions
	Sections  []*Section
	Component dom.Component
}

// SectionKind identifies the category of a landing page section.
type SectionKind string

const (
	KindInfoBar  SectionKind = "infobar"
	KindHeader   SectionKind = "header"
	KindHero     SectionKind = "hero"
	KindSplit    SectionKind = "split"
	KindCards    SectionKind = "cards"
	KindStats    SectionKind = "stats"
	KindForm     SectionKind = "form"
	KindHours    SectionKind = "hours"
	KindMapEmbed SectionKind = "map"
	KindFooter   SectionKind = "footer"
)

// Section is an explicit page section with optional navigation anchor ID.
type Section struct {
	id        string
	kind      SectionKind
	title     string
	component dom.Component
}

// At sets the navigation anchor ID explicitly for this section.
func (s *Section) At(id string) *Section {
	s.id = id
	return s
}

// ID returns the navigation anchor ID of the section.
func (s *Section) ID() string {
	return s.id
}

// Kind returns the SectionKind.
func (s *Section) Kind() SectionKind {
	return s.kind
}

// Title returns the section title.
func (s *Section) Title() string {
	return s.title
}

// Component returns the underlying dom.Component for custom or wrapped sections.
func (s *Section) Component() dom.Component {
	return s.component
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

	ib := &infobar.InfoBar{
		Items: items,
	}

	return &Section{
		kind:      KindInfoBar,
		component: ib,
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

	sn := &sitenav.SiteNav{
		Links: navItems,
	}

	return &Section{
		kind:      KindHeader,
		component: sn,
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

	var actions []dom.Component
	if cta.Label != "" || cta.Href != "" {
		actions = append(actions, html.A(cta.Label).Attr("href", cta.Href))
	}

	hb := &herobanner.HeroBanner{
		Title:    title,
		Subtitle: subtitle,
		Images:   images,
		Actions:  actions,
	}

	return &Section{
		kind:      KindHero,
		title:     title,
		component: hb,
	}
}

// Split creates a two-column story/content section with title, image, and paragraphs.
func Split(title string, image string, paragraphs ...string) *Section {
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

	if image != "" {
		media := html.Div().Set(clsMedia.AsAttr())
		media.Child(dom.NewElement("img").Attr("src", image).Attr("alt", title))
		container.Child(media)
	}

	return &Section{
		kind:      KindSplit,
		title:     title,
		component: container,
	}
}

// Cards wraps contentcard.ContentCard instances in a responsive card grid section.
func Cards(title string, cards ...Card) *Section {
	container := html.Div().Set(clsCards.AsAttr())
	if title != "" {
		container.Child(html.H2().Set(clsTitle.AsAttr()).Text(title))
	}

	grid := html.Div().Set(clsGrid.AsAttr())
	for _, c := range cards {
		var head dom.Component
		if c.Image != "" {
			head = dom.NewElement("img").Attr("src", c.Image).Attr("alt", c.Title)
		}

		body := html.Div()
		if c.Title != "" {
			body.Child(html.H3().Text(c.Title))
		}
		if c.Description != "" {
			body.Child(html.P().Text(c.Description))
		}

		var foot dom.Component
		if c.Href != "" {
			foot = html.A(c.Title).Attr("href", c.Href)
		}

		cc := &contentcard.ContentCard{
			Header: head,
			Body:   body,
			Footer: foot,
		}
		grid.Child(cc)
	}
	container.Child(grid)

	return &Section{
		kind:      KindCards,
		title:     title,
		component: container,
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

	sg := &statgrid.StatGrid{
		Items: items,
	}

	return &Section{
		kind:      KindStats,
		component: sg,
	}
}

// Form creates a form section embedding intro text and form component.
func Form(title string, intro string, formComp dom.Component) *Section {
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

	return &Section{
		kind:      KindForm,
		title:     title,
		component: container,
	}
}

// MapEmbed creates a map iframe embed section.
func MapEmbed(title string, mapURL string) *Section {
	container := html.Div().Set(clsMap.AsAttr())
	if title != "" {
		container.Child(html.H2().Set(clsTitle.AsAttr()).Text(title))
	}
	if mapURL != "" {
		iframe := dom.NewElement("iframe").
			Attr("src", mapURL).
			Attr("loading", "lazy")
		container.Child(iframe)
	}
	return &Section{
		kind:      KindMapEmbed,
		title:     title,
		component: container,
	}
}

// Hours creates an operational hours and contact section.
func Hours(title string, contact Contact, schedules ...Schedule) *Section {
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
		contactCol.Child(html.P().Child(html.A(contact.Phone).Attr("href", "tel:"+contact.Phone)))
	}
	if contact.Email != "" {
		contactCol.Child(html.P().Child(html.A(contact.Email).Attr("href", "mailto:"+contact.Email)))
	}
	grid.Child(contactCol)

	if len(schedules) > 0 {
		schedCol := html.Div().Set(clsHours.AsAttr())
		ul := html.Ul()
		for _, s := range schedules {
			li := html.Li().Text(s.Days + ": " + s.Hours)
			ul.Child(li)
		}
		schedCol.Child(ul)
		grid.Child(schedCol)
	}

	container.Child(grid)
	return &Section{
		kind:      KindHours,
		title:     title,
		component: container,
	}
}

// Footer creates a site footer section with navigation menu links.
func Footer(menu ...Link) *Section {
	container := html.Div().Set(clsFooter.AsAttr())
	if len(menu) > 0 {
		nav := html.Nav()
		for _, item := range menu {
			a := html.A(item.Label).Attr("href", item.Href)
			if item.Active {
				a.Attr("aria-current", "page")
			}
			nav.Child(a)
		}
		container.Child(nav)
	}
	return &Section{
		kind:      KindFooter,
		component: container,
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

// Render compiles the Page DOM structure.
func (p *Page) Render() *dom.Element {
	root := html.Div().Set(clsLanding.AsAttr())

	for _, s := range p.Sections {
		if s == nil {
			continue
		}

		if sn, ok := s.component.(*sitenav.SiteNav); ok {
			if sn.WideLogoSrc == "" {
				sn.WideLogoSrc = p.Brand.WideLogoSrc
			}
			if sn.CompactLogoSrc == "" {
				sn.CompactLogoSrc = p.Brand.CompactLogoSrc
			}
			if sn.LogoAlt == "" {
				sn.LogoAlt = p.Brand.LogoAlt
			}
		}

		var cls = clsSection
		switch s.kind {
		case KindInfoBar:
			cls = clsHeader
		case KindHeader:
			cls = clsHeader
		case KindHero:
			cls = clsSection
		case KindSplit:
			cls = clsSplit
		case KindCards:
			cls = clsCards
		case KindStats:
			cls = clsStats
		case KindForm:
			cls = clsForm
		case KindHours:
			cls = clsHours
		case KindMapEmbed:
			cls = clsMap
		case KindFooter:
			cls = clsFooter
		default:
			cls = clsSection
		}

		tag := "section"
		if s.kind == KindHeader {
			tag = "header"
		} else if s.kind == KindFooter {
			tag = "footer"
		}

		secEl := dom.NewElement(tag).Set(cls.AsAttr())
		if s.id != "" {
			secEl.Attr("id", s.id)
		}

		if s.component != nil {
			secEl.Child(s.component)
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

	seenTitles := make(map[string]string)
	seenDescs := make(map[string]string)

	for _, page := range pages {
		t := page.Doc.Title
		if t != "" {
			if prevPath, exists := seenTitles[t]; exists {
				panic("landing: duplicate page Title '" + t + "' found on paths '" + prevPath + "' and '" + page.Path + "'")
			}
			seenTitles[t] = page.Path
		}

		d := page.Doc.Description
		if d != "" {
			if prevPath, exists := seenDescs[d]; exists {
				panic("landing: duplicate page Description '" + d + "' found on paths '" + prevPath + "' and '" + page.Path + "'")
			}
			seenDescs[d] = page.Path
		}
	}

	return pages
}

func (p *Page) renderSubPage(sp SubPage) string {
	if sp.Component != nil {
		return sp.Component.String()
	}
	if len(sp.Sections) > 0 {
		var subSections []*Section
		for _, s := range p.Sections {
			if s != nil && (s.kind == KindInfoBar || s.kind == KindHeader) {
				subSections = append(subSections, s)
			}
		}
		subSections = append(subSections, sp.Sections...)
		for _, s := range p.Sections {
			if s != nil && s.kind == KindFooter {
				subSections = append(subSections, s)
			}
		}

		subPage := New(p.Brand, subSections...)
		return subPage.String()
	}
	return ""
}
