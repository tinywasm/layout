panic: dom: element div  is already a child of another element; one element has one parent — build a second instance instead of sharing this one
script.js:2 
script.js:2 goroutine 7 [running]:
script.js:2 github.com/tinywasm/dom.(*Element).Child(0x4d3130, {0x413bc8, 0x1, 0x1})
script.js:2 	/home/cesar/go/pkg/mod/github.com/tinywasm/dom@v0.12.1/element.go:118 +0x28
script.js:2 github.com/tinywasm/components/modaldialog.(*ModalDialog).Render.func2()
script.js:2 	/home/cesar/Dev/Project/tinywasm/components/modaldialog/modaldialog.go:74 +0x58
script.js:2 github.com/tinywasm/dom.Show.func1()
script.js:2 	/home/cesar/go/pkg/mod/github.com/tinywasm/dom@v0.12.1/dom_frontend.go:1317 +0x26
script.js:2 github.com/tinywasm/dom.notify({0x40f4a0, 0x1, 0x1})
script.js:2 	/home/cesar/go/pkg/mod/github.com/tinywasm/dom@v0.12.1/signal.go:21 +0xb
script.js:2 github.com/tinywasm/dom.(*SignalBool).Set(...)
script.js:2 	/home/cesar/go/pkg/mod/github.com/tinywasm/dom@v0.12.1/signal.go:105
script.js:2 github.com/tinywasm/components/modaldialog.(*ModalDialog).Open(...)
script.js:2 	/home/cesar/Dev/Project/tinywasm/components/modaldialog/modaldialog.go:79
script.js:2 github.com/tinywasm/layout/crudview.(*CrudView).deleteRequest(0x466000, {0x2f5c0, 0x2})
script.js:2 	/home/cesar/Dev/Project/tinywasm/layout/crudview/crudview.go:188 +0x48
script.js:2 github.com/tinywasm/layout/crudview.(*CrudView).Init.func3({0x2f5c0, 0x2})
script.js:2 	/home/cesar/Dev/Project/tinywasm/layout/crudview/crudview.go:103 +0x2
script.js:2 github.com/tinywasm/components/targetlist.(*TargetList).buildRow.func5({0x546a0, 0x5027f0})
script.js:2 	/home/cesar/Dev/Project/tinywasm/components/targetlist/targetlist.go:209 +0x9
script.js:2 github.com/tinywasm/dom.(*elementWasm).On.func1({{}, 0x7ff800010000006d, 0x504a00}, {0x5027e0, 0x1, 0x1})
script.js:2 	/home/cesar/go/pkg/mod/github.com/tinywasm/dom@v0.12.1/element_wasm.go:56 +0x8
script.js:2 syscall/js.handleEvent()
script.js:2 	/usr/local/go/src/syscall/js/func.go:117 +0x28
script.js:2 exit code: 2
exit @ script.js:2
runtime.wasmExit @ script.js:2
$runtime.wasmExit @ client.wasm:0xcc054
$runtime.exit @ client.wasm:0x10672c
$runtime.fatalpanic.func2 @ client.wasm:0xf6f4b
$runtime.systemstack @ client.wasm:0x103da2
$runtime.fatalpanic @ client.wasm:0x85d4d
$runtime.gopanic @ client.wasm:0xfb53d
$github.com_tinywasm_dom.__Element_.Child @ client.wasm:0x14dce4
$github.com_tinywasm_components_modaldialog.__ModalDialog_.Render.func2 @ client.wasm:0x197908
$github.com_tinywasm_dom.Show.func1 @ client.wasm:0x14c434
$github.com_tinywasm_dom.notify @ client.wasm:0x1578a5
$github.com_tinywasm_layout_crudview.__CrudView_.deleteRequest @ client.wasm:0x1a91dc
$github.com_tinywasm_layout_crudview.__CrudView_.Init.func3 @ client.wasm:0x1a6e34
$github.com_tinywasm_components_targetlist.__TargetList_.buildRow.func5 @ client.wasm:0x19f535
$github.com_tinywasm_dom.__elementWasm_.On.func1 @ client.wasm:0x155234
$syscall_js.handleEvent @ client.wasm:0x10cae4
$runtime.handleEvent @ client.wasm:0x24bba
$wasm_export_resume @ client.wasm:0x106711
_resume @ script.js:2
(anonymous) @ script.js:2
script.js:2 Uncaught Error: Go program has already exited
    at globalThis.Go._resume (script.js:2:7584)
    at HTMLDetailsElement.<anonymous> (script.js:2:7813)
_resume @ script.js:2
(anonymous) @ script.js:2
script.js:2 Uncaught Error: Go program has already exited
    at globalThis.Go._resume (script.js:2:7584)
    at HTMLElement.<anonymous> (script.js:2:7813)
_resume @ script.js:2
(anonymous) @ script.js:2
script.js:2 Uncaught Error: Go program has already exited
    at globalThis.Go._resume (script.js:2:7584)
    at HTMLLIElement.<anonymous> (script.js:2:7813)
_resume @ script.js:2
(anonymous) @ script.js:2
script.js:2 Uncaught Error: Go program has already exited
    at globalThis.Go._resume (script.js:2:7584)
    at HTMLDetailsElement.<anonymous> (script.js:2:7813)
_resume @ script.js:2
(anonymous) @ script.js:2
script.js:2 Uncaught Error: Go program has already exited
    at globalThis.Go._resume (script.js:2:7584)