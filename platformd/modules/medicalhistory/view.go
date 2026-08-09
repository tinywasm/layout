package medicalhistory

import (
	"github.com/tinywasm/components/selectsearch"
	"github.com/tinywasm/layout/rightpanel"

	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/html"
)

func agendaOptions() []selectsearch.SsOption {
	opts := make([]selectsearch.SsOption, len(todayAgenda))
	for i, p := range todayAgenda {
		opts[i] = selectsearch.SsOption{ID: p.ID, Label: p.Name, Description: p.Time}
	}
	return opts
}

// View is the module's content. Embeds Element as a VALUE (never a
// pointer — TinyGo heap constraint, see AGENTS.md).
type View struct {
	Element

	picker     selectsearch.SelectSearch
	selectedID *SignalString
	rows       *SignalNodes
}

func (v *View) Init(ctx Ctx) {
	v.selectedID = NewString("")
	v.rows = NewNodes(v.emptyState()...)

	v.picker.Placeholder = "Seleccione un paciente..."
	v.picker.Options = agendaOptions()
	// The point of this whole demo: SelectSearch now satisfies
	// widget.Filterable, so picking a patient reaches this view through the
	// SAME generic OnFilterChange(term string) contract a
	// *searchbar.SearchBar would use in a crudview.Filter slot — not a
	// bespoke OnSelect callback. See components/docs/PLAN.md Stage 1.
	v.picker.OnFilterChange(func(id string) { v.selectPatient(id) })
}

func (v *View) selectPatient(id string) {
	v.selectedID.Set(id)
	v.rows.Set(v.historyRows(id))
}

func (v *View) emptyState() []*Element {
	return []*Element{Li().Text("Seleccione un paciente de la agenda para ver su historial.")}
}

func (v *View) historyRows(patientID string) []*Element {
	history := historyFor(patientID)
	if len(history) == 0 {
		return []*Element{Li().Text("Sin antecedentes registrados.")}
	}
	nodes := make([]*Element, len(history))
	for i, visit := range history {
		nodes[i] = Li().Key(patientID + "-" + visit.Date).Child(
			Span().Text(visit.Date),
			Span().Text(visit.Reason),
			Span().Text(visit.Diagnosis),
		)
	}
	return nodes
}

// agendaRows renders today's full agenda as plain clickable rows — a
// SECOND entry point into selectPatient, alongside the SelectSearch above
// it, deliberately: it proves selectPatient (and therefore
// OnFilterChange's narrowing) is reachable from any control shape, not
// just a combobox. widget.Filterable's own doc comment names "a search
// bar, a date picker, a category select" as interchangeable Filterable
// sources; a plain row click is the same idea, it just skips the search
// step.
func (v *View) agendaRows() []*Element {
	nodes := make([]*Element, len(todayAgenda))
	for i, p := range todayAgenda {
		p := p
		nodes[i] = Li().Key(p.ID).
			Text(p.Time + " — " + p.Name).
			On("click", func(Event) { v.selectPatient(p.ID) })
	}
	return nodes
}

func (v *View) Render() *Element {
	agenda := Ul()
	for _, n := range v.agendaRows() {
		agenda.Child(n)
	}

	history := Ul().BindChildren(v.rows)
	// elementToHTML/SSR does not process "children" bindings — initial
	// nodes must be added manually too. Same gotcha platformd.go's msgSlot
	// works around; see that file's comment for the full explanation.
	for _, n := range v.rows.Get() {
		history.Child(n)
	}

	panel := &rightpanel.RightPanel{
		Title:        "Agenda de hoy",
		HeadControls: &v.picker,
		Article:      agenda,
		Aside:        history,
	}
	return panel.Render()
}
