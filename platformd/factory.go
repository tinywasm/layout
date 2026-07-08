package platformd

import (
	. "github.com/tinywasm/dom"
)

// NewUIModule returns a private implementation of UIModule.
func NewUIModule(id, label, iconID string, view Component) UIModule {
	return &uiModule{
		id:     id,
		label:  label,
		iconID: iconID,
		view:   view,
	}
}

type uiModule struct {
	id     string
	label  string
	iconID string
	view   Component
}

func (m *uiModule) ModelName() string { return m.id }
func (m *uiModule) Label() string     { return m.label }
func (m *uiModule) IconID() string    { return m.iconID }
func (m *uiModule) View() Component   { return m.view }
