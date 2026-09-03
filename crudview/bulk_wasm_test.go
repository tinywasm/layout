//go:build wasm

package crudview

import (
	"syscall/js"
	"testing"

	. "github.com/tinywasm/dom"
	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
	"github.com/tinywasm/view/conformance"
)

// The SSR bulk tests cannot mark a row: a row is checked by a DOM click and
// there is no click to dispatch without a browser. So the two assertions the
// whole plan exists for — "the batch ships in ONE call" and "only the fields
// the user touched are written" — live here, where the click is real and the
// path runs end to end through crudview's own methods.
//
// Anything that calls deleter.Delete(...) or updater.Update(...) directly is
// testing view, not crudview: it would pass with bulkDeleteAction and
// bulkEditAction entirely broken.

func mountBulk(t *testing.T, withUpdate bool) (*CrudView, *conformance.FakeCaller, js.Value) {
	t.Helper()
	doc := js.Global().Get("document")
	root := doc.Call("createElement", "div")
	root.Set("id", "cv-bulk-root")
	doc.Get("body").Call("appendChild", root)
	t.Cleanup(func() { root.Set("innerHTML", "") })

	caller := &conformance.FakeCaller{
		Reply: func(op string, into model.Decodable) {
			if op != "device_list" {
				return
			}
			dl := into.(*DeviceList)
			for _, id := range []string{"id-1", "id-2", "id-3"} {
				d := dl.Append().(*Device)
				d.Id = id
				d.Name = "Device " + id
				d.Ip = "10.0.0." + id
			}
		},
	}
	opts := []view.Option{view.WithDeleteOp("device_delete"), view.WithSaveOp("device_save")}
	if withUpdate {
		opts = append(opts, view.WithUpdateOp("device_update"))
	}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} }, opts...)

	// Through New(), not a bare struct literal: New is what builds the real
	// form from the record's schema, and bulk edit is nothing without it.
	v, err := New(Config{ParentID: "cv-bulk-root", Presenter: p, IDs: testIDs})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	v.Init(&mockCtxWasm{})
	v.SetID("cvb")
	_ = v.Reload()
	if err := Render("cv-bulk-root", v); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if v.form == nil {
		t.Fatal("expected the real form from form.New")
	}
	return v, caller, doc
}

func clickRow(t *testing.T, doc js.Value, id string) {
	t.Helper()
	row := doc.Call("getElementById", id)
	if row.IsNull() {
		t.Fatalf("row %s not mounted", id)
	}
	row.Call("click")
}

func callsFor(caller *conformance.FakeCaller, op string) []conformance.FakeCall {
	var out []conformance.FakeCall
	for _, c := range caller.Calls {
		if c.Op == op {
			out = append(out, c)
		}
	}
	return out
}

func TestBulkDelete_ShipsEveryCheckedIDInOneCall(t *testing.T) {
	v, caller, doc := mountBulk(t, false)

	v.setMode(modeDeleting)
	clickRow(t, doc, "tl-id-1")
	clickRow(t, doc, "tl-id-3")

	if got := v.list.CheckedIDs(); len(got) != 2 {
		t.Fatalf("expected 2 rows checked after two clicks, got %v", got)
	}

	v.confirmDeleteAction()

	calls := callsFor(caller, "device_delete")
	if len(calls) != 1 {
		t.Fatalf("expected exactly 1 delete call for a 2-row batch, got %d", len(calls))
	}
	sent := conformance.Payload(calls[0].Args)
	if !conformance.Has(sent, "ids", "id-1") || !conformance.Has(sent, "ids", "id-3") {
		t.Errorf("both checked ids must ship in the one call, got %v", sent)
	}
	if conformance.Has(sent, "ids", "id-2") {
		t.Errorf("an unchecked row must not be deleted, got %v", sent)
	}
}

func TestBulkDelete_ChecksFollowRenderOrder(t *testing.T) {
	v, _, doc := mountBulk(t, false)

	v.setMode(modeDeleting)
	// Tap the LAST row first: the order that reaches the confirmation message
	// must be the order on screen, not the order of tapping.
	clickRow(t, doc, "tl-id-3")
	clickRow(t, doc, "tl-id-1")

	got := v.list.CheckedIDs()
	if len(got) != 2 || got[0] != "id-1" || got[1] != "id-3" {
		t.Errorf("CheckedIDs must follow render order [id-1 id-3], got %v", got)
	}
}

func TestBulkEdit_WritesOnlyTheFieldsTheUserTouched(t *testing.T) {
	v, caller, doc := mountBulk(t, true)

	v.setMode(modeEditing)
	clickRow(t, doc, "tl-id-1")
	clickRow(t, doc, "tl-id-2")

	// One field out of the record's several.
	v.form.SetValues("name", "Renamed")
	v.autoSaveAction() // the form's own commit hook; must NOT save in this mode

	if n := len(callsFor(caller, "device_save")); n != 0 {
		t.Errorf("auto-save must stay suspended in bulk edit, got %d save calls", n)
	}

	v.bulkEditAction()

	calls := callsFor(caller, "device_update")
	if len(calls) != 1 {
		t.Fatalf("expected exactly 1 update call for a 2-row batch, got %d", len(calls))
	}
	sent := conformance.Payload(calls[0].Args)
	if !conformance.Has(sent, "ids", "id-1") || !conformance.Has(sent, "ids", "id-2") {
		t.Errorf("both checked ids must ship, got %v", sent)
	}
	if !conformance.Has(sent, "fields", "name") {
		t.Errorf("the touched field must be named, got %v", sent)
	}
	if conformance.Has(sent, "fields", "ip") {
		t.Errorf("an untouched field must never be written — that is the lost update this design exists to prevent, got %v", sent)
	}
}

func TestBulkDelete_OnDeletedReceivesEveryID(t *testing.T) {
	v, _, doc := mountBulk(t, false)

	v.setMode(modeDeleting)
	clickRow(t, doc, "tl-id-1")
	clickRow(t, doc, "tl-id-2")

	var gotIDs []string
	var gotErr error
	called := false
	v.OnDeleted = func(ids []string, err error) {
		called = true
		gotIDs = ids
		gotErr = err
	}

	v.confirmDeleteAction()

	if !called {
		t.Fatal("expected OnDeleted to fire for the bulk delete")
	}
	if gotErr != nil {
		t.Fatalf("unexpected error: %v", gotErr)
	}
	if len(gotIDs) != 2 || gotIDs[0] != "id-1" || gotIDs[1] != "id-2" {
		t.Errorf("expected OnDeleted ids [id-1 id-2], got %v", gotIDs)
	}
}

func TestBulkEdit_ApplyIsDeadUntilSomethingIsEdited(t *testing.T) {
	v, _, doc := mountBulk(t, true)

	v.setMode(modeEditing)
	clickRow(t, doc, "tl-id-1")

	if v.hasEdits.Get() {
		t.Error("entering bulk edit with a blank form must leave nothing to apply")
	}

	v.form.SetValues("name", "Renamed")
	v.autoSaveAction()

	if !v.hasEdits.Get() {
		t.Error("touching a field must arm the apply button")
	}
}
