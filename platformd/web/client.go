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
	"github.com/tinywasm/svg"
	"github.com/tinywasm/view"
)

// Tiny model stub so layouts have an ID source.
type mod struct {
	name string
	icon svg.Icon
	p    *platformd.Platform
}

func (m mod) ModelName() string { return m.name }
func (m mod) Label() string     { return m.name }
func (m mod) Icon() svg.Icon    { return m.icon }

type mockModel struct{}
func (m *mockModel) ModelName() string { return "device" }
func (m *mockModel) IsNil() bool { return m == nil }
func (m *mockModel) Schema() []model.Field {
	return []model.Field{
		{Name: "id", Type: model.Text()},
	}
}
func (m *mockModel) Pointers() []any { return []any{new(string)} }
func (m *mockModel) EncodeFields(w model.FieldWriter) {}
func (m *mockModel) DecodeFields(r model.FieldReader) {}

type mockPresenter struct {
	items []view.Item
}

func (p *mockPresenter) Title() string             { return "Computadores" }
func (p *mockPresenter) SearchPlaceholder() string { return "Buscar..." }
func (p *mockPresenter) Record() model.Model       { return &mockModel{} }
func (p *mockPresenter) Items() []view.Item        { return p.items }
func (p *mockPresenter) Reload() error             { return nil }
func (p *mockPresenter) Selected() string          { return "" }
func (p *mockPresenter) Select(id string) model.Model { return &mockModel{} }
func (p *mockPresenter) CanSave() bool             { return true }
func (p *mockPresenter) Save(payload model.Model) error { return nil }
func (p *mockPresenter) CanDelete() bool           { return true }
func (p *mockPresenter) Delete(id string) error    { return nil }

func (m mod) View() Component {
	if m.name == "crud" {
		pres := &mockPresenter{
			items: []view.Item{
				{ID: "1", Label: "Pc Administracion", Description: "192.168.122.10"},
				{ID: "2", Label: "Pc Ventas", Description: "192.168.122.11"},
				{ID: "3", Label: "Servidor Web", Description: "192.168.122.20"},
			},
		}
		cv, err := crudview.New(crudview.Config{
			ParentID:  "crud",
			Presenter: pres,
		})
		if err != nil {
			panic(err)
		}
		cv.OnSelect = func(it view.Item) {
			m.p.Notify(Msg.Info, "Seleccionado: "+it.Label, 2000)
		}
		cv.OnNew = func() { m.p.Notify(Msg.Info, "Nuevo", 2000) }
		cv.OnSave = func(done func(err error)) { m.p.Notify(Msg.Success, "Guardado", 2000); done(nil) }
		cv.OnDelete = func(id string, done func(err error)) { m.p.Notify(Msg.Error, "Eliminado "+id, 2000); done(nil) }
		cv.OnCancel = func() { m.p.Notify(Msg.Info, "Cancelado", 2000) }
		return cv
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
		mod{"crud", platformd.IconProducts, p},
		mod{"mod1", platformd.IconHome, p},
		mod{"mod2", platformd.IconInfo, p},
		mod{"hidden", platformd.IconInfo, p},
	}

	Append("body", p)

	p.Notify(Msg.Success, "Plataforma cargada", 3000)

	select {}
}
