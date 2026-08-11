module github.com/tinywasm/layout

go 1.25.2

require (
	github.com/tinywasm/components v0.5.6
	github.com/tinywasm/css v0.4.12
	github.com/tinywasm/dom v0.13.5
	github.com/tinywasm/fmt v0.25.5
	github.com/tinywasm/form v0.3.28
	github.com/tinywasm/html v0.0.16
	github.com/tinywasm/input v0.0.3
	github.com/tinywasm/model v0.1.4
	github.com/tinywasm/orm v0.11.6
	github.com/tinywasm/storage v0.0.2
	github.com/tinywasm/svg v0.1.20
	github.com/tinywasm/time v0.5.2
	github.com/tinywasm/view v0.1.17
	github.com/tinywasm/widget v0.6.7
)

// TEMPORAL — solo para probar en el iPhone el fix de Focus(preventScroll) en
// dom. Quitar y volver a la versión publicada en cuanto el test manual confirme
// el resultado (funcione o no).

require (
	github.com/tinywasm/color v0.1.1 // indirect
	github.com/tinywasm/font v0.0.4 // indirect
	github.com/tinywasm/json v0.5.19 // indirect
	github.com/tinywasm/router v0.1.21 // indirect
	github.com/tinywasm/unixid v0.2.26 // indirect
)
