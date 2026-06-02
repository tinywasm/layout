//go:build wasm

package main

import (
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	. "github.com/tinywasm/svg"
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
				ID:      "mod1",
				Label:   "Módulo 1",
				Default: true,
				Icon:    Icon("icon-home", "pd-nav-icon"),
				View: &rightpanel.RightPanel{
					Module: mod{"mod1"},
					Title:  "Módulo 1",
				},
			},
			{
				ID:    "mod2",
				Label: "Módulo 2",
				Icon:  Icon("icon-products", "pd-nav-icon"),
				View: &rightpanel.RightPanel{
					Module: mod{"mod2"},
					Title:  "Módulo 2",
				},
			},
			{
				ID:    "mod3",
				Label: "Módulo 3",
				Icon:  Icon("icon-info", "pd-nav-icon"),
				View: &rightpanel.RightPanel{
					Module: mod{"mod3"},
					Title:  "Módulo 3",
				},
			},
		},
	}
	Append("body", p)

	p.Notify(Msg.Success, "Plataforma cargada", 3000)

	select {}
}
