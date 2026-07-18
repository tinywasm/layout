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
	"github.com/tinywasm/form/input"
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

// deviceModel is a model with real widgets (input.Text() is a model.Kind)
var deviceDef = model.Definition{
	Name: "device",
	Fields: model.Fields{
		{Name: "id", Type: input.Text()},
		{Name: "name", Type: input.Text(), NotNull: true},
		{Name: "ip", Type: input.Text()},
	},
}

type Device struct {
	Id, Name, Ip string
}

func (d *Device) ModelName() string     { return "device" }
func (d *Device) Schema() []model.Field { return deviceDef.Fields }
func (d *Device) Pointers() []any       { return []any{&d.Id, &d.Name, &d.Ip} }
func (d *Device) IsNil() bool           { return d == nil }

func (d *Device) EncodeFields(w model.FieldWriter) {
	w.String("id", d.Id)
	w.String("name", d.Name)
	w.String("ip", d.Ip)
}

func (d *Device) DecodeFields(r model.FieldReader) {
	d.Id, _ = r.String("id")
	d.Name, _ = r.String("name")
	d.Ip, _ = r.String("ip")
}

func (d *Device) Item() view.Item {
	return view.Item{ID: d.Id, Label: d.Name, Description: d.Ip}
}

var _ model.Model = (*Device)(nil) // Verify compile-time implementation
var _ view.Itemizer = (*Device)(nil)

type DeviceList struct {
	Items []*Device
}

func (l *DeviceList) IsNil() bool { return l == nil }
func (l *DeviceList) DecodeFields(r model.FieldReader) {}
func (l *DeviceList) Schema() []model.Field { return nil }
func (l *DeviceList) Pointers() []any { return nil }
func (l *DeviceList) Len() int { return len(l.Items) }
func (l *DeviceList) At(i int) model.Fielder { return l.Items[i] }
func (l *DeviceList) Append() model.Fielder {
	d := &Device{}
	l.Items = append(l.Items, d)
	return d
}

var _ model.ModelSlice = (*DeviceList)(nil)

type demoCaller struct{}

func (c *demoCaller) Call(op string, args model.Encodable, into model.Decodable, done func(err error)) {
	if op == "device_list" {
		dl, ok := into.(*DeviceList)
		if ok {
			// Pc Administracion
			d1 := dl.Append().(*Device)
			d1.Id = "1"
			d1.Name = "Pc Administracion"
			d1.Ip = "192.168.122.10"

			// Pc Ventas
			d2 := dl.Append().(*Device)
			d2.Id = "2"
			d2.Name = "Pc Ventas"
			d2.Ip = "192.168.122.11"

			// Servidor Web
			d3 := dl.Append().(*Device)
			d3.Id = "3"
			d3.Name = "Servidor Web"
			d3.Ip = "192.168.122.20"
		}
	}
	done(nil)
}

func (c *demoCaller) Dispatch(op string, args model.Encodable) {}

func (m mod) View() Component {
	if m.name == "crud" {
		pres := view.New(&demoCaller{}, &Device{}, "device_list",
			func() model.ModelSlice { return &DeviceList{} },
			view.WithTitle("Computadores"),
			view.WithSearchPlaceholder("Buscar..."),
			view.WithSaveOp("device_save"),
			view.WithDeleteOp("device_delete"),
		)
		cv, err := crudview.New(crudview.Config{
			ParentID:  "crud",
			Presenter: pres,
		})
		if err != nil {
			panic(err)
		}
		cv.OnSelect  = func(it view.Item) { m.p.Notify(Msg.Info, "Seleccionado: "+it.Label, 2000) }
		cv.OnNew     = func() { m.p.Notify(Msg.Info, "Nuevo", 2000) }
		cv.OnSaved   = func(err error) { if err == nil { m.p.Notify(Msg.Success, "Guardado", 2000) } }
		cv.OnDeleted = func(id string, err error) { if err == nil { m.p.Notify(Msg.Error, "Eliminado "+id, 2000) } }
		cv.OnCancel  = func() { m.p.Notify(Msg.Info, "Cancelado", 2000) }
		return cv
	}

	return &rightpanel.RightPanel{
		Module:  m,
		Title:   m.name,
		Article: Div().Text("Contenido de " + m.name),
	}
}

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
