package platformd

import (
	"sync"

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
	clsRoot          = NamePlatform.Root()
	clsHeader        = NamePlatform.Class("header")
	clsUserName      = NamePlatform.Class("user-name")
	clsDrawerUser    = NamePlatform.Class("drawer-user")
	clsDrawerArea    = NamePlatform.Class("drawer-area")
	clsDrawerPanel   = NamePlatform.Class("drawer-panel")
	clsDrawerHead    = NamePlatform.Class("drawer-head")
	clsAppName       = NamePlatform.Class("app-name")
	clsDrawerActions = NamePlatform.Class("drawer-actions")
	clsMsgSlot       = NamePlatform.Class("msg-slot")
	clsMsg           = NamePlatform.Class("msg")
	clsMsgInfo       = NamePlatform.Class("msg-info")
	clsMsgSuccess    = NamePlatform.Class("msg-success")
	clsMsgWarning    = NamePlatform.Class("msg-warning")
	clsMsgError      = NamePlatform.Class("msg-error")
	clsHeaderRight   = NamePlatform.Class("header-right")
	clsArea          = NamePlatform.Class("area")
	clsBody          = NamePlatform.Class("body")
	clsStage         = NamePlatform.Class("stage")
	clsPanel         = NamePlatform.Class("panel")
	clsMenu          = NamePlatform.Class("menu")
	clsNavbar        = NamePlatform.Class("navbar")
	clsNavItem       = NamePlatform.Class("nav-item")
	clsNavLink       = NamePlatform.Class("nav-link")
	clsLinkText      = NamePlatform.Class("link-text")
	ClsNavIcon       = NamePlatform.Class("nav-icon")
	clsHamburger     = NamePlatform.Class("hamburger")
	clsNavOverlay    = NamePlatform.Class("nav-overlay")
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
// Every method has a place in the chrome: the name and the area are the
// header's outer thirds on a wide screen and the drawer's first entry on a
// phone, and the icon is what that entry shows while the rail is collapsed.
type Identity interface {
	// UserName is who is logged in.
	UserName() string
	// UserArea is the area they are working in — a department, a tenant, a
	// role. It is NOT the current route: the module already names itself.
	UserArea() string
	// UserIcon is the glyph that stands for them in the collapsed rail.
	UserIcon() svg.Icon
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

	// HeaderActions slot — shown at the header RIGHT, next to the work-area name
	// (e.g. the light/dark theme toggle). Optional.
	HeaderActions Component

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
func (p *Platform) Render() *Element {
	root := Div().Set(clsRoot.AsAttr())

	// Grab the state attributes for exact string matches
	open := widget.Open.Attr()
	cur := widget.Current.Attr()

	// ── header ───────────────────────────────────────────────────────────────
	// Three parts: who is logged in, what the platform is telling them, and the
	// area they are working in. None of it echoes the route — the module names
	// itself, and repeating that here says nothing new.
	header := Header().Set(clsHeader.AsAttr())

	if p.User != nil {
		header.Child(Div().Set(clsUserName.AsAttr()).Text(p.User.UserName()))
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
		right.Child(Div().Set(clsArea.AsAttr()).Text(p.User.UserArea()))
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
		BindAttrFunc(open.Key, func() string {
			if p.menuOpen.Get() {
				return open.Value
			}
			return ""
		})
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
			BindAttrFunc(cur.Key, func() string {
				if p.active.Get() == id {
					return cur.Value
				}
				return ""
			})

		if v := m.View(); v != nil {
			panel.Child(v)
		}
		stage.Child(panel)
	}
	body.Child(stage)

	// ── navigation menu (rail) ───────────────────────────────────────────────
	nav := Nav().Set(clsMenu.AsAttr()).
		BindAttrFunc(open.Key, func() string {
			if p.menuOpen.Get() {
				return open.Value
			}
			return ""
		})
	navbar := Ul().Set(clsNavbar.AsAttr())

	for _, m := range p.Modules {
		m := m
		id := m.ModelName()
		if !p.isViewable(id) {
			continue
		}
		link := A("#"+id).Set(clsNavLink.AsAttr()).
			Attr("data-id", id).
			BindAttrFunc(cur.Key, func() string {
				if p.active.Get() == id {
					return cur.Value
				}
				return ""
			})

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

	// AppName only on a phone: the drawer opens wholesale there, so a line of
	// text appearing costs nothing. In the collapsed rail it would have to
	// materialise out of nowhere when the rail expands, and push everything
	// under it down — the exact shift the user entry below is shaped to avoid.
	if p.AppName != "" {
		drawerPanel.Child(Div().Set(clsAppName.AsAttr()).Text(p.AppName))
	}

	// The user entry is ALWAYS here, icon and all, exactly like a nav item.
	// Every platform behind a login has one, so it is not chrome that comes and
	// goes: revealing it only on expansion inserted a fourth element above three
	// icons and shoved them down, then took it away again.
	if p.User != nil {
		head := Div().Set(clsDrawerHead.AsAttr())
		head.Child(p.User.UserIcon().Render(string(ClsNavIcon)))
		head.Child(Div().Set(clsDrawerUser.AsAttr()).Text(p.User.UserName()))
		head.Child(Div().Set(clsDrawerArea.AsAttr()).Text(p.User.UserArea()))
		drawerPanel.Child(head)
	}

	if p.HeaderActions != nil {
		drawerPanel.Child(Div().Set(clsDrawerActions.AsAttr()).Child(p.HeaderActions))
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
