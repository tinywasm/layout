---
PLAN: "feat(platformd): medicalhistory demo — SelectSearch-driven agenda/history master-detail"
TAG: v0.1.12
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 761761415412234203
PR: https://github.com/tinywasm/layout/pull/24
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Stage B of a 2-repo change — DO NOT DISPATCH before Stage A is merged and
> published.**
>
> | | Repo | Plan | What |
> |---|---|---|---|
> | A | `components` | `components/docs/PLAN.md` | `selectsearch.SelectSearch` satisfies `widget.Filterable`; its dropdown gets a mobile/desktop-anchored fix |
> | **B** | **`layout`** | **this plan** | a new demo module (`platformd/modules/medicalhistory`) that visually exercises what A built, before either capability is wired into a real product module |

## 0. Dependency gate — verify before writing any code

```bash
go get github.com/tinywasm/components@v0.5.0
go doc github.com/tinywasm/components/selectsearch OnFilterChange
```

The second command must print the method signature
(`func (c *SelectSearch) OnFilterChange(fn func(term string))`). If it
errors ("no such symbol"), **stop** — Stage A is not published yet, this
plan cannot proceed.

---

## 1. Why this module exists

The user wants to see, in a running app, the exact pattern a real feature
will use before committing to it: **a doctor opens their daily agenda,
picks a patient, and that patient's medical history loads** — the same
shape as `http://192.168.122.10:1100/#medicalhistory` in the legacy
reference system (list of today's patients on one side, detail on the
other).

