//go:build wasm

package main

import (
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	"github.com/tinywasm/layout/platformd"
	"github.com/tinywasm/layout/rightpanel"
)

// Tiny model stub so rightpanel has an ID source.
type mod struct{ name string }

func (m mod) ModelName() string { return m.name }

func main() {
	p := &platformd.Platform{
		AppName: "Demo Platform",
		Modules: []platformd.Module{
			{
				ID:    "mod1",
				Label: "Módulo 1",
				Default: true,
				Icon:  Div().Class("icon-home"), // dummy icon
				View: &rightpanel.RightPanel{
					Module: mod{"mod1"},
					Title:  "Módulo 1",
				},
			},
			{
				ID:    "mod2",
				Label: "Módulo 2",
				Icon:  Div().Class("icon-products"), // dummy icon
				View: &rightpanel.RightPanel{
					Module: mod{"mod2"},
					Title:  "Módulo 2",
				},
			},
			{
				ID:    "mod3",
				Label: "Módulo 3",
				Icon:  Div().Class("icon-info"), // dummy icon
				View: &rightpanel.RightPanel{
					Module: mod{"mod3"},
					Title:  "Módulo 3",
				},
			},
		},
	}
	Append("body", p)

	// demo of typed Notify — uses fmt.MessageType / fmt.Msg
	p.Notify(Msg.Success, "Plataforma cargada", 3000)
}
