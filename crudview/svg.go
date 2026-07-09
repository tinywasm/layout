//go:build !wasm

package crudview

import (
	"github.com/tinywasm/svg"
)

// IconSvg returns the sprite with the captured Pa100T icons.
// Captured paths from Pa100T v3.0.1 (base64 SVGs in IE_MODULE_CONTENT.md §2).
func IconSvg() *svg.Sprite {
	return svg.NewSprite(
		// btn_crudnew: plus (+)
		svg.Define("icon-crud-new", "0 0 16 16", svg.Path("M8 1v14M1 8h14")),
		// btn_cruddel: minus (−)
		svg.Define("icon-crud-del", "0 0 16 16", svg.Path("M1 8h14")),
		// btn_crudcancel: undo-arrow (↺)
		svg.Define("icon-crud-cancel", "0 0 16 16", svg.Path("M4 4h-4v-4M0 4c2-4 8-4 10-2s4 8 2 10-8 4-10 2")),
		// btn_crudsave: save-disk (💾)
		svg.Define("icon-crud-save", "0 0 16 16", svg.Path("M2 1h10l3 3v11h-13zM10 1v4h-7v-4M4 15v-5h8v5")),
		// aside-search: magnifier (lupa)
		svg.Define("icon-crud-search", "0 0 16 16", svg.Path("M11 11l4 4M1 7a6 6 0 1 0 12 0a6 6 0 1 0-12 0")),
	)
}