This module is a **visual prototype only**, following the exact precedent
`platformd/modules/devices` already set for this repo (its own doc comment:
*"el módulo demo de un CRUD completo... es el patrón a copiar"*) — except
this one demos "pick one, load its detail" instead of full CRUD. Every
patient/visit record here is **fake, package-level, in-memory data**. A
real integration replaces it with `router.Caller` calls into
`appointment_booking` (today's reservations for a staff member) and
`clinical_encounter` (`ListVisitsByPatient`) — this repo (`tinywasm/layout`)
never imports either, since those are `veltylabs/*` business modules and
`layout` is a generic UI-kit one level below them in the dependency graph.
Say this explicitly in the code (see Stage 2) so nobody mistakes the fake
slices for real wiring later.

**Layout primitive used**: `tinywasm/layout/rightpanel.RightPanel` — the
same two-column (`Article` main / `Aside` detail) skeleton `about` and
`devices` already use, which is why this prototype gets responsive
desktop/mobile master-detail behavior (`MasterDetail` scroll-snap) for
free, with zero new CSS in this repo.

---

## 2. New package `platformd/modules/medicalhistory`

Create the directory `platformd/modules/medicalhistory/` with 3 files.

### 2a. `medicalhistory.go`

```go
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
```

### 2b. `view.go`

```go
package medicalhistory

import (
	"github.com/tinywasm/components/selectsearch"
	"github.com/tinywasm/layout/rightpanel"

	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/html"
)

// Patient and Visit are LOCAL FAKE types for this demo only. A real
// deployment replaces todayAgenda/visits below with router.Caller calls
// into appointment_booking (today's reservations for a staff member) and
// clinical_encounter (ListVisitsByPatient) — this repo never imports either.
type Patient struct {
	ID   string
	Name string
	Time string
}

type Visit struct {
	Date      string
	Reason    string
	Diagnosis string
}

type patientVisit struct {
	PatientID string
	Visit     Visit
}

// todayAgenda is the demo doctor's (see web/client.go's demoIdentity) fake
// schedule for today, seeded once at package init.
var todayAgenda = []Patient{
	{ID: "p1", Name: "Juan Pérez", Time: "09:00"},
	{ID: "p2", Name: "María Soto", Time: "09:30"},
	{ID: "p3", Name: "Diego Rojas", Time: "10:15"},
	{ID: "p4", Name: "Camila Vidal", Time: "11:00"},
}

// visits is a flat slice, scanned linearly, never a map — 4 patients and a
// handful of visits each is small enough that a map buys nothing, and this
// keeps the demo consistent with the ecosystem's no-Go-map-in-WASM-paths
// discipline (a map ships TinyGo's map runtime into the binary for no
// benefit at this size).
var visits = []patientVisit{
	{"p1", Visit{Date: "2026-07-20", Reason: "Control", Diagnosis: "Sin hallazgos"}},
	{"p1", Visit{Date: "2026-03-11", Reason: "Dolor abdominal", Diagnosis: "Gastritis"}},
	{"p2", Visit{Date: "2026-06-02", Reason: "Chequeo anual", Diagnosis: "Saludable"}},
	{"p3", Visit{Date: "2026-01-15", Reason: "Fractura", Diagnosis: "Fractura de radio"}},
}

func historyFor(patientID string) []Visit {
	var out []Visit
	for _, pv := range visits {
		if pv.PatientID == patientID {
			out = append(out, pv.Visit)
		}
	}
	return out
}

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

// ModelName satisfies layout.Module — RightPanel uses it for the wrapper's
// element id (rightpanel.go's panelID()).
func (v *View) ModelName() string { return "medicalhistory" }

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
		Module:       v,
		Title:        "Agenda de hoy",
		HeadControls: &v.picker,
		Article:      agenda,
		Aside:        history,
	}
	return panel.Render()
}
```

### 2c. `svg.go`

```go
//go:build !wasm

package medicalhistory

import "github.com/tinywasm/svg/sprite"

// IconSvg registers the module's glyph: a plain plus/cross built only from
// straight lines (no curves to get wrong) in a 24x24 box. tinywasm/ssr
// fuses every IconSvg() in the graph into one sprite injected into <body>.
//
// The receiver is instantiated as a zero value (&medicalhistory.Module{}),
// so this method must not touch any field — it only returns definitions.
func (m *Module) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		sprite.Define(Icon, "0 0 24 24", sprite.Path("M9 0H15V9H24V15H15V24H9V15H0V9H9V0Z")),
	)
}
```

**Acceptance for Stage 2**: `go build ./...` succeeds (backend target).
`GOOS=js GOARCH=wasm go build ./...` also succeeds (this package has no
`!wasm`-only imports outside `svg.go`, which is correctly tagged).

---

## 3. Register the module — `platformd/web/client.go`

This file is the demo app's composition root; `devices`/`about` are already
wired here. Add the new module the same way, changing only the `import`
block and the `p.Modules` slice — do not touch `demoBrand`/`demoIdentity`/
`hiddenModule`/`p := &platformd.Platform{...}` field values.

```go
import (
	"github.com/tinywasm/components/themetoggle"
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	_ "github.com/tinywasm/components/fieldset"
	"github.com/tinywasm/layout/platformd"
	"github.com/tinywasm/layout/platformd/modules/about"
	"github.com/tinywasm/layout/platformd/modules/devices"
	"github.com/tinywasm/layout/platformd/modules/medicalhistory"
	"github.com/tinywasm/svg"
)
```

```go
	p.Modules = []platformd.UIModule{
		devices.New(p),
		medicalhistory.New(),
		about.New(),
		hiddenModule{},
	}
```

Do **not** change `DefaultID` (stays `"devices"`) — this is an additive
registration, not a change of the demo's landing module. Navigate to
`#medicalhistory` manually to see it, exactly like the URL the user
referenced.

**Acceptance for Stage 3**: `grep -n "medicalhistory" platformd/web/client.go` shows the import and the `p.Modules` entry.

---

## 4. Documentation touch

`docs/ARCHITECTURE.md` line 7 currently reads:

```
│   └── modules/    # Demo modules, one package each (devices, home, about):
```

Update the parenthetical to include the new one:

```
│   └── modules/    # Demo modules, one package each (devices, medicalhistory, about):
```

(If the literal text differs slightly from the quote above by the time this
plan runs — e.g. the "home" module doesn't currently exist as a real
package — preserve whatever is actually on that line and only insert
`medicalhistory` into the list; do not invent or remove other entries.)

Do not create a new `docs/ARCHITECTURE.md` section, a per-module `README`,
or any other new doc file — `about`/`devices` have none either, and this
module is documented the same minimal way they are: doc comments in the
code itself (already written into Stage 2's file contents above).

**Acceptance for Stage 4**: `docs/ARCHITECTURE.md`'s modules line lists `medicalhistory`.

---

## 5. Final checklist

- [ ] `go.mod` requires `github.com/tinywasm/components v0.5.0` (or newer);
      `go.sum` updated via `go mod tidy`.
- [ ] `platformd/modules/medicalhistory/{medicalhistory,view,svg}.go`
      created exactly as specified in Stage 2.
- [ ] `platformd/web/client.go` imports and registers the module (Stage 3).
- [ ] `docs/ARCHITECTURE.md` modules line updated (Stage 4).
- [ ] `go build ./...` and `GOOS=js GOARCH=wasm go build ./...` both succeed.
- [ ] `gotest` all green (no new test files required by this plan — this is
      a visual demo; the behavior it exercises, `OnFilterChange`, is already
      covered by Stage A's tests in `components`).
- [ ] Manual browser check, dev server running (`tinywasm` hot reload — do
      **not** hand-compile, see `AGENTS.md`'s "Hot reload" section):
  - Navigate to `#medicalhistory`. The agenda (4 patients) renders on the
    main side, `SelectSearch` above it.
  - Clicking a row in the agenda list loads that patient's history in the
    aside.
  - Picking the same patient via the `SelectSearch` dropdown does the same
    thing (proves both entry points commit through `OnFilterChange`).
  - At a phone width: `SelectSearch`'s dropdown opens in-flow (not floating,
    not clipped) — confirms Stage A's CSS fix; the agenda/history panels
    behave as a `MasterDetail` scroll-snap strip, same as `devices`/`about`.
  - At desktop width: `SelectSearch`'s dropdown floats anchored under the
    toggle; agenda and history sit side by side.

## Stages

| Stage | File(s) | Depends on |
|---|---|---|
| 0 | — (dependency gate) | Stage A (`components` v0.5.0) published |
| 2 | `platformd/modules/medicalhistory/*.go` | 0 |
| 3 | `platformd/web/client.go` | 2 |
| 4 | `docs/ARCHITECTURE.md` | 2 |
| 5 | Final checklist + `gotest` + manual browser check | 3, 4 |
