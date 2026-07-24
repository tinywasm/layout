module github.com/tinywasm/layout

go 1.25.2

require (
	github.com/tinywasm/components v0.1.8
	github.com/tinywasm/css v0.1.4
	github.com/tinywasm/dom v0.11.4
	github.com/tinywasm/fmt v0.25.5
	github.com/tinywasm/form v0.2.25
	github.com/tinywasm/html v0.0.8
	github.com/tinywasm/model v0.1.0
	github.com/tinywasm/orm v0.11.4
	github.com/tinywasm/storage v0.0.2
	github.com/tinywasm/svg v0.1.8
	github.com/tinywasm/time v0.5.0
	github.com/tinywasm/view v0.1.2
)

require (
	github.com/tinywasm/json v0.5.12 // indirect
	github.com/tinywasm/router v0.1.13 // indirect
)

// TEMP: local dom checkout carries the BindChildren initial-row wiring fix
// (rows present at first render now get their nested bindings wired). Remove
// once that fix is released and this go.mod bumps to the published version.
replace github.com/tinywasm/dom => ../dom

// TEMP: local components checkout so demo edits render live via the dev server.
replace github.com/tinywasm/components => ../components

// TEMP: local form checkout — fixes the field-wrapper/input id collision that
// stopped the input from displaying loaded values.
replace github.com/tinywasm/form => ../form

// TEMP: local css checkout — brand palette set to the Pa100T reference (steel
// blue + white text).
replace github.com/tinywasm/css => ../css

// TEMP: local view checkout — conformance.Driver gained New/Edit/FocusedFieldID
// for the "+ / Editar focuses the first field" behavior.
replace github.com/tinywasm/view => ../view
