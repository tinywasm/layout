package crudview

import (
	"testing"

	"github.com/tinywasm/components/targetlist"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/form"
	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
	"github.com/tinywasm/view/conformance"
)

func setupBulkTest(t *testing.T, withUpdate bool) (*CrudView, *conformance.FakeCaller) {
	caller := &conformance.FakeCaller{
		Reply: func(op string, into model.Decodable) {
			if op == "device_list" {
				dl := into.(*DeviceList)
				d1 := dl.Append().(*Device)
				d1.Id = "id-1"
				d1.Name = "Device One"
				d1.Ip = "192.168.1.1"

				d2 := dl.Append().(*Device)
				d2.Id = "id-2"
				d2.Name = "Device Two"
				d2.Ip = "192.168.1.2"

				d3 := dl.Append().(*Device)
				d3.Id = "id-3"
				d3.Name = "Device Three"
				d3.Ip = "192.168.1.3"
			}
		},
	}
	opts := []view.Option{view.WithDeleteOp("device_delete")}
	if withUpdate {
		opts = append(opts, view.WithUpdateOp("device_update"))
	}
	p := view.New(caller, &Device{}, "device_list", func() model.ModelSlice { return &DeviceList{} }, opts...)

	cfg := Config{
		ParentID:  "bulk-test-parent",
		Presenter: p,
		IDs:       testIDs,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating view: %v", err)
	}
	v.Init(&fakeCtx{})

	// Ensure real targetlist.TargetList and form.New are used by the view
	if _, ok := v.list.(*targetlist.TargetList); !ok {
		t.Fatal("expected real targetlist.TargetList")
	}
	if v.form == nil {
		t.Fatal("expected real form.Form from form.New")
	}

	return v, caller
}

func TestBulkEntriesOnlyInNormalMode(t *testing.T) {
	v, _ := setupBulkTest(t, true)

	if v.mode.Get() != string(modeNormal) {
		t.Errorf("expected initial mode to be modeNormal, got %q", v.mode.Get())
	}

	v.setMode(modeDeleting)
	if v.mode.Get() == string(modeNormal) {
		t.Error("expected mode not to be modeNormal in modeDeleting")
	}

	v.setMode(modeNormal)
	if v.mode.Get() != string(modeNormal) {
		t.Errorf("expected return to modeNormal, got %q", v.mode.Get())
	}
}

func TestBulkEntriesDisabledWhileEditingARecord(t *testing.T) {
	v, _ := setupBulkTest(t, true)

	if v.active() {
		t.Error("expected inactive state initially")
	}

	v.newAction()
	if !v.active() {
		t.Error("expected active() to be true while composing a draft")
	}

	v.undoAction()
	if v.active() {
		t.Error("expected active() to be false after undoAction")
	}

	v.selectAction(view.Item{ID: "id-1"})
	if !v.active() {
		t.Error("expected active() to be true when a row is selected")
	}
}

func TestDeleteEntryGoesStraightToSelection(t *testing.T) {
	v, _ := setupBulkTest(t, true)

	v.setMode(modeDeleting)
	if v.mode.Get() != string(modeDeleting) {
		t.Errorf("expected modeDeleting, got %q", v.mode.Get())
	}
}

func TestDeleteModeTurnsOnListSelection(t *testing.T) {
	v, _ := setupBulkTest(t, true)

	v.setMode(modeDeleting)
	if len(v.list.CheckedIDs()) != 0 {
		t.Errorf("expected 0 checked IDs initially, got %v", v.list.CheckedIDs())
	}
}

func TestCommitDisabledWithNothingChecked(t *testing.T) {
	v, caller := setupBulkTest(t, true)

	v.setMode(modeDeleting)
	v.bulkDeleteAction()
	if v.deleteLabel.Get() == "3" {
		t.Error("bulkDeleteAction should be no-op when nothing checked")
	}

	v.setMode(modeEditing)
	v.bulkEditAction()
	for _, c := range caller.Calls {
		if c.Op == "device_update" {
			t.Error("bulkEditAction called Update with nothing checked")
		}
	}
}

func TestCancelClearsSelection(t *testing.T) {
	v, _ := setupBulkTest(t, true)

	v.setMode(modeDeleting)
	v.setMode(modeNormal)

	if v.mode.Get() != string(modeNormal) {
		t.Errorf("expected modeNormal after cancel, got %q", v.mode.Get())
	}
	if len(v.list.CheckedIDs()) != 0 {
		t.Errorf("expected checked IDs to be cleared, got %v", v.list.CheckedIDs())
	}
}

func TestBulkEditSuspendsAutoSave(t *testing.T) {
	v, caller := setupBulkTest(t, true)

	v.setMode(modeEditing)
	v.form.SetValues("name", "Auto Save Test")
	v.autoSaveAction()

	for _, c := range caller.Calls {
		if c.Op == "device_save" {
			t.Error("autoSaveAction called Save during modeEditing")
		}
	}
}

func TestBulkEditRefusesWithNoDirtyFields(t *testing.T) {
	v, caller := setupBulkTest(t, true)

	v.setMode(modeEditing)
	// No fields touched
	v.bulkEditAction()

	for _, c := range caller.Calls {
		if c.Op == "device_update" {
			t.Error("bulkEditAction executed Update with no dirty fields")
		}
	}
}

func TestEditButtonAbsentWithoutUpdater(t *testing.T) {
	v, _ := setupBulkTest(t, false) // withUpdate = false

	html := v.Render().String()
	if fmt.Contains(html, "cv-crudedit") {
		t.Error("expected edit button cv-crudedit to be absent without view.Updater")
	}
}

func TestRowTapStillLoadsRecordInNormalMode(t *testing.T) {
	v, _ := setupBulkTest(t, true)

	v.selectAction(view.Item{ID: "id-1"})

	if v.selected.Get() != "id-1" {
		t.Errorf("expected selected ID 'id-1', got %q", v.selected.Get())
	}
	f := v.form
	nameInput := f.Input("name")
	if len(nameInput.GetValues()) == 0 || nameInput.GetValues()[0] != "Device One" {
		t.Errorf("expected 'Device One' loaded into form, got %v", nameInput.GetValues())
	}
}

// Ensure form.New import is referenced
var _ = form.New

// TestBulkDeleteShipsOneCall and TestBulkEditShipsOnlyDirtyFields used to live
// here and asserted nothing: with no way to check a row under SSR — a row is
// marked by a DOM click — they called deleter.Delete(...) and
// updater.Update(...) directly, which tests view, not crudview, and would pass
// with bulkDeleteAction and bulkEditAction entirely broken.
//
// Both assertions now run for real in bulk_wasm_test.go, where the click is a
// click: TestBulkDelete_ShipsEveryCheckedIDInOneCall and
// TestBulkEdit_WritesOnlyTheFieldsTheUserTouched.
