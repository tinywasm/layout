package crudview

import (
	"github.com/tinywasm/components/searchbar"
	"github.com/tinywasm/dom"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/form"
	"github.com/tinywasm/view"
)

// Config is what a renderer needs to draw a view.Presenter — nothing about ops, transport, or
// codec: that is ALL inside the Presenter already. The module builds the Presenter via
// view.New(...) (importing view+model+router, never layout) and hands it here.
type Config struct {
	// ParentID is the DOM id the form mounts under.
	ParentID string
	// Presenter is built by the module via view.New(...). Required.
	Presenter view.Presenter

	// Filter is the control that narrows the list. Optional: nil renders no
	// controls band. When nil, New installs a searchbar.SearchBar carrying the
	// presenter's placeholder — the ergonomic default, not a decision imposed:
	// pass any widget.Filterable to replace it.
	Filter dom.Component
}

// New builds the renderer around an already-constructed Presenter. It generates the form from
// Presenter.Record().Schema(), and wires save/delete to sync the form INTO Record() before
// calling Presenter.Save/Delete — crudview never talks to a Caller directly anymore.
func New(cfg Config) (*CrudView, error) {
	if cfg.Presenter == nil {
		return nil, fmt.Errf("crudview.New: Presenter is required")
	}

	f, err := form.New(cfg.ParentID, cfg.Presenter.Record())
	if err != nil {
		return nil, err // a record with no widgets fails HERE, loudly
	}
	f.HideSubmit() // there is no Save button — auto-save (OnFieldChange) replaces it

	// The default filter keeps every existing consumer compiling: a host that
	// does not care which control it gets still gets a working search bar with
	// the presenter's placeholder. The demo injects its own — see platformd.
	filter := cfg.Filter
	if filter == nil {
		filter = &searchbar.SearchBar{Placeholder: cfg.Presenter.SearchPlaceholder()}
	}

	v := &CrudView{
		Title:     cfg.Presenter.Title(),
		Form:      f,
		form:      f,
		Presenter: cfg.Presenter,
		Filter:    filter,
	}

	// Auto-save: every field commit (blur/change) persists immediately — see
	// docs/ROADMAP.md "Save: auto-save". Only wired when the presenter can save;
	// autoSaveAction is a no-op otherwise.
	f.OnFieldChange(func() { v.autoSaveAction() })

	return v, nil
}
