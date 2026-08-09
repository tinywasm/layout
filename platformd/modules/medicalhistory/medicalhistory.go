// Package medicalhistory is a demo module: selectsearch.SelectSearch (now
// widget.Filterable — see github.com/tinywasm/components v0.5.0) drives a
// rightpanel.RightPanel master-detail. Pick a patient from today's agenda,
// their history loads on the other side. It is the pattern to copy for a
// real "select one, load its detail" module — the same role
// platformd/modules/devices plays for full CRUD.
package medicalhistory

import (
	"github.com/tinywasm/layout/platformd"
	"github.com/tinywasm/svg"

	. "github.com/tinywasm/dom"
)

const Icon = svg.Icon("mod-medicalhistory")

// Module is the platformd.UIModule. It carries no state of its own — View()
// builds a fresh *View per call, which is where the demo's state (the
// signals, the fake agenda) actually lives.
type Module struct{}

func New() *Module { return &Module{} }

var _ platformd.UIModule = (*Module)(nil)

func (m *Module) ModelName() string { return "medicalhistory" }
func (m *Module) Label() string     { return "Historial Médico" }
func (m *Module) Icon() svg.Icon    { return Icon }
func (m *Module) View() Component   { return &View{} }
