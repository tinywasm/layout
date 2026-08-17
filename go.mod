module github.com/tinywasm/layout

go 1.25.2

require (
	github.com/tinywasm/components v0.5.12
	github.com/tinywasm/css v0.4.14
	github.com/tinywasm/dom v0.13.5
	github.com/tinywasm/fmt v0.25.5
	github.com/tinywasm/form v0.3.29
	github.com/tinywasm/html v0.0.17
	github.com/tinywasm/image v0.0.18
	github.com/tinywasm/input v0.0.3
	github.com/tinywasm/model v0.1.4
	github.com/tinywasm/orm v0.11.6
	github.com/tinywasm/storage v0.0.2
	github.com/tinywasm/svg v0.1.21
	github.com/tinywasm/time v0.5.2
	github.com/tinywasm/unixid v0.2.26
	github.com/tinywasm/view v0.1.17
	github.com/tinywasm/widget v0.6.10
)

// TEMPORAL — solo para probar en el iPhone el fix de Focus(preventScroll) en
// dom. Quitar y volver a la versión publicada en cuanto el test manual confirme
// el resultado (funcione o no).

require (
	github.com/tinywasm/color v0.1.1 // indirect
	github.com/tinywasm/context v0.0.18 // indirect
	github.com/tinywasm/fetch v0.1.24 // indirect
	github.com/tinywasm/font v0.0.4 // indirect
	github.com/tinywasm/js v0.0.6 // indirect
	github.com/tinywasm/json v0.5.19 // indirect
	github.com/tinywasm/router v0.1.21 // indirect
)
