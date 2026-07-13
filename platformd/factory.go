package platformd

import (
	. "github.com/tinywasm/dom"
	"github.com/tinywasm/svg"
)

// NewUIModule returns a private implementation of UIModule.
func NewUIModule(id, label string, icon svg.Icon, view Component) UIModule {
	return &uiModule{
		id:    id,
		label: label,
		icon:  icon,
		view:  view,
	}
}

type uiModule struct {
	id    string
	label string
	icon  svg.Icon
	view  Component
}

func (m *uiModule) ModelName() string { return m.id }
func (m *uiModule) Label() string     { return m.label }
func (m *uiModule) Icon() svg.Icon    { return m.icon }
func (m *uiModule) View() Component   { return m.view }
