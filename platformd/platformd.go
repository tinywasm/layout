package platformd

import (
	"sync"

	"github.com/tinywasm/layout"
	"github.com/tinywasm/svg"
	"github.com/tinywasm/widget"

	. "github.com/tinywasm/css"
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	. "github.com/tinywasm/html"
	"github.com/tinywasm/time"
)

const NamePlatform widget.Name = "pd"

var (
	clsRoot            = NamePlatform.Root()
	clsHeader          = NamePlatform.Class("header")
	clsUserBlock       = NamePlatform.Class("user-block")
	clsHeaderRight     = NamePlatform.Class("header-right")
	clsMsgDesktop      = NamePlatform.Class("msg-desktop")
	clsArea            = NamePlatform.Class("area")
	clsMsgMobile       = NamePlatform.Class("msg-mobile")
	clsMenu            = NamePlatform.Class("menu")
	clsNavbar          = NamePlatform.Class("navbar")
	clsNavItem         = NamePlatform.Class("nav-item")
	clsNavLink         = NamePlatform.Class("nav-link")
	clsLinkText        = NamePlatform.Class("link-text")
	ClsNavIcon         = NamePlatform.Class("nav-icon")
	clsNavActive       = NamePlatform.Class("nav-active")
	clsStage           = NamePlatform.Class("stage")
	clsPanel           = NamePlatform.Class("panel")
	clsPanelActive     = NamePlatform.Class("panel-active")
	clsOrientationWarn = NamePlatform.Class("orientation-warn")
	clsMsg             = NamePlatform.Class("msg")
	clsHamburger       = NamePlatform.Class("hamburger")
	clsNavOverlay      = NamePlatform.Class("nav-overlay")
	clsMenuOpen        = NamePlatform.Class("menu-open")
)

const (
	IconHome     = svg.Icon("home")
	IconProducts = svg.Icon("products")
	IconInfo     = svg.Icon("info")
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

// Platform is the typed skeleton root.
type Platform struct {
	Element

	// AppName appears in the header (left side, near UserBlock).
	AppName string

	// UserBlock slot — the logged-in user (name/avatar), shown at the header LEFT.
	UserBlock Component

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
	root := Div().Set(clsRoot.AsAttr()).
		BindClass(string(clsMenuOpen), p.menuOpen)

	// ── header ───────────────────────────────────────────────────────────────
	header := Header().Set(clsHeader.AsAttr())

	userBlock := Div().Set(clsUserBlock.AsAttr())
	if p.UserBlock != nil {
		userBlock.Child(p.UserBlock)
	}
	header.Child(userBlock)

	msgDesktop := Div().Set(clsMsgDesktop.AsAttr()).ID("pd-msg-desktop").
		BindChildren(p.notifications)
	for _, n := range p.notifications.Get() {
		msgDesktop.Child(n)
	}
	header.Child(msgDesktop)

	// header right: work-area name + actions (theme toggle) grouped together.
	right := Div().Set(clsHeaderRight.AsAttr())
	right.Child(H2().Set(clsArea.AsAttr()).
		BindText(DeriveString(func() string {
			id := p.active.Get()
			for _, m := range p.Modules {
				if m.ModelName() == id {
					return m.Label()
				}
			}
			return ""
		})))
	if p.HeaderActions != nil {
		right.Child(p.HeaderActions)
	}
	header.Child(right)

	root.Child(header)

	// ── mobile message slot ──────────────────────────────────────────────────
	msgMobile := Div().Set(clsMsgMobile.AsAttr()).ID("pd-msg-mobile").
		BindChildren(p.notifications)
	for _, n := range p.notifications.Get() {
		msgMobile.Child(n)
	}
	root.Child(msgMobile)

	// ── hamburger button (mobile only — hidden via CSS on desktop) ───────────
	hamburger := Button().Set(clsHamburger.AsAttr()).
		Attr("aria-label", "Menú").
		Child(Span(), Span(), Span())
	hamburger.On("click", func(Event) {
		p.menuOpen.Toggle()
	})
	root.Child(hamburger)

	// ── nav overlay backdrop (mobile) ────────────────────────────────────────
	overlay := Div().Set(clsNavOverlay.AsAttr())
	overlay.On("click", func(Event) {
		p.menuOpen.Set(false)
	})
	root.Child(overlay)

	// ── navigation menu ──────────────────────────────────────────────────────
	nav := Nav().Set(clsMenu.AsAttr())
	navbar := Ul().Set(clsNavbar.AsAttr())

	for _, m := range p.Modules {
		m := m
		id := m.ModelName()
		if !p.isViewable(id) {
			continue
		}
		link := A("#"+id).Set(clsNavLink.AsAttr()).
			Attr("data-id", id).
			BindClass(string(clsNavActive), DeriveBool(func() bool {
				return p.active.Get() == id
			}))

		if icon := m.Icon(); icon != "" {
			link.Child(icon.Render(string(ClsNavIcon)))
		}
		link.Child(Span().Set(clsLinkText.AsAttr()).Text(m.Label()))

		navbar.Child(Li().Set(clsNavItem.AsAttr()).Child(link))
	}

	nav.Child(navbar)
	root.Child(nav)

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
			BindClass(string(clsPanelActive), DeriveBool(func() bool {
				return p.active.Get() == id
			}))

		if v := m.View(); v != nil {
			panel.Child(v)
		}
		stage.Child(panel)
	}

	root.Child(stage)

	// orientation warning (placeholder as per PLAN.md A.5)
	root.Child(Div().Set(clsOrientationWarn.AsAttr()))

	return root
}

func (p *Platform) buildToasts() []*Element {
	p.mu.Lock()
	defer p.mu.Unlock()

	nodes := make([]*Element, 0, len(p.rawNotifications))
	for _, n := range p.rawNotifications {
		typeCls := "pd-msg-" + Convert(n.Type.String()).ToLower().String()
		nodes = append(nodes, Div().Set(clsMsg.AsAttr(), Class(typeCls).AsAttr()).
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

	// Update window hash if needed
	if GetHash() != "#"+moduleID {
		SetHash("#" + moduleID)
	}
}
