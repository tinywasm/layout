package platformd

import (
	"sync"

	"github.com/tinywasm/components/usermenu"
	"github.com/tinywasm/layout"
	"github.com/tinywasm/svg"
	"github.com/tinywasm/widget"

	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	. "github.com/tinywasm/html"
	"github.com/tinywasm/time"
)

const NamePlatform widget.Name = "pd"

var (
	clsRoot           = NamePlatform.Root()
	clsHeader         = NamePlatform.Class("header")
	clsDrawerPanel    = NamePlatform.Class("drawer-panel")
	clsAppName        = NamePlatform.Class("app-name")
	clsDrawerIdentity = NamePlatform.Class("drawer-identity")
	clsBrand          = NamePlatform.Class("brand")
	clsBrandMark      = NamePlatform.Class("brand-mark")
	clsBrandName      = NamePlatform.Class("brand-name")
	clsMsgSlot        = NamePlatform.Class("msg-slot")
	clsMsg            = NamePlatform.Class("msg")
	clsMsgInfo        = NamePlatform.Class("msg-info")
	clsMsgSuccess     = NamePlatform.Class("msg-success")
	clsMsgWarning     = NamePlatform.Class("msg-warning")
	clsMsgError       = NamePlatform.Class("msg-error")
	clsHeaderRight    = NamePlatform.Class("header-right")
	clsBody           = NamePlatform.Class("body")
	clsStage          = NamePlatform.Class("stage")
	clsPanel          = NamePlatform.Class("panel")
	clsMenu           = NamePlatform.Class("menu")
	clsNavbar         = NamePlatform.Class("navbar")
	clsNavItem        = NamePlatform.Class("nav-item")
	clsNavLink        = NamePlatform.Class("nav-link")
	clsLinkText       = NamePlatform.Class("link-text")
	ClsNavIcon        = NamePlatform.Class("nav-icon")
	clsHamburger      = NamePlatform.Class("hamburger")
	clsNavOverlay     = NamePlatform.Class("nav-overlay")
)

const (
	IconUser  = svg.Icon("pd-user")
	IconBrand = svg.Icon("pd-brand")
	iconMenu  = svg.Icon("pd-menu")
)

// UIModule is a module that provides its UI to the platform chassis.
// The chassis takes id/hash/route from ModelName() and the rest of the
// presentation from these methods.
type UIModule interface {
	layout.Module    // identity: ModelName() → used as ID
	Label() string   // text in the nav rail
	Icon() svg.Icon  // chassis renders via the sprite
	View() Component // module content (often a *rightpanel.RightPanel)
}

func (p *Platform) WidgetName() widget.Name { return NamePlatform }
func (p *Platform) WidgetKind() widget.Kind { return widget.Menu }

// Identity is what the platform needs to know about whoever is logged in.
//
// It is a READ contract, not a store: platformd renders it and never mutates
// it. Whatever owns authentication — github.com/tinywasm/user in a real
// application — supplies an implementation; the platform neither knows nor
// cares where the values come from.
//
// It asks for facts, not for presentation. The glyph drawn when there is no
// avatar is IconUser, which this package owns — an authentication package has
// no business choosing a sprite, and asking it to would put a rendering
// decision behind a login.
//
// Roles, plural, because they are plural: a user holds N of them and no
// ordering exists to say which one matters. Asking for a single "area" forced
// every implementation to pick arbitrarily, and that decision was invisible.
type Identity interface {
	// UserName is who is logged in.
	UserName() string
	// UserAvatar is the URL of their picture. Empty is normal and expected.
	UserAvatar() string
	// UserRoles are display names, not authorization codes. May be empty.
	UserRoles() []string
}

