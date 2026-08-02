module github.com/tinywasm/layout

go 1.25.2

require (
	github.com/tinywasm/components v0.3.6
	github.com/tinywasm/css v0.4.2
	github.com/tinywasm/dom v0.12.1
	github.com/tinywasm/fmt v0.25.5
	github.com/tinywasm/form v0.3.9
	github.com/tinywasm/html v0.0.8
	github.com/tinywasm/model v0.1.0
	github.com/tinywasm/orm v0.11.4
	github.com/tinywasm/storage v0.0.2
	github.com/tinywasm/svg v0.1.8
	github.com/tinywasm/time v0.5.0
	github.com/tinywasm/view v0.1.10
	github.com/tinywasm/widget v0.5.1
)

require (
	github.com/tinywasm/json v0.5.12 // indirect
	github.com/tinywasm/router v0.1.13 // indirect
)

// ── replaces de desarrollo local ─────────────────────────────────────────────
// PLAN v0.2.0 spans layout/components/form/widget; everything works against
// local checkouts and is NOT published. Revert to published versions when the
// work lands.
replace github.com/tinywasm/widget => ../widget

replace github.com/tinywasm/components => ../components

replace github.com/tinywasm/form => ../form

replace github.com/tinywasm/css => ../css

//
//
