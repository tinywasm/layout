//go:build !wasm

package rightpanel

import _ "embed"

//go:embed rightpanel.css
var css string

// RenderCSS implements dom.CSSProvider.
// tinywasm/site collects this during SSR to inject into <head>.
func (r *RightPanel) RenderCSS() string {
	return css
}