// Brand is what the platform calls itself in its own chrome. The shell asks for
// a mark and a name; how they are drawn, sized and spaced is platformd's
// business.
//
// It is a READ contract, not a store, mirroring Identity: platformd renders it
// and never mutates it. The consumer supplies facts, not presentation — a
// Brand never hands the shell an svg.Icon, because picking the sprite is a
// rendering decision this package owns (the same reasoning that keeps IconUser
// out of Identity).
type Brand interface {
	// BrandName is shown beside the mark, and is the mark's alt text.
	BrandName() string
	// BrandMark is a URL or inline SVG data URI. Empty is normal and expected:
	// the shell falls back to its own glyph, exactly as UserAvatar does.
	BrandMark() string
}

// Platform is the typed skeleton root.
type Platform struct {
	Element

	// AppName titles the drawer on a phone, where the panel opens wholesale and
	// there is room for it. The collapsed rail never shows it: text with no icon
	// beside it would have to appear from nowhere when the rail expands.
	//
	// It is NOT the header brand: AppName belongs to the phone drawer and Brand
	// belongs to the desktop header, two surfaces of which only one exists at a
	// time. A Brand does not supersede it and it is not a fallback for a missing
	// Brand — a platform without a Brand renders no header brand slot at all.
	AppName string

	// Brand is what the platform calls itself in the header's leading slot:
	// the mark and the name at the header's start, mirrored against the user
	// menu at its end. Optional — a platform without a logo renders no brand
	// slot and the message block stays centred between nothing and the menu.
	Brand Brand

	// User is the logged-in identity. Required: the header's outer thirds and
	// the drawer's first entry are built from it.
	User Identity

	// UserActions slot — shown at the header RIGHT, next to the work-area name
	// (e.g. the light/dark theme toggle). Optional.
	// A factory, not a Component: the shell renders one menu per surface and a
	// single element cannot have two parents. Passing one instance put the same
	// element in both menus, which rendered twice with one id and left the copy
	// in the drawer inert.
	UserActions func() Component

	// Modules registered in order — appearance order in the nav rail.
	Modules []UIModule

	// CanView filters which modules the shell presents. nil = show all.
	CanView func(resource string) bool

	// DefaultID is the ModelName() of the module to show initially.
	// If empty, the first module is used.
	DefaultID string

	// internal state
	active        *SignalString
	menuOpen      *SignalBool
	notifications *SignalNodes
	navIcon       *SignalNodes
	navStowed     *SignalBool

	rawNotifications []notification
	mu               sync.Mutex

	// lastScrollTop es el testigo de la última posición leída por onScroll. Los
	// eventos de scroll llegan por el hilo de JS, igual que los click, así que no
	// va bajo p.mu — ese mutex existe por time.AfterFunc, que descarta
	// notificaciones desde otra goroutine.
	lastScrollTop float64
}

// scrollStowThreshold son los píxeles de desplazamiento que hacen falta para que
// el cromo reaccione. Sin umbral, un píxel de ruido lo haría entrar y salir; y
// como en la página conviven varios scrollers y sus posiciones se intercalan en
// un mismo handler, un salto pequeño puede venir de otro elemento y no de un
// gesto.
const scrollStowThreshold = 8

type notification struct {
	Type MessageType
	Msg  string
	ID   string
}

// Init initializes the platform state and routing.
func (p *Platform) Init(ctx Ctx) {
	p.active = NewString("")
	p.menuOpen = NewBool(false)
	p.notifications = NewNodes()
	p.navIcon = NewNodes()
	p.navStowed = NewBool(false)

	OnHashChange(func(hash string) {
		if len(hash) > 0 && hash[0] == '#' {
			p.Activate(hash[1:])
		}
	})

	// El scroll no burbujea: en captura sobre el documento es la única forma de que
	// el chasis vea el desplazamiento de un contenedor que pertenece a otro paquete.
	OnScrollCapture(func(top float64) {
		p.onScroll(top)
	})

	hash := GetHash()
	if hash != "" && len(hash) > 0 && hash[0] == '#' {
		id := hash[1:]
		if p.isViewable(id) {
			p.Activate(id)
		} else {
			p.fallback()
		}
	} else {
		if p.DefaultID != "" && p.isViewable(p.DefaultID) {
			p.Activate(p.DefaultID)
		} else {
			p.fallback()
		}
	}
}

