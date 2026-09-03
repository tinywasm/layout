package crudview

import (
	"strings"
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

// A loaded record deletes directly: the same click that opens selection mode
// with nothing loaded opens that record's own confirmation instead — no mode
// change, one dialog naming the record.
func TestSingleDeleteEntryConfirmsLoadedRecord(t *testing.T) {
	v, caller := setupBulkTest(t, true)

	v.selectAction(view.Item{ID: "id-1", Label: "Device One"})
	v.deleteEntryAction()

	if v.mode.Get() != string(modeNormal) {
		t.Errorf("a single delete must not enter selection mode, got %q", v.mode.Get())
	}
	if v.deleteID.Get() != "id-1" {
		t.Errorf("the confirmation must name the loaded record, got %q", v.deleteID.Get())
	}

	v.confirmDeleteAction()

	count := 0
	for _, c := range caller.Calls {
		if c.Op != "device_delete" {
			continue
		}
		sent := conformance.Payload(c.Args)
		if !conformance.Has(sent, "ids", "id-1") {
			t.Errorf("the single delete must ship id-1, got %v", sent)
		}
		for _, kv := range sent {
			if kv.Key == "ids" {
				count++
			}
		}
	}
	if count != 1 {
		t.Errorf("expected exactly one deleted id across all calls, got %d", count)
	}
}

// An unsaved draft has no id to name in the dialog: the button is dead for
// it (and disabled in the skin), so nothing happens and no mode opens.
func TestSingleDeleteEntryWithDraftDoesNothing(t *testing.T) {
	v, _ := setupBulkTest(t, true)

	v.newAction()
	v.deleteEntryAction()

	if v.mode.Get() != string(modeNormal) {
		t.Errorf("a draft delete must not enter selection mode, got %q", v.mode.Get())
	}
	if v.deleteID.Get() != "" {
		t.Errorf("no record may pend confirmation for a draft, got %q", v.deleteID.Get())
	}
}

// With nothing loaded the same button still opens selection mode.
func TestDeleteEntryWithNothingEntersSelection(t *testing.T) {
	v, _ := setupBulkTest(t, true)

	v.deleteEntryAction()

	if v.mode.Get() != string(modeDeleting) {
		t.Errorf("expected modeDeleting, got %q", v.mode.Get())
	}
}

// footerTag returns the footer's own opening tag: the tone lives on the
// footer div itself, while the buttons inside carry their own visibility
// states — grepping the whole subtree would mix the two.
func footerTag(v *CrudView) string {
	v.Render()
	html := v.panel.AsideFooter.String()
	if i := strings.Index(html, ">"); i != -1 {
		return html[:i+1]
	}
	return html
}

// The footer's Open state is the delete-mode tone: it holds exactly while
// deleting, so the stylesheet's Within rule paints the delete button red
// only there.
func TestFooterToneHoldsOnlyWhileDeleting(t *testing.T) {
	v, _ := setupBulkTest(t, true)

	if tag := footerTag(v); tag != "<div class='crudview__footer'>" {
		t.Errorf("the footer must not carry the tone in normal mode, tag: %s", tag)
	}

	v.setMode(modeDeleting)
	if tag := footerTag(v); !strings.Contains(tag, "data-open='true'") {
		t.Errorf("the footer must carry data-open='true' while deleting, tag: %s", tag)
	}

	v.setMode(modeNormal)
	if tag := footerTag(v); tag != "<div class='crudview__footer'>" {
		t.Errorf("leaving delete mode must drop the tone, tag: %s", tag)
	}
}

// Gap fix: the "+" create action is gated on view.Saver, like 🗑 on Deleter
// and ✏ on Updater. setupBulkTest builds a Deleter(+Updater)-only presenter
// (no WithSaveOp), so the toggle's "+" must be a no-op and render disabled in
// normal mode — but the SAME button still has to work as "↺" to leave
// selection mode.
func TestCreateActionGatedOnSaver(t *testing.T) {
	v, _ := setupBulkTest(t, true) // no WithSaveOp → not a view.Saver
	if _, ok := v.Presenter.(view.Saver); ok {
		t.Fatal("precondition: this presenter must not be a view.Saver")
	}

	// "+" is inert: toggleAction in normal mode must not start a draft.
	v.toggleAction()
	if v.composing.Get() || v.selected.Get() != "" {
		t.Errorf("the + must not create without a Saver (composing=%v selected=%q)",
			v.composing.Get(), v.selected.Get())
	}

	// The toggle renders disabled in that state.
	html := v.Render().String()
	i := strings.Index(html, "name='cv-crudtoggle'")
	if i == -1 {
		t.Fatal("the toggle button must still render (it is also the ↺)")
	}
	tag := html[strings.LastIndex(html[:i], "<"):]
	if e := strings.Index(tag, ">"); e != -1 {
		tag = tag[:e+1]
	}
	if !strings.Contains(tag, "disabled") {
		t.Errorf("the toggle must render disabled when its only meaning is a dead +, tag: %s", tag)
	}

	// But it still leaves selection mode: enter delete mode, the toggle is
	// live again (it means ↺ now).
	v.setMode(modeDeleting)
	html = v.Render().String()
	i = strings.Index(html, "name='cv-crudtoggle'")
	tag = html[strings.LastIndex(html[:i], "<"):]
	if e := strings.Index(tag, ">"); e != -1 {
		tag = tag[:e+1]
	}
	if strings.Contains(tag, "disabled='true'") {
		t.Errorf("the toggle must be live as ↺ while in a selection mode, tag: %s", tag)
	}
}

// buttonOpenTag returns the opening <button ...> tag for the named control.
func buttonOpenTag(html, name string) string {
	i := strings.Index(html, "name='"+name+"'")
	if i == -1 {
		return ""
	}
	start := strings.LastIndex(html[:i], "<")
	end := strings.Index(html[i:], ">")
	if start == -1 || end == -1 {
		return ""
	}
	return html[start : i+end+1]
}

// The footer collapses to what is actionable: 🗑/✏ show only when there is a
// record to act on. Empty list → just "+". Composing a new record → just "↺".
// A loaded existing row → 🗑 (delete it) but not ✏ (bulk would discard the
// form). Same spirit as the two selection modes, which already hide the button
// they are not.
func TestFooterHidesBulkActionsWhenNothingToActOn(t *testing.T) {
	v, _ := setupBulkTest(t, true) // 3 seeded devices, Deleter+Updater

	shown := func(name string) bool {
		return strings.Contains(buttonOpenTag(v.Render().String(), name), "data-open='true'")
	}

	// Rows exist, nothing loaded: both bulk actions offered.
	if !shown("cv-cruddelete") || !shown("cv-crudedit") {
		t.Fatalf("with rows and nothing loaded, 🗑 and ✏ must both show")
	}

	// Empty list: only "+".
	v.search.Set("no-such-device-xyz")
	v.filter()
	if shown("cv-cruddelete") || shown("cv-crudedit") {
		t.Errorf("an empty list must hide 🗑 and ✏ (only + / ↺ remain)")
	}
	if !strings.Contains(v.Render().String(), "name='cv-crudtoggle'") {
		t.Errorf("the + toggle must always be present")
	}

	// Exactly ONE row: 🗑 stays (single delete is real), ✏ hides (bulk edit
	// needs a set).
	v.search.Set("Device One") // the seeded label — filters to a single row
	v.filter()
	if v.list.Count() != 1 {
		t.Fatalf("precondition: expected exactly 1 row, got %d", v.list.Count())
	}
	if !shown("cv-cruddelete") {
		t.Errorf("with one row 🗑 must still show (single delete is valid)")
	}
	if shown("cv-crudedit") {
		t.Errorf("with one row ✏ must hide — bulk edit needs 2+ rows")
	}

	// Back to a populated list, then compose a new record.
	v.search.Set("")
	v.filter()
	v.newAction()
	if shown("cv-cruddelete") || shown("cv-crudedit") {
		t.Errorf("composing a new record must hide 🗑 and ✏ (only ↺ remains)")
	}

	// Cancel the draft, load an existing row.
	v.undoAction()
	v.selectAction(view.Item{ID: "id-1"})
	if !shown("cv-cruddelete") {
		t.Errorf("a loaded existing row must still offer 🗑 to delete it")
	}
	if shown("cv-crudedit") {
		t.Errorf("a loaded row must hide ✏ — bulk edit would discard the form")
	}
}
