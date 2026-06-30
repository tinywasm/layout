package platformd

import (
	"sync"

	. "github.com/tinywasm/css"
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	. "github.com/tinywasm/html"
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

// Module describes one registered route/page in the shell.
// Pure data: the consumer creates these and passes them to Platform.
type Module struct {
	ID      string    // hash slug, e.g. "products" → routed via "#products"
	Label   string    // text shown next to icon in expanded rail
	Icon    Component // icon component (usually a tinywasm/icons Symbol)
	View    Component // the module's content (often a *rightpanel.RightPanel)
	Default bool      // if true and no hash set, this module is shown initially
}

// Platform is the typed skeleton root.
type Platform struct {
	Element

	// AppName appears in the header (left side, near UserBlock).
	AppName string

	// UserBlock slot — usually an avatar/name/logout link. Optional.
	UserBlock Component

	// Modules registered in order — appearance order in the nav rail.
	Modules []Module

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
		foundDefault := false
		for _, m := range p.Modules {
			if m.Default {
				p.Activate(m.ID)
				foundDefault = true
				break
			}
		}
		if !foundDefault && len(p.Modules) > 0 {
			p.Activate(p.Modules[0].ID)
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
				if m.ID == id {
					return m.Label
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

	for _, mod := range p.Modules {
		mod := mod
		link := A("#"+mod.ID).Set(clsNavLink.AsAttr()).
			Attr("data-id", mod.ID).
			BindClass(string(clsNavActive), DeriveBool(func() bool {
				return p.active.Get() == mod.ID
			}))

		if mod.Icon != nil {
			link.Child(mod.Icon)
		}
		link.Child(Span().Set(clsLinkText.AsAttr()).Text(mod.Label))

		navbar.Child(Li().Set(clsNavItem.AsAttr()).Child(link))
	}

	nav.Child(navbar)
	root.Child(nav)

	// ── main stage ───────────────────────────────────────────────────────────
	stage := Main().Set(clsStage.AsAttr())

	for _, mod := range p.Modules {
		mod := mod
		panel := Section().Set(clsPanel.AsAttr()).
			ID(mod.ID).
			Attr("data-id", mod.ID).
			BindClass(string(clsPanelActive), DeriveBool(func() bool {
				return p.active.Get() == mod.ID
			}))

		if mod.View != nil {
			panel.Child(mod.View)
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
