//go:build wasm

package main

import (
	"github.com/tinywasm/components/searchbar"
	"github.com/tinywasm/components/themetoggle"
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	. "github.com/tinywasm/html"
	// Global form skin: one import at the composition root makes EVERY form in the
	// app render as labeled fieldset boxes (CSS-only, collected via SSR).
	_ "github.com/tinywasm/components/fieldset"
	"github.com/tinywasm/form/input"
	"github.com/tinywasm/layout/crudview"
	"github.com/tinywasm/layout/platformd"
	"github.com/tinywasm/layout/rightpanel"
	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/storage"
	"github.com/tinywasm/storage/mem"
	"github.com/tinywasm/svg"
	"github.com/tinywasm/view"
)

// Tiny model stub so layouts have an ID source. name is the route; label is
// what the user reads — the rail and the module's own title have to say the
// same thing, or the nav claims you are somewhere you are not.
type mod struct {
	name  string
	label string
	icon  svg.Icon
	p     *platformd.Platform
}

func (m mod) ModelName() string { return m.name }
func (m mod) Label() string     { return m.label }
func (m mod) Icon() svg.Icon    { return m.icon }

// deviceModel is a model built from real widgets (input.Text() is a model.Kind).
// All three fields are NotNull (a device with a blank id/name/ip isn't a
// valid record) and ip uses the dedicated input.IP() widget — not
// input.Text() — so a malformed address fails Form.Validate() instead of
// silently persisting.
var deviceDef = model.Definition{
	Name: "device",
	Fields: model.Fields{
		{Name: "id", Type: input.Text(), NotNull: true},
		{Name: "name", Type: input.Text(), NotNull: true},
		{Name: "ip", Type: input.IP(), NotNull: true},
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

func (l *DeviceList) IsNil() bool                      { return l == nil }
func (l *DeviceList) DecodeFields(r model.FieldReader) {}
func (l *DeviceList) Schema() []model.Field            { return nil }
func (l *DeviceList) Pointers() []any                  { return nil }
func (l *DeviceList) Len() int                         { return len(l.Items) }
func (l *DeviceList) At(i int) model.Fielder           { return l.Items[i] }
func (l *DeviceList) Append() model.Fielder {
	d := &Device{}
	l.Items = append(l.Items, d)
	return d
}

var _ model.ModelSlice = (*DeviceList)(nil)

// deviceDB is a real (in-memory) CRUD backend for the demo, so the UI's actual
// behavior — save-on-blur, the new item landing in the list, delete removing
// it — can be exercised for real instead of against a static 3-item fixture.
// Package-level and built once: it must survive across View() calls (switching
// tabs and back) so the demo doesn't lose its data every time the module is
// re-rendered. It DOES reset on a full page reload/process restart — that is
// the expected trade-off of an in-memory store, not a bug.
var deviceDB = newSeededDeviceDB()

func newSeededDeviceDB() *orm.DB {
	db := orm.New(mem.New())
	for _, d := range []*Device{
		{Id: "10", Name: "Pc Administracion", Ip: "192.168.122.10"},
		{Id: "11", Name: "Pc Ventas", Ip: "192.168.122.11"},
		{Id: "12", Name: "Servidor Web", Ip: "192.168.122.20"},
	} {
		_ = db.Create(d)
	}
	return db
}

// memCaller adapts deviceDB (an *orm.DB over storage/mem) to router.Caller —
// the seam view.Presenter drives. Device-specific (type-asserts *Device)
// rather than generic: this is app/demo wiring, not a shared library.
type memCaller struct{ db *orm.DB }

func (c *memCaller) Call(op string, args model.Encodable, into model.Decodable, done func(err error)) {
	var err error
	switch op {
	case "device_list":
		if dl, ok := into.(*DeviceList); ok {
			var rows []*Device
			err = c.db.Query(&Device{}).ReadAll(
				func() model.Model { return &Device{} },
				func(m model.Model) { rows = append(rows, m.(*Device)) },
			)
			// Newest first: mem has no timestamp/sequence column, so this just
			// reverses creation order — the same order Create() appended in.
			for i := len(rows) - 1; i >= 0 && err == nil; i-- {
				dst := dl.Append().(*Device)
				*dst = *rows[i]
			}
		}
	case "device_save":
		d, ok := args.(*Device)
		if !ok {
			err = Errf("memCaller: device_save: unexpected payload type")
			break
		}
		if findErr := c.db.Query(&Device{}).Where("id").Eq(d.Id).ReadOne(); findErr != nil {
			err = c.db.Create(d) // no existing row with this id — new record
		} else {
			err = c.db.Update(d, storage.Eq("id", d.Id))
		}
	case "device_delete":
		d, ok := args.(*Device)
		if !ok {
			err = Errf("memCaller: device_delete: unexpected payload type")
			break
		}
		err = c.db.Delete(d, storage.Eq("id", d.Id))
	}
	done(err)
}

func (c *memCaller) Dispatch(op string, args model.Encodable) {}

func (m mod) View() Component {
	if m.name == "crud" {
		pres := view.New(&memCaller{db: deviceDB}, &Device{}, "device_list",
			func() model.ModelSlice { return &DeviceList{} },
			view.WithTitle("Computadores"),
			view.WithSaveOp("device_save"),
			view.WithDeleteOp("device_delete"),
		)
		cv, err := crudview.New(crudview.Config{
			ParentID:  "crud",
			Presenter: pres,
			// The filter is the application's choice, not the layout's. This is
			// the line a deployment swaps for a calendar or a category select;
			// nothing in crudview or rightpanel changes when it does.
			Filter: &searchbar.SearchBar{Placeholder: "Buscar..."},
		})
		if err != nil {
			panic(err)
		}
		// No OnSelect notification: selecting a row is navigation, not an
		// action — a toast there adds no value and, on mobile (where the
		// module fills the screen), the message overlays the whole view. Save/
		// delete below DO notify, since those confirm a real mutation.
		cv.OnNew = func() { m.p.Notify(Msg.Info, "Nuevo", 2000) }
		cv.OnSaved = func(err error) {
			if err == nil {
				m.p.Notify(Msg.Success, "Guardado", 2000)
			}
		}
		cv.OnDeleted = func(id string, err error) {
			if err == nil {
				m.p.Notify(Msg.Error, "Eliminado "+id, 2000)
			}
		}
		cv.OnCancel = func() { m.p.Notify(Msg.Info, "Cancelado", 2000) }
		return cv
	}

	// No Module: inside platformd the SECTION of the deck owns the route id, and
	// RightPanel stamps Module.ModelName() on its own root — two nodes, one id,
	// which is what Get(moduleID) resolves for the scroll. Title carries the
	// label, which is all this panel needs from the module.
	return &rightpanel.RightPanel{
		Title:   m.label,
		Article: Div().Text("Contenido de " + m.label),
	}
}

// demoBrand mocks the platform's own identity — what the shell is called and
// marked with. In a real application this comes from the deployment, not from
// a login. The mark must be NON-empty: demoIdentity's empty avatar already
// proves the fallback path, and this stub exists to prove the non-empty one —
// the two render different shapes in the header.
type demoBrand struct{}

func (demoBrand) BrandName() string { return "Demo Platform" }
func (demoBrand) BrandMark() string {
	// An inline SVG data URI, so the demo needs no asset server. The
	// percent-encoding of # is load-bearing: a raw # would read as a URL
	// fragment and the image would 404.
	return "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 512 512'%3E%3Cpath d='M272.6 17.9L323.8 158.5l143.7 12.3c27.5 2.4 38.6 37.6 17.4 55.2l-109.5 96.5 32 142.6c6.1 27.2-23.7 48.5-47.9 34.2L256 403.9 152.9 499.3c-24.2 14.3-54-7-47.9-34.2l32-142.6-109.5-96.5C7.9 208.4 19 173.2 46.5 170.8l143.7-12.3L241.4 17.9c9.5-23.9 42.6-23.9 52.2 0z' fill='%23ffffff'/%3E%3C/svg%3E"
}

// demoIdentity mocks the login for the demo. In a real application this is
// whatever github.com/tinywasm/user hands back for the current session — see
// that repository's docs/PLAN.md for the work that makes it satisfy this
// contract directly.
type demoIdentity struct{}

func (demoIdentity) UserName() string   { return "Thor Odinson" }
func (demoIdentity) UserAvatar() string { return "" } // sin imagen: cae al glifo
func (demoIdentity) UserRoles() []string {
	// Tres, para que el caso N se vea en el demo.
	return []string{"Administrador", "Operaciones", "Soporte"}
}

func main() {
	p := &platformd.Platform{
		AppName:   "Demo Platform",
		Brand:     demoBrand{},
		DefaultID: "crud",
		User:      demoIdentity{},
		// A factory, not an instance: the shell renders one menu for the header
		// and one for the drawer, and a single component in both places would
		// put one id on two nodes — the second copy inert.
		UserActions: func() Component {
			return &themetoggle.ThemeToggle{}
		},
		CanView: func(id string) bool {
			return id != "hidden"
		},
	}

	p.Modules = []platformd.UIModule{
		mod{"crud", "Computadores", platformd.IconProducts, p},
		mod{"mod1", "Inicio", platformd.IconHome, p},
		mod{"mod2", "Acerca de", platformd.IconInfo, p},
		mod{"hidden", "Oculto", platformd.IconInfo, p},
	}

	Append("body", p)

	p.Notify(Msg.Success, "Plataforma cargada", 3000)

	select {}
}
