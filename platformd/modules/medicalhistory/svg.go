//go:build !wasm

package medicalhistory

import "github.com/tinywasm/svg/sprite"

// IconSvg registers the module's glyph: a plain plus/cross built only from
// straight lines (no curves to get wrong) in a 24x24 box. tinywasm/ssr
// fuses every IconSvg() in the graph into one sprite injected into <body>.
//
// The receiver is instantiated as a zero value (&medicalhistory.Module{}),
// so this method must not touch any field — it only returns definitions.
func (m *Module) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		sprite.Define(Icon, "0 0 24 24", sprite.Path("M9 0H15V9H24V15H15V24H9V15H0V9H9V0Z")),
	)
}
