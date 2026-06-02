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

// navIcon builds an SVG sprite reference with the pd-nav-icon class.
func navIcon(id string) *Element {
	return Svg(
		Use().Attr("href", "#"+id),
	).Class("pd-nav-icon").
		Attr("aria-hidden", "true").
		Attr("focusable", "false")
}

func main() {
	p := &platformd.Platform{
		AppName: "Demo Platform",
		Modules: []platformd.Module{
			{
				ID:      "mod1",
				Label:   "Módulo 1",
				Default: true,
				Icon:    navIcon("icon-home"),
				View: &rightpanel.RightPanel{
					Module: mod{"mod1"},
					Title:  "Módulo 1",
				},
			},
			{
				ID:    "mod2",
				Label: "Módulo 2",
				Icon:  navIcon("icon-products"),
				View: &rightpanel.RightPanel{
					Module: mod{"mod2"},
					Title:  "Módulo 2",
				},
			},
			{
				ID:    "mod3",
				Label: "Módulo 3",
				Icon:  navIcon("icon-info"),
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
