// Package home es el módulo demo más simple: un rightpanel con un artículo y nada
// más. Es el patrón a copiar para una pantalla de contenido estático.
package home

import (
	"github.com/tinywasm/layout/platformd"
	"github.com/tinywasm/layout/rightpanel"
	"github.com/tinywasm/svg"

	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/html"
)

const Icon = svg.Icon("mod-home")

type Module struct{}

func New() *Module { return &Module{} }

var _ platformd.UIModule = (*Module)(nil)

func (m *Module) ModelName() string { return "home" }
func (m *Module) Label() string     { return "Inicio" }
func (m *Module) Icon() svg.Icon    { return Icon }

// Sin Module: dentro de platformd la SECCIÓN del escenario es la dueña del id de
// ruta, y RightPanel estampa Module.ModelName() en su propia raíz — dos nodos con
// un mismo id. Title lleva la etiqueta, que es todo lo que este panel necesita.
func (m *Module) View() Component {
	return &rightpanel.RightPanel{
		Title:   m.Label(),
		Article: Div().Text("Contenido de " + m.Label()),
	}
}