//go:build !wasm

package platformd

import . "github.com/tinywasm/css"

var (
	tokenFontSizeNormal = Token{Name: "--pd-font-size-normal", Fallback: "1.1rem"}
	tokenFontSizeSmall  = Token{Name: "--pd-font-size-small", Fallback: ".6rem"}

	tokenColorPrimary    = Token{Name: "--pd-color-primary", Fallback: "#ffffff"}
	tokenColorSecondary  = Token{Name: "--pd-color-secondary", Fallback: "#3f88bf"}
	tokenColorTertiary   = Token{Name: "--pd-color-tertiary", Fallback: "#c2c1c1"}
	tokenColorQuaternary = Token{Name: "--pd-color-quaternary", Fallback: "#000000"}
	tokenColorGray       = Token{Name: "--pd-color-gray", Fallback: "#e9e9e9"}
	tokenColorSelection  = Token{Name: "--pd-color-selection", Fallback: "#ff9300"}
	tokenColorHover      = Token{Name: "--pd-color-hover", Fallback: "#ff95008e"}
	tokenColorSuccess    = Token{Name: "--pd-color-success", Fallback: "#aadaff7c"}
	tokenColorError      = Token{Name: "--pd-color-error", Fallback: "#f20707"}

	tokenMenuSize      = Token{Name: "--pd-menu-size", Fallback: "6vh"}
	tokenHeaderHeight  = Token{Name: "--pd-header-height", Fallback: "5vh"}
	tokenContentHeight = Token{Name: "--pd-content-height", Fallback: "94vh"}
	tokenContentWidth  = Token{Name: "--pd-content-width", Fallback: "100vw"}

	tokenSlideDur      = Token{Name: "--pd-slide-duration", Fallback: "0.6s"}
	tokenTransitionWait = Token{Name: "--pd-transition-wait", Fallback: "0s"}
)
