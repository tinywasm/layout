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
	IconHome     = svg.Icon("home")
	IconProducts = svg.Icon("products")
	IconInfo     = svg.Icon("info")
	IconUser     = svg.Icon("pd-user")
	iconMenu     = svg.Icon("pd-menu")
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

// Platform is the typed skeleton root.
type Platform struct {
	Element

	// AppName titles the drawer on a phone, where the panel opens wholesale and
	// there is room for it. The collapsed rail never shows it: text with no icon
	// beside it would have to appear from nowhere when the rail expands.
	AppName string

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

	rawNotifications []notification
	mu               sync.Mutex
}

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

	OnHashChange(func(hash string) {
		if len(hash) > 0 && hash[0] == '#' {
			p.Activate(hash[1:])
		}
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
	// Three parts: who is logged in, what the platform is telling them, and the
	// area they are working in. None of it echoes the route — the module names
	// itself, and repeating that here says nothing new.
	header := Header().Set(clsHeader.AsAttr())

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
		Attr("aria-label", "Menú").
		Child(iconMenu.Render(string(ClsNavIcon)))
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

	// The stage is a Deck: the panels are all mounted side by side and this is
	// what slides between them. `display` is discrete and cannot transition, so
	// the movement has to come from the scroller.
	if el, ok := Get(moduleID); ok {
		el.ScrollIntoView()
	}

	// Update window hash if needed
	if GetHash() != "#"+moduleID {
		SetHash("#" + moduleID)
	}
}
