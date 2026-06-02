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
	clsNavIcon         Class = "pd-nav-icon"
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
	activeModuleID string
	menuOpen       bool

	notifications []notification
	mu            sync.Mutex
}

type notification struct {
	Type MessageType
	Msg  string
	ID   string
}

// OnMount implements Mountable.
func (p *Platform) OnMount() {
	OnHashChange(func(hash string) {
		if len(hash) > 0 && hash[0] == '#' {
			p.Activate(hash[1:])
		}
	})

	hash := GetHash()
	if hash != "" && len(hash) > 0 && hash[0] == '#' {
		p.Activate(hash[1:])
	} else {
		// Pick default or first
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
	root := Div(clsRoot.AsAttr())
	if p.menuOpen {
		root.Add(clsMenuOpen.AsAttr())
	}

	activeLabel := ""
	for _, mod := range p.Modules {
		if mod.ID == p.activeModuleID {
			activeLabel = mod.Label
			break
		}
	}

	// ── header ───────────────────────────────────────────────────────────────
	header := Header(clsHeader.AsAttr())

	userBlock := Div(clsUserBlock.AsAttr())
	if p.UserBlock != nil {
		userBlock.Add(p.UserBlock)
	}
	header.Add(userBlock)

	msgDesktop := Div(clsMsgDesktop.AsAttr()).ID("pd-msg-desktop")
	for _, n := range p.notifications {
		msgDesktop.Add(p.renderNotification(n))
	}
	header.Add(msgDesktop)

	header.Add(H2(clsArea.AsAttr()).Text(activeLabel))

	root.Add(header)

	// ── mobile message slot ──────────────────────────────────────────────────
	msgMobile := Div(clsMsgMobile.AsAttr()).ID("pd-msg-mobile")
	for _, n := range p.notifications {
		msgMobile.Add(p.renderNotification(n))
	}
	root.Add(msgMobile)

	// ── hamburger button (mobile only — hidden via CSS on desktop) ───────────
	hamburger := Button(clsHamburger.AsAttr()).
		Attr("aria-label", "Menú").
		Add(Span(), Span(), Span())
	hamburger.On("click", func(Event) {
		p.mu.Lock()
		p.menuOpen = !p.menuOpen
		p.mu.Unlock()
		p.Update()
	})
	root.Add(hamburger)

	// ── nav overlay backdrop (mobile) ────────────────────────────────────────
	overlay := Div(clsNavOverlay.AsAttr())
	overlay.On("click", func(Event) {
		p.mu.Lock()
		p.menuOpen = false
		p.mu.Unlock()
		p.Update()
	})
	root.Add(overlay)

	// ── navigation menu ──────────────────────────────────────────────────────
	nav := Nav(clsMenu.AsAttr())
	navbar := Ul(clsNavbar.AsAttr())

	for _, mod := range p.Modules {
		item := Li(clsNavItem.AsAttr())
		link := A("#"+mod.ID, clsNavLink.AsAttr()).
			Attr("data-id", mod.ID)

		if mod.ID == p.activeModuleID {
			link.Add(clsNavActive.AsAttr())
		}

		if mod.Icon != nil {
			link.Add(mod.Icon)
		}
		link.Add(Span(clsLinkText.AsAttr()).Text(mod.Label))

		item.Add(link)
		navbar.Add(item)
	}

	nav.Add(navbar)
	root.Add(nav)

	// ── main stage ───────────────────────────────────────────────────────────
	stage := Main(clsStage.AsAttr())

	for _, mod := range p.Modules {
		panel := Section(clsPanel.AsAttr()).
			ID(mod.ID).
			Attr("data-id", mod.ID)

		if mod.ID == p.activeModuleID {
			panel.Add(clsPanelActive.AsAttr())
		}

		if mod.View != nil {
			panel.Add(mod.View)
		}
		stage.Add(panel)
	}

	root.Add(stage)

	// orientation warning (placeholder as per PLAN.md A.5)
	root.Add(Div(clsOrientationWarn.AsAttr()))

	return root
}

func (p *Platform) renderNotification(n notification) *Element {
	typeCls := "pd-msg-" + Convert(n.Type.String()).ToLower().String()
	return Div(clsMsg.AsAttr(), Class(typeCls)).ID(n.ID).Text(n.Msg)
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
	p.notifications = append(p.notifications, n)
	p.mu.Unlock()

	p.Update()

	if durationMs > 0 {
		time.AfterFunc(durationMs, func() {
			p.dismiss(n.ID)
		})
	}
}

func (p *Platform) dismiss(id string) {
	p.mu.Lock()
	for i, n := range p.notifications {
		if n.ID == id {
			p.notifications = append(p.notifications[:i], p.notifications[i+1:]...)
			p.mu.Unlock()
			p.Update()
			return
		}
	}
	p.mu.Unlock()
}

func (p *Platform) notificationCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.notifications)
}

// Activate programmatically switches to a module by ID
// (also updates window.location.hash on wasm builds).
func (p *Platform) Activate(moduleID string) {
	if p.activeModuleID == moduleID && !p.menuOpen {
		return
	}

	p.activeModuleID = moduleID
	p.menuOpen = false

	// Update window hash if needed
	if GetHash() != "#"+moduleID {
		SetHash("#" + moduleID)
	}

	p.Update()
}
