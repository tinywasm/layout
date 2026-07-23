//go:build !wasm

package crudview

import (
	"github.com/tinywasm/svg/sprite"
)

// IconSvg returns the sprite with the captured Pa100T icons.
// Captured paths from Pa100T v3.0.1 (base64 SVGs in IE_MODULE_CONTENT.md §2).
// Method receiver (not a free function) so tinywasm/ssr detects a single
// receiver type for the package and emits RenderCSS + IconSvg together.
func (v *CrudView) IconSvg() *sprite.Sprite {
	// Icons: the toggle button (+ / ↺) and the search magnifier. These are FILL
	// shapes (closed outlines), never stroke lines: the sprite renders every path
	// with fill="currentColor" and no stroke, so a line-only path (e.g. "M8 1v14")
	// has zero area and is invisible. Solid FontAwesome-family glyphs match the
	// platformd nav icons and render identically.
	return sprite.NewSprite(
		// The single toggle button: plus (+) when nothing is selected (create).
		sprite.Define(iconCrudNew, "0 0 448 512", sprite.Path("M256 80c0-17.7-14.3-32-32-32s-32 14.3-32 32V224H48c-17.7 0-32 14.3-32 32s14.3 32 32 32H192V432c0 17.7 14.3 32 32 32s32-14.3 32-32V288H400c17.7 0 32-14.3 32-32s-14.3-32-32-32H256V80z")),
		// The single toggle button: undo-arrow (↺) when a row is selected — undo
		// deselects and clears the form, back to the "+" state.
		sprite.Define(iconCrudCancel, "0 0 512 512", sprite.Path("M125.7 160H176c17.7 0 32 14.3 32 32s-14.3 32-32 32H48c-17.7 0-32-14.3-32-32V64c0-17.7 14.3-32 32-32s32 14.3 32 32v51.2L97.6 97.6c87.5-87.5 229.3-87.5 316.8 0s87.5 229.3 0 316.8s-229.3 87.5-316.8 0c-12.5-12.5-12.5-32.8 0-45.3s32.8-12.5 45.3 0c62.5 62.5 163.8 62.5 226.3 0s62.5-163.8 0-226.3s-163.8-62.5-226.3 0L125.7 160z")),
		// aside-search: magnifier (lupa). The lens hole renders via the inner
		// circle subpath winding opposite the outer — same as the platformd "info".
		sprite.Define(iconCrudSearch, "0 0 512 512", sprite.Path("M416 208c0 45.9-14.9 88.3-40 122.7L502.6 457.4c12.5 12.5 12.5 32.8 0 45.3s-32.8 12.5-45.3 0L330.7 376c-34.4 25.2-76.8 40-122.7 40C93.1 416 0 322.9 0 208S93.1 0 208 0S416 93.1 416 208zM208 352a144 144 0 1 0 0-288 144 144 0 1 0 0 288z")),
		// Mobile "‹ back" button: chevron-left, returns to the list panel
		// without clearing the selection (see clsBackBtn's click handler).
		sprite.Define(iconCrudBack, "0 0 320 512", sprite.Path("M9.4 233.4c-12.5 12.5-12.5 32.8 0 45.3l192 192c12.5 12.5 32.8 12.5 45.3 0s12.5-32.8 0-45.3L109.2 256 246.6 118.6c12.5-12.5 12.5-32.8 0-45.3s-32.8-12.5-45.3 0l-192 192z")),
	)
}
