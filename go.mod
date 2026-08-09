module github.com/tinywasm/layout

go 1.25.2

require (
	github.com/tinywasm/components v0.5.0
	github.com/tinywasm/css v0.4.11
	github.com/tinywasm/dom v0.13.5
	github.com/tinywasm/fmt v0.25.5
	github.com/tinywasm/form v0.3.17
	github.com/tinywasm/html v0.0.12
	github.com/tinywasm/model v0.1.2
	github.com/tinywasm/orm v0.11.4
	github.com/tinywasm/storage v0.0.2
	github.com/tinywasm/svg v0.1.8
	github.com/tinywasm/time v0.5.0
	github.com/tinywasm/view v0.1.10
	github.com/tinywasm/widget v0.6.3
)

// TEMPORAL — solo para probar en el iPhone el fix de Focus(preventScroll) en
// dom. Quitar y volver a la versión publicada en cuanto el test manual confirme
// el resultado (funcione o no).

require (
	github.com/tinywasm/color v0.1.1 // indirect
	github.com/tinywasm/font v0.0.4 // indirect
	github.com/tinywasm/json v0.5.17 // indirect
	github.com/tinywasm/router v0.1.13 // indirect
)
