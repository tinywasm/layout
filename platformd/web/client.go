//go:build wasm

package main

import (
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	. "github.com/tinywasm/html"
	"github.com/tinywasm/layout/crudview"
	"github.com/tinywasm/layout/platformd"
	"github.com/tinywasm/layout/rightpanel"
	"github.com/tinywasm/model"
)

// Tiny model stub so layouts have an ID source.
type mod struct {
	name string
	icon string
	p    *platformd.Platform
}

func (m mod) ModelName() string { return m.name }
func (m mod) Label() string     { return m.name }
func (m mod) IconID() string    { return m.icon }
func (m mod) View() Component {
	if m.name == "crud" {
		return &crudview.CrudView{
			Title: "Computadores",
			Form:  Div().Text("Formulario de edición"),
			Source: crudview.Source{
				Caller: &mockCaller{},
				ListOp: "list",
				Decode: func(raw []byte) ([]crudview.Item, error) {
					// Minimal hand-rolled decoder since we can't use encoding/json
					// and we don't have model generation for Item here.
					return []crudview.Item{
						{ID: "1", Label: "Pc Administracion", Description: "192.168.122.10"},
						{ID: "2", Label: "Pc Ventas", Description: "192.168.122.11"},
						{ID: "3", Label: "Servidor Web", Description: "192.168.122.20"},
					}, nil
				},
			},
			OnSelect: func(it crudview.Item) {
				m.p.Notify(Msg.Info, "Seleccionado: "+it.Label, 2000)
			},
			OnNew:    func() { m.p.Notify(Msg.Info, "Nuevo", 2000) },
			OnSave:   func(done func(err error)) { m.p.Notify(Msg.Success, "Guardado", 2000); done(nil) },
			OnDelete: func(id string, done func(err error)) { m.p.Notify(Msg.Error, "Eliminado "+id, 2000); done(nil) },
			OnCancel: func() { m.p.Notify(Msg.Info, "Cancelado", 2000) },
		}
	}

	return &rightpanel.RightPanel{
		Module:  m,
		Title:   m.name,
		Article: Div().Text("Contenido de " + m.name),
	}
}

type mockCaller struct{}

func (c *mockCaller) Call(op string, args model.Encodable, callback func(result []byte, err error)) {
	callback(nil, nil)
}

func (c *mockCaller) Dispatch(op string, args model.Encodable) {}

func main() {
	p := &platformd.Platform{
		AppName:   "Demo Platform",
		DefaultID: "crud",
		CanView: func(id string) bool {
			return id != "hidden"
		},
	}

	p.Modules = []platformd.UIModule{
		mod{"crud", "icon-products", p},
		mod{"mod1", "icon-home", p},
		mod{"mod2", "icon-info", p},
		mod{"hidden", "icon-info", p},
	}

	Append("body", p)

	p.Notify(Msg.Success, "Plataforma cargada", 3000)

	select {}
}
