//go:build !wasm

package crudview

import (
	"github.com/tinywasm/icons/pencil"
	"github.com/tinywasm/icons/plus"
	"github.com/tinywasm/icons/trash"
	"github.com/tinywasm/icons/undo"
	"github.com/tinywasm/svg/sprite"
)

// IconSvg ships the four crud glyphs the footer buttons render. All come from
// tinywasm/icons: plus/undo for the single toggle button, trash/pencil for the
// bulk delete/edit buttons — the same trash/pencil the row marks
// (targetlist/targetdate) draw from their own IconSvg(), so assetmin collapses
// each to one symbol and the button and the rows it acts on never drift.
//
// Method receiver (not a free function) so tinywasm/ssr detects a single
// receiver type for the package and emits RenderCSS + IconSvg together.
func (v *CrudView) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		plus.Def(),
		undo.Def(),
		trash.Def(),
		pencil.Def(),
	)
}
