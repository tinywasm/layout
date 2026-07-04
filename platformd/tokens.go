//go:build !wasm

package platformd

import . "github.com/tinywasm/css"

var (
	tokenFontSizeNormal = Token{Name: "--pd-font-size-normal", Fallback: "1.1rem"}
	tokenFontSizeSmall  = Token{Name: "--pd-font-size-small", Fallback: ".6rem"}

	tokenMenuSize      = Token{Name: "--pd-menu-size", Fallback: "4vw"}
	tokenHeaderHeight  = Token{Name: "--pd-header-height", Fallback: "3vh"}
	tokenContentHeight = Token{Name: "--pd-content-height", Fallback: "97vh"}
	tokenContentWidth  = Token{Name: "--pd-content-width", Fallback: "96vw"}

	tokenSlideDur       = Token{Name: "--pd-slide-duration", Fallback: "0.6s"}
	tokenTransitionWait = Token{Name: "--pd-transition-wait", Fallback: "0s"}
)