func (p *Platform) isViewable(id string) bool {
	if p.CanView == nil {
		return true
	}
	return p.CanView(id)
}

// activeIcon es el glifo del módulo en el que estamos. iconMenu es el respaldo:
// UIModule permite que Icon() devuelva la cadena vacía, y un botón sin glifo no
// se puede pulsar porque no se ve.
func (p *Platform) activeIcon() svg.Icon {
	id := p.active.Get()
	for _, m := range p.Modules {
		if m.ModelName() == id {
			if ic := m.Icon(); ic != "" {
				return ic
			}
			break
		}
	}
	return iconMenu
}

// onScroll guarda el botón de menú mientras el usuario baja y lo devuelve en
// cuanto sube. Arriba del todo siempre está a mano: una pantalla que no llega a
// desplazarse — el listado de este demo, sin ir más lejos — dejaría el menú
// inalcanzable si el botón naciera guardado.
//
// Limitación conocida: el handler recibe la posición de CUALQUIER scroller. Si
// dos están en pantalla y se desplazan alternándose, sus posiciones se
// intercalan en un mismo lastScrollTop y el botón puede parpadear. En móvil
// solo hay una columna visible a la vez (MasterDetail enseña una), así que en
// la práctica no ocurre, y el umbral de 8px absorbe el resto.
func (p *Platform) onScroll(top float64) {
	if top <= 0 {
		p.lastScrollTop = 0
		p.navStowed.Set(false)
		return
	}
	switch {
	case top > p.lastScrollTop+scrollStowThreshold:
		p.lastScrollTop = top
		p.navStowed.Set(true)
	case top < p.lastScrollTop-scrollStowThreshold:
		p.lastScrollTop = top
		p.navStowed.Set(false)
	}
}

func (p *Platform) fallback() {
	for _, m := range p.Modules {
		id := m.ModelName()
		if p.isViewable(id) {
			p.Activate(id)
			return
		}
	}
}

// Render builds the DOM tree (implements ViewRenderer).
// userMenu adapts the Identity contract into the component's plain props. The
// component must not know the contract: it lives below layout in the graph, and
// typing Identity there would invert the dependency.
//
// A fresh instance per call on purpose — the shell renders one for the header
// and one for the drawer, and a single *UserMenu cannot have two parents.
// brand builds the header's leading slot from the Brand contract. The mark is
// an <img> when the contract returns a URL and the default glyph otherwise —
// the avatar's exact treatment, mirrored.
func (p *Platform) brand() *Element {
	slot := Div().Set(clsBrand.AsAttr())
	if url := p.Brand.BrandMark(); url != "" {
		slot.Child(NewElement("img").
			Set(clsBrandMark.AsAttr()).
			Attr("src", url).
			Attr("alt", p.Brand.BrandName()).
			Attr("loading", "lazy"))
	} else {
		slot.Child(IconBrand.Render(string(clsBrandMark)))
	}
	slot.Child(Span().Set(clsBrandName.AsAttr()).Text(p.Brand.BrandName()))
	return slot
}

func (p *Platform) userMenu() Component {
	var actions Component
	if p.UserActions != nil {
		actions = p.UserActions()
	}
	return &usermenu.UserMenu{
		Name:     p.User.UserName(),
		Avatar:   p.User.UserAvatar(),
		Roles:    p.User.UserRoles(),
		Fallback: IconUser,
		Actions:  actions,
	}
}

