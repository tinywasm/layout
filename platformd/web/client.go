//go:build wasm

package main

import (
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	"github.com/tinywasm/layout/platformd"
	"github.com/tinywasm/layout/rightpanel"
	"github.com/tinywasm/svg"
)

// Tiny model stub so rightpanel has an ID source.
type mod struct {
	name string
	ico  svg.Icon
}

func (m mod) ModelName() string { return m.name }
func (m mod) Label() string     { return m.name }
func (m mod) Icon() svg.Icon    { return m.ico }
func (m mod) View() Component {
	return &rightpanel.RightPanel{
		Module: m,
		Title:  m.name,
	}
}

func main() {
	p := &platformd.Platform{
		AppName:   "Demo Platform",
		DefaultID: "mod1",
		Modules: []platformd.UIModule{
			mod{"mod1", platformd.IconHome},
			mod{"mod2", platformd.IconProducts},
			mod{"mod3", platformd.IconInfo},
		},
	}
	Append("body", p)

	p.Notify(Msg.Success, "Plataforma cargada", 3000)

	select {}
}
