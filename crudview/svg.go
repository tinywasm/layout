//go:build !wasm

package crudview

import (
	"github.com/tinywasm/svg/sprite"
)

// IconSvg returns the sprite with the captured Pa100T icons.
// Captured paths from Pa100T v3.0.1 (base64 SVGs in IE_MODULE_CONTENT.md §2).
func IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		// btn_crudnew: plus (+)
		sprite.Define(iconCrudNew, "0 0 16 16", sprite.Path("M8 1v14M1 8h14")),
		// btn_cruddel: minus (−)
		sprite.Define(iconCrudDel, "0 0 16 16", sprite.Path("M1 8h14")),
		// btn_crudcancel: undo-arrow (↺)
		sprite.Define(iconCrudCancel, "0 0 16 16", sprite.Path("M4 4h-4v-4M0 4c2-4 8-4 10-2s4 8 2 10-8 4-10 2")),
		// btn_crudsave: save-disk (💾)
		sprite.Define(iconCrudSave, "0 0 16 16", sprite.Path("M2 1h10l3 3v11h-13zM10 1v4h-7v-4M4 15v-5h8v5")),
		// aside-search: magnifier (lupa)
		sprite.Define(iconCrudSearch, "0 0 16 16", sprite.Path("M11 11l4 4M1 7a6 6 0 1 0 12 0a6 6 0 1 0-12 0")),
	)
}
