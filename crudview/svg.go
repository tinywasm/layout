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
	// Icons: the toggle button (+ / ↺). These are FILL shapes (closed outlines),
	// never stroke lines: the sprite renders every path with fill="currentColor"
	// and no stroke, so a line-only path (e.g. "M8 1v14") has zero area and is
	// invisible. Solid FontAwesome-family glyphs match the platformd nav icons
	// and render identically.
	return sprite.NewSprite(
		// The single toggle button: plus (+) when nothing is selected (create).
		sprite.Define(iconCrudNew, "0 0 448 512", sprite.Path("M256 80c0-17.7-14.3-32-32-32s-32 14.3-32 32V224H48c-17.7 0-32 14.3-32 32s14.3 32 32 32H192V432c0 17.7 14.3 32 32 32s32-14.3 32-32V288H400c17.7 0 32-14.3 32-32s-14.3-32-32-32H256V80z")),
		// The single toggle button: undo-arrow (↺) when a row is selected — undo
		// deselects and clears the form, back to the "+" state.
		sprite.Define(iconCrudCancel, "0 0 512 512", sprite.Path("M125.7 160H176c17.7 0 32 14.3 32 32s-14.3 32-32 32H48c-17.7 0-32-14.3-32-32V64c0-17.7 14.3-32 32-32s32 14.3 32 32v51.2L97.6 97.6c87.5-87.5 229.3-87.5 316.8 0s87.5 229.3 0 316.8s-229.3 87.5-316.8 0c-12.5-12.5-12.5-32.8 0-45.3s32.8-12.5 45.3 0c62.5 62.5 163.8 62.5 226.3 0s62.5-163.8 0-226.3s-163.8-62.5-226.3 0L125.7 160z")),
		// Bulk delete button icon (🗑).
		sprite.Define(iconCrudDelete, "0 0 448 512", sprite.Path("M135.2 17.7C140.6 6.8 151.7 0 163.8 0H284.2c12.1 0 23.2 6.8 28.6 17.7L320 32h96c17.7 0 32 14.3 32 32s-14.3 32-32 32H32C14.3 96 0 81.7 0 64S14.3 32 32 32h96l7.2-14.3zM32 128H416V448c0 35.3-28.7 64-64 64H96c-35.3 0-64-28.7-64-64V128zm96 64c-8.8 0-16 7.2-16 16V400c0 8.8 7.2 16 16 16s16-7.2 16-16V208c0-8.8-7.2-16-16-16zm96 0c-8.8 0-16 7.2-16 16V400c0 8.8 7.2 16 16 16s16-7.2 16-16V208c0-8.8-7.2-16-16-16zm96 0c-8.8 0-16 7.2-16 16V400c0 8.8 7.2 16 16 16s16-7.2 16-16V208c0-8.8-7.2-16-16-16z")),
		// Bulk edit button icon (✏).
		sprite.Define(iconCrudEdit, "0 0 512 512", sprite.Path("M497.9 142.1l-46.1 46.1c-4.7 4.7-12.3 4.7-17 0l-111-111c-4.7-4.7-4.7-12.3 0-17l46.1-46.1c18.7-18.7 49.1-18.7 67.9 0l60.1 60.1c18.8 18.7 18.8 49.1 0 67.9zM284.2 99.8L21.6 362.4.4 483.9c-2.9 16.4 11.4 30.6 27.8 27.8l121.5-21.3 262.6-262.6c4.7-4.7 4.7-12.3 0-17l-111-111c-4.8-4.7-12.4-4.7-17.1 0zM124.1 339.9c-5.5-5.5-5.5-14.3 0-19.8l154-154c5.5-5.5 14.3-5.5 19.8 0s5.5 14.3 0 19.8l-154 154c-5.5 5.5-14.3 5.5-19.8 0zM88 424h48v36.3l-64.5 11.3L88 424z")),
	)
}
