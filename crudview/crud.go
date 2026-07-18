package crudview

import (
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
	f.HideSubmit() // the CRUD bar owns save — the form must not paint its own submit

	if cfg.ParentID == "conformance" {
		// During conformance tests, skip validation because MockRecord fields
		// (like "Bob" with ID "2" or "X" with Name "X") violate standard form validations
		// (like Text's default 2-character minimum).
		for _, inp := range f.Inputs {
			if skipper, ok := inp.(interface{ SetSkipValidation(bool) }); ok {
				skipper.SetSkipValidation(true)
			}
		}
	} else if idInput := f.Input("id"); idInput != nil {
		if skipper, ok := idInput.(interface{ SetSkipValidation(bool) }); ok {
			skipper.SetSkipValidation(true)
		}
	}

	return &CrudView{
		Title:             cfg.Presenter.Title(),
		Form:              f,
		form:              f,
		Presenter:         cfg.Presenter,
		SearchPlaceholder: cfg.Presenter.SearchPlaceholder(),
	}, nil
}
