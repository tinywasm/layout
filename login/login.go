package login

import (
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/html"
	"github.com/tinywasm/widget"
)

// NameLogin is the widget identity; it produces the class prefix "login" and
// the part classes "login__card", "login__header", "login__title",
// "login__subtitle", "login__mark".
const NameLogin = widget.Name("login")

const (
	PartCard     = widget.Part("card")     // the centered card: header + form
	PartHeader   = widget.Part("header")   // title + subtitle, packed as one block
	PartTitle    = widget.Part("title")    // the app's own name, leading the card
	PartSubtitle = widget.Part("subtitle") // the one line of orientation under it
	PartMark     = widget.Part("mark")     // the corner brand mark, independent of the card
)

var (
	clsLogin    = NameLogin.Root()
	clsCard     = NameLogin.Class(PartCard)
	clsHeader   = NameLogin.Class(PartHeader)
	clsTitle    = NameLogin.Class(PartTitle)
	clsSubtitle = NameLogin.Class(PartSubtitle)
	clsMark     = NameLogin.Class(PartMark)
)

// Login is the pre-authentication screen: an elevated card (title, subtitle,
// form) centered on the app's own page background, with an optional corner
// mark pinned independently of that card — the reference this productionizes
// (a legacy pa100t deployment) keeps its own crest bottom-left regardless of
// how tall the form above it grows, which a mark living inside the card would
// not survive.
//
// The backdrop is Page — the same neutral surface every other screen sits
// on — not Primary: a brand color saturated enough to read well on a button
// rarely survives being stretched across an entire viewport, and the token
// system already reserves Primary for bounded controls (see widget/style's
// own Surface.resolve()). The card itself, plus an optional LogoMark, carry
// the brand instead of the backdrop.
//
// It owns none of the form's fields or validation — Form is built by the
// composition root exactly like Platform.Modules are, so this package never
// assumes a shape for what it centers, only that it Renders.
type Login struct {
	Element // value embed — NEVER *dom.Element (TinyGo heap constraint)

	// Title leads the card — the app's own name, or a translated
	// "Ingreso"/"Sign in" if the app prefers that framing over its brand.
	// Required: the screen reads as unbranded without it.
	Title string

	// Subtitle is the single line under the title telling the user what this
	// screen wants from them ("Ingrese sus credenciales para continuar").
	// Optional — omitted, the title sits alone and the card is that much
	// tighter.
	Subtitle string

	// Form is the actual login form, built and validated by the caller.
	Form Component

	// LogoMark is a data-URI or URL for the corner brand mark, mirroring
	// platformd.Brand.BrandMark's own contract (a string, not an svg.Icon:
	// a crest or seal is a full-color image, not a currentColor glyph this
	// package could recolor). Optional — a Login with no mark simply does
	// not render the corner.
	LogoMark string
}

func (l *Login) WidgetName() widget.Name { return NameLogin }

// Region: this screen is a landmark a user orients by, not a control with
// its own interaction pattern — the Form inside it is the interactive
// piece, and carries its own Kind.
func (l *Login) WidgetKind() widget.Kind { return widget.Region }

func (l *Login) Render() *Element {
	header := Div().Set(clsHeader.AsAttr()).Child(
		H1().Set(clsTitle.AsAttr()).Text(l.Title),
	)
	if l.Subtitle != "" {
		header.Child(P().Set(clsSubtitle.AsAttr()).Text(l.Subtitle))
	}

	root := Div().Set(clsLogin.AsAttr())
	root.Child(
		Div().Set(clsCard.AsAttr()).Child(header, l.Form),
	)

	if l.LogoMark != "" {
		root.Child(NewElement("img").NoCloseTag().Set(clsMark.AsAttr()).Attr("src", l.LogoMark).Attr("alt", ""))
	}

	return root
}
