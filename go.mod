module github.com/tinywasm/layout

go 1.25.2

require (
	github.com/tinywasm/components v0.6.6
	github.com/tinywasm/css v0.4.20
	github.com/tinywasm/dom v0.13.8
	github.com/tinywasm/fmt v0.25.7
	github.com/tinywasm/form v0.4.0
	github.com/tinywasm/html v0.0.19
	github.com/tinywasm/image v0.1.0
	github.com/tinywasm/input v0.0.3
	github.com/tinywasm/model v0.1.7
	github.com/tinywasm/svg v0.3.3
	github.com/tinywasm/time v0.5.4
	github.com/tinywasm/view v0.2.3
	github.com/tinywasm/widget v0.6.21
)

// TEMPORAL — solo para probar en el iPhone el fix de Focus(preventScroll) en
// dom. Quitar y volver a la versión publicada en cuanto el test manual confirme
// el resultado (funcione o no).

require (
	github.com/tinywasm/color v0.1.1 // indirect
	github.com/tinywasm/context v0.0.18 // indirect
	github.com/tinywasm/fetch v0.1.24 // indirect
	github.com/tinywasm/font v0.0.4 // indirect
	github.com/tinywasm/icons v0.0.2
	github.com/tinywasm/js v0.0.6 // indirect
	github.com/tinywasm/json v0.5.23 // indirect
	github.com/tinywasm/router v0.1.30 // indirect
)

// TEMP: local until FloatingChrome consume + Filled cue + CueSibling ship in tinywasm/widget.
