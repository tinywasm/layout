package platformd

import (
	"sync"

	"github.com/tinywasm/layout"

	. "github.com/tinywasm/css"
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	. "github.com/tinywasm/html"
	"github.com/tinywasm/svg"
	"github.com/tinywasm/time"
)

var (
	clsRoot            Class = "pd-root"
	clsHeader          Class = "pd-header"
	clsUserBlock       Class = "pd-user-block"
	clsMsgDesktop      Class = "pd-msg-desktop"
	clsArea            Class = "pd-area"
	clsMsgMobile       Class = "pd-msg-mobile"
	clsMenu            Class = "pd-menu"
	clsNavbar          Class = "pd-navbar"
	clsNavItem         Class = "pd-nav-item"
	clsNavLink         Class = "pd-nav-link"
	clsLinkText        Class = "pd-link-text"
	ClsNavIcon         Class = "pd-nav-icon"
	clsNavActive       Class = "pd-nav-active"
	clsStage           Class = "pd-stage"
	clsPanel           Class = "pd-panel"
	clsPanelActive     Class = "pd-panel-active"
	clsOrientationWarn Class = "pd-orientation-warn"
	clsMsg             Class = "pd-msg"
	clsHamburger       Class = "pd-hamburger"
	clsNavOverlay      Class = "pd-nav-overlay"
	clsMenuOpen        Class = "pd-menu-open"
)

// UIModule is a module that provides its UI to the platform chassis.
// The chassis takes id/hash/route from ModelName() and the rest of the
// presentation from these methods.
type UIModule interface {
	layout.Module    // identity: ModelName() → used as ID
	Label() string   // text in the nav rail
	Icon() svg.Icon  // icon (unrendered); chassis paints it with ClsNavIcon
	View() Component // module content (often a *rightpanel.RightPanel)
}

// Platform is the typed skeleton root.
type Platform struct {
	Element

	// AppName appears in the header (left side, near UserBlock).
	AppName string

	// UserBlock slot — usually an avatar/name/logout link. Optional.
	UserBlock Component

	// Modules registered in order — appearance order in the nav rail.
	Modules []UIModule

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
		p.Activate(hash[1:])
	} else {
		if p.DefaultID != "" {
			p.Activate(p.DefaultID)
		} else if len(p.Modules) > 0 {
			p.Activate(p.Modules[0].ModelName())
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

	header.Child(H2().Set(clsArea.AsAttr()).
		BindText(DeriveString(func() string {
			id := p.active.Get()
			for _, m := range p.Modules {
				if m.ModelName() == id {
					return m.Label()
				}
			}
			return ""
		})))

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
		link := A("#"+id).Set(clsNavLink.AsAttr()).
			Attr("data-id", id).
			BindClass(string(clsNavActive), DeriveBool(func() bool {
				return p.active.Get() == id
			}))

		ico := m.Icon()
		if ico.ID() != "" {
			link.Child(ico.Render(string(ClsNavIcon)))
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