func (p *Platform) Render() *Element {
	root := Div().Set(clsRoot.AsAttr())

	// ── header ───────────────────────────────────────────────────────────────
	// Three parts: who the platform is (brand), what it is telling them
	// (messages), and who is logged in (user menu). None of it echoes the route
	// — the module names itself, and repeating that here says nothing new.
	header := Header().Set(clsHeader.AsAttr())

	// The leading slot: the brand's mark and name. Mirror of the user menu at
	// the other end, so the two edges of the header read as the same kind of
	// object — facts in, rendering owned by the shell.
	if p.Brand != nil {
		header.Child(p.brand())
	}

	msgSlot := Div().Set(clsMsgSlot.AsAttr()).ID("pd-msg-slot").
		BindChildren(p.notifications)
	// Because elementToHTML/SSR doesn't process "children" bindings, initial nodes must be manually added
	for _, n := range p.notifications.Get() {
		msgSlot.Child(n)
	}
	header.Child(msgSlot)

	right := Div().Set(clsHeaderRight.AsAttr())
	if p.User != nil {
		right.Child(p.userMenu())
	}

	header.Child(right)

	root.Child(header)

	// ── hamburger button (mobile only) ───────────────────────────────────────
	// A sibling of the header, not a child of it: on a phone the header is
	// display:none, and a fixed descendant of a hidden ancestor is not rendered
	// either. It stays out of .pd__body, which is what the Sidebar contract
	// there requires.
	hamburger := Button().Set(clsHamburger.AsAttr()).
		Attr("aria-label", "Menu").
		// Open aquí significa "el cromo está desplegado", que es lo mismo que
		// significa en el cajón: un control guardado no está desplegado. La señal
		// dice lo contrario de lo que se pinta, de ahí la negación.
		BindStateFunc(widget.Open, func() bool { return !p.navStowed.Get() }).
		BindChildren(p.navIcon).
		ID("pd-hamburger-btn")
	hamburger.On("click", func(Event) {
		p.menuOpen.Toggle()
	})
	root.Child(hamburger)

	// ── nav overlay backdrop (mobile) ────────────────────────────────────────
	overlay := Div().Set(clsNavOverlay.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return p.menuOpen.Get() })
	overlay.On("click", func(Event) {
		p.menuOpen.Set(false)
	})
	root.Child(overlay)

	// ── body (Sidebar wrapping stage and nav) ────────────────────────────────
	body := Div().Set(clsBody.AsAttr())

	// ── main stage ───────────────────────────────────────────────────────────
	stage := Main().Set(clsStage.AsAttr())

	for _, m := range p.Modules {
		m := m
		id := m.ModelName()
		if !p.isViewable(id) {
			continue
		}
		panel := Section().Set(clsPanel.AsAttr()).
			ID(id).
			Attr("data-id", id).
			BindStateFunc(widget.Current, func() bool { return p.active.Get() == id })

		if v := m.View(); v != nil {
			panel.Child(v)
		}
		stage.Child(panel)
	}
	body.Child(stage)

	// ── navigation menu (rail) ───────────────────────────────────────────────
	nav := Nav().Set(clsMenu.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return p.menuOpen.Get() })
	navbar := Ul().Set(clsNavbar.AsAttr())

	for _, m := range p.Modules {
		m := m
		id := m.ModelName()
		if !p.isViewable(id) {
			continue
		}
		link := A("#"+id).Set(clsNavLink.AsAttr()).
			Attr("data-id", id).
			BindStateFunc(widget.Current, func() bool { return p.active.Get() == id })

		if icon := m.Icon(); icon != "" {
			link.Child(icon.Render(string(ClsNavIcon)))
		}
		link.Child(Span().Set(clsLinkText.AsAttr()).Text(m.Label()))

		navbar.Child(Li().Set(clsNavItem.AsAttr()).Child(link))
	}

	// The drawer's head. On a phone this is where the identity chrome lives —
	// there is no header to hold it — and on a wide screen it rides above the
	// rail, revealed with it on hover.
	// One panel holding the head and the navbar. On a wide screen the whole
	// panel is what floats out over the content on hover — if the head expanded
	// inside the rail's flow instead, the rail would widen and push the stage.
	drawerPanel := Div().Set(clsDrawerPanel.AsAttr())

	// Only on a phone, and deliberately with no CueWithin to reveal it on a
	// wide screen: the drawer opens wholesale on a phone so a line of text
	// costs nothing, while in the collapsed rail it would have to materialise
	// when the rail expands and push everything under it down.
	if p.AppName != "" {
		drawerPanel.Child(Div().Set(clsAppName.AsAttr()).Text(p.AppName))
	}

	// A second menu, for the phone: there is no header there to hold the first
	// one. This costs nothing now that the contract returns strings — two
	// elements built from the same facts, of which exactly one is ever visible.
	// It was impossible while identity was a Component slot: one element has
	// one parent.
	// Wrapped in a part this package can switch off: the component's own class
	// belongs to the component, and the shell needs somewhere to say "not on a
	// wide screen, where the header already has one".
	if p.User != nil {
		drawerPanel.Child(Div().Set(clsDrawerIdentity.AsAttr()).Child(p.userMenu()))
	}

	drawerPanel.Child(navbar)
	nav.Child(drawerPanel)
	body.Child(nav)

	root.Child(body)

	return root
}

