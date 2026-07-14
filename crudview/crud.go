package crudview

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/form"
	"github.com/tinywasm/model"
	"github.com/tinywasm/router"
)

// Config is everything a module decides about its CRUD view: its record and its ops.
// Everything else — generating the form from the record, filling it on select, validating
// and shipping it on save, reloading the list — is wired HERE, once, for every app.
type Config struct {
	// ParentID is the DOM id the form mounts under (a module passes its own ID).
	ParentID string

	Caller router.Caller
	Title  string // h1, top-left

	// Record is BOTH the form's source of truth and the payload sent over the wire.
	// model.Model (v0.0.14) is exactly that contract — never declare a local intersection.
	Record model.Model

	ListOp   string // required
	SaveOp   string // "" → no save button
	DeleteOp string // "" → no delete button

	Args   func() model.Encodable           // list args (nil → none)
	Decode func(raw []byte) ([]Item, error) // list response → cards

	// Fill returns the full record for id, from what the module ALREADY decoded in Decode.
	// A nil return (including a typed-nil pointer) resets the form — that is not an error.
	Fill func(id string) model.Model

	OnError func(err error)

	// SearchPlaceholder overrides the search box text ("" → "Search…").
	SearchPlaceholder string
}

// New builds the standard CRUD view and wires the whole form↔list↔transport cycle.
// A module calls this and passes its record and its ops. Nothing else.
func New(cfg Config) (*CrudView, error) {
	if cfg.Caller == nil {
		return nil, fmt.Errf("crudview.New: Caller is required")
	}
	if cfg.ListOp == "" {
		return nil, fmt.Errf("crudview.New: ListOp is required")
	}
	if model.IsNil(cfg.Record) {
		return nil, fmt.Errf("crudview.New: Record is required")
	}

	f, err := form.New(cfg.ParentID, cfg.Record)
	if err != nil {
		return nil, err // a record with no widgets fails HERE, loudly
	}
	f.HideSubmit() // the CRUD bar owns save — the form must not paint its own submit

	v := &CrudView{
		Title:             cfg.Title,
		Form:              f,
		Source:            Source{Caller: cfg.Caller, ListOp: cfg.ListOp, Args: cfg.Args, Decode: cfg.Decode},
		OnError:           cfg.OnError,
		SearchPlaceholder: cfg.SearchPlaceholder,
	}

	v.OnSelect = func(it Item) {
		if cfg.Fill == nil {
			return
		}
		_ = f.LoadValues(cfg.Fill(it.ID)) // nil record → LoadValues resets. Not an error.
	}
	v.OnNew = func() { f.Reset() }
	v.OnCancel = func() { f.Reset() }

	if cfg.SaveOp != "" {
		v.OnSave = func(done func(err error)) {
			if err := f.Validate(); err != nil {
				done(err)
				return
			}
			if err := f.SyncValues(cfg.Record); err != nil {
				done(err)
				return
			}
			cfg.Caller.Call(cfg.SaveOp, cfg.Record, func(_ []byte, err error) { done(err) })
		}
	}

	if cfg.DeleteOp != "" {
		v.OnDelete = func(id string, done func(err error)) {
			rec := cfg.Fill(id)
			if model.IsNil(rec) {
				done(fmt.Errf("crudview: no record for id %s", id))
				return
			}
			cfg.Caller.Call(cfg.DeleteOp, rec, func(_ []byte, err error) { done(err) })
		}
	}

	return v, nil
}