func (p *Platform) buildToasts() []*Element {
	p.mu.Lock()
	defer p.mu.Unlock()

	nodes := make([]*Element, 0, len(p.rawNotifications))
	for _, n := range p.rawNotifications {
		var variantCls widget.Class
		switch n.Type {
		case Msg.Info:
			variantCls = clsMsgInfo
		case Msg.Success:
			variantCls = clsMsgSuccess
		case Msg.Warning:
			variantCls = clsMsgWarning
		case Msg.Error:
			variantCls = clsMsgError
		default:
			variantCls = clsMsgInfo
		}

		nodes = append(nodes, Div().Set(clsMsg.AsAttr(), variantCls.AsAttr()).
			ID(n.ID).
			Key(n.ID).
			Text(n.Msg))
	}
	return nodes
}

// Notify queues a typed notification in the proper viewport slot.
// Any non-zero durationMs → schedule dismissal; duration 0 → persistent message.
func (p *Platform) Notify(t MessageType, msg string, durationMs int) {
	p.mu.Lock()
	n := notification{
		Type: t,
		Msg:  msg,
		ID:   "pd-notification-" + p.GetID() + "-" + Sprint(time.Now()),
	}
	p.rawNotifications = append(p.rawNotifications, n)
	p.mu.Unlock()

	p.notifications.Set(p.buildToasts())

	if durationMs > 0 {
		time.AfterFunc(durationMs, func() {
			p.dismiss(n.ID)
		})
	}
}

func (p *Platform) dismiss(id string) {
	p.mu.Lock()
	for i, n := range p.rawNotifications {
		if n.ID == id {
			p.rawNotifications = append(p.rawNotifications[:i], p.rawNotifications[i+1:]...)
			p.mu.Unlock()
			p.notifications.Set(p.buildToasts())
			return
		}
	}
	p.mu.Unlock()
}

func (p *Platform) notificationCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.rawNotifications)
}

// Activate programmatically switches to a module by ID
// (also updates window.location.hash on wasm builds).
func (p *Platform) Activate(moduleID string) {
	if !p.isViewable(moduleID) {
		return
	}

	if p.active.Get() == moduleID && !p.menuOpen.Get() {
		return
	}

	p.active.Set(moduleID)
	p.menuOpen.Set(false)

	// El botón de menú lleva el estado de la navegación: en móvil no hay cabecera
	// ni rail visible, así que su glifo es lo único que dice en qué sección estás.
	p.navIcon.Set([]*Element{p.activeIcon().Render(string(ClsNavIcon)).Key(moduleID)})

	// Cambiar de sección reinicia el cromo: el módulo nuevo empieza desde arriba y
	// con el botón a mano.
	p.lastScrollTop = 0
	p.navStowed.Set(false)

	// Update window hash if needed
	if GetHash() != "#"+moduleID {
		SetHash("#" + moduleID)
	}
}
