package crudview

import (
	"testing"

	"github.com/tinywasm/fmt"
	"github.com/tinywasm/form"
	"github.com/tinywasm/form/input"
	"github.com/tinywasm/model"
)

// fakeCtx is a simple implementation of Ctx for testing
type fakeCtx struct{}

func (f *fakeCtx) OnCleanup(fn func()) {}

// deviceModel is a model with real widgets (input.Text() is a model.Kind)
var deviceModel = model.Definition{
	Name: "device",
	Fields: model.Fields{
		{Name: "id", Type: input.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "name", Type: input.Text(), NotNull: true},
		{Name: "ip", Type: input.Text()},
	},
}

type Device struct {
	Id, Name, Ip string
}

func (d *Device) ModelName() string     { return "device" }
func (d *Device) Schema() []model.Field { return deviceModel.Fields }
func (d *Device) Pointers() []any       { return []any{&d.Id, &d.Name, &d.Ip} }
func (d *Device) IsNil() bool           { return d == nil }

func (d *Device) EncodeFields(w model.FieldWriter) {
	w.String("id", d.Id)
	w.String("name", d.Name)
	w.String("ip", d.Ip)
}

func (d *Device) DecodeFields(r model.FieldReader) {
	d.Id, _ = r.String("id")
	d.Name, _ = r.String("name")
	d.Ip, _ = r.String("ip")
}

var _ model.Model = (*Device)(nil) // Verify compile-time implementation

// deviceNoWidgetsModel is a model without widgets
var deviceNoWidgetsModel = model.Definition{
	Name: "device_no_widgets",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "name", Type: model.Text(), NotNull: true},
	},
}

type DeviceNoWidgets struct {
	Id, Name string
}

func (d *DeviceNoWidgets) ModelName() string     { return "device_no_widgets" }
func (d *DeviceNoWidgets) Schema() []model.Field { return deviceNoWidgetsModel.Fields }
func (d *DeviceNoWidgets) Pointers() []any       { return []any{&d.Id, &d.Name} }
func (d *DeviceNoWidgets) IsNil() bool           { return d == nil }

func (d *DeviceNoWidgets) EncodeFields(w model.FieldWriter) {
	w.String("id", d.Id)
	w.String("name", d.Name)
}

func (d *DeviceNoWidgets) DecodeFields(r model.FieldReader) {
	d.Id, _ = r.String("id")
	d.Name, _ = r.String("name")
}

var _ model.Model = (*DeviceNoWidgets)(nil) // Verify compile-time implementation

// fakeCaller registers called operations, arguments and returns stubbed values.
type fakeCaller struct {
	ops  []string
	args []model.Encodable
	resp []byte
	err  error
}

func (c *fakeCaller) Call(op string, args model.Encodable, cb func([]byte, error)) {
	c.ops = append(c.ops, op)
	c.args = append(c.args, args)
	cb(c.resp, c.err)
}

func (c *fakeCaller) Dispatch(op string, args model.Encodable) {
	c.Call(op, args, func([]byte, error) {})
}

// Case 1: New with a model without widgets fails
func TestConsumer_NewNoWidgets(t *testing.T) {
	caller := &fakeCaller{}
	cfg := Config{
		ParentID: "my-id",
		Caller:   caller,
		Record:   &DeviceNoWidgets{},
		ListOp:   "list_devices",
	}
	v, err := New(cfg)
	if err == nil {
		t.Error("expected New with widget-less record to fail, but it succeeded")
	}
	if v != nil {
		t.Error("expected returned view to be nil on error")
	}
}

// Case 2: The list operation on wiring is ListOp
func TestConsumer_ListOp(t *testing.T) {
	caller := &fakeCaller{}
	cfg := Config{
		ParentID: "my-id",
		Caller:   caller,
		Record:   &Device{},
		ListOp:   "list_devices",
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	if len(caller.ops) != 1 || caller.ops[0] != "list_devices" {
		t.Errorf("expected ListOp 'list_devices' to be called on Init, got %v", caller.ops)
	}
}

// Case 3: The list renders cards returned by Decode
func TestConsumer_ListRendersCards(t *testing.T) {
	caller := &fakeCaller{resp: []byte("some-raw-data")}
	decoded := []Item{
		{ID: "12", Label: "Device One", Description: "192.168.1.1"},
		{ID: "23", Label: "Device Two", Description: "192.168.1.2"},
	}
	cfg := Config{
		ParentID: "my-id",
		Caller:   caller,
		Record:   &Device{},
		ListOp:   "list_devices",
		Decode: func(raw []byte) ([]Item, error) {
			if string(raw) != "some-raw-data" {
				t.Errorf("unexpected raw data to decode: %s", string(raw))
			}
			return decoded, nil
		},
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	if len(v.allItems) != 2 {
		t.Errorf("expected 2 items, got %d", len(v.allItems))
	}
	if len(v.items.Get()) != 2 {
		t.Errorf("expected 2 rendered cards, got %d", len(v.items.Get()))
	}
}

// Case 4: Selecting a card populates the form using form.LoadValues
func TestConsumer_SelectPopulatesForm(t *testing.T) {
	caller := &fakeCaller{}
	devices := map[string]*Device{
		"12": {Id: "12", Name: "Device One", Ip: "192.168.1.1"},
		"23": {Id: "23", Name: "Device Two", Ip: "192.168.1.2"},
	}
	cfg := Config{
		ParentID: "my-id",
		Caller:   caller,
		Record:   &Device{},
		ListOp:   "list_devices",
		Fill: func(id string) model.Model {
			return devices[id]
		},
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	v.OnSelect(Item{ID: "12"})

	// Check form values
	f := v.Form.(*form.Form)
	idInput := f.Input("id")
	nameInput := f.Input("name")
	ipInput := f.Input("ip")

	if len(idInput.GetValues()) == 0 || idInput.GetValues()[0] != "12" {
		t.Errorf("expected id input to be '12', got %v", idInput.GetValues())
	}
	if len(nameInput.GetValues()) == 0 || nameInput.GetValues()[0] != "Device One" {
		t.Errorf("expected name input to be 'Device One', got %v", nameInput.GetValues())
	}
	if len(ipInput.GetValues()) == 0 || ipInput.GetValues()[0] != "192.168.1.1" {
		t.Errorf("expected ip input to be '192.168.1.1', got %v", ipInput.GetValues())
	}

	// Nil record on Fill should reset the form
	v.OnSelect(Item{ID: "unknown"})
	if len(idInput.GetValues()) != 0 && idInput.GetValues()[0] != "" {
		t.Errorf("expected id input to be reset/empty, got %v", idInput.GetValues())
	}
}

// Case 5: Save calls SaveOp with form data, not original Record
func TestConsumer_SaveWithFormData(t *testing.T) {
	caller := &fakeCaller{}
	rec := &Device{Id: "12", Name: "Original Name", Ip: "192.168.1.1"}
	cfg := Config{
		ParentID: "my-id",
		Caller:   caller,
		Record:   rec,
		ListOp:   "list_devices",
		SaveOp:   "save_device",
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	// Clear Init calls to focus on OnSave operations
	caller.ops = nil
	caller.args = nil

	f := v.Form.(*form.Form)
	f.SetValues("id", "12")
	f.SetValues("name", "New Name")
	f.SetValues("ip", "10.0.0.1")

	var saveDoneCalled bool
	var saveErr error
	v.OnSave(func(err error) {
		saveDoneCalled = true
		saveErr = err
	})

	if !saveDoneCalled {
		t.Fatal("expected save callback (done) to be called")
	}
	if saveErr != nil {
		t.Fatalf("expected no save error, got: %v", saveErr)
	}

	if len(caller.ops) != 1 || caller.ops[0] != "save_device" {
		t.Fatalf("expected SaveOp 'save_device' to be called, got %v", caller.ops)
	}

	sentDev, ok := caller.args[0].(*Device)
	if !ok {
		t.Fatalf("expected sent arg to be *Device, got %T", caller.args[0])
	}
	if sentDev.Name != "New Name" {
		t.Errorf("expected sent device name to be 'New Name', got '%s'", sentDev.Name)
	}
	if sentDev.Ip != "10.0.0.1" {
		t.Errorf("expected sent device ip to be '10.0.0.1', got '%s'", sentDev.Ip)
	}
}

// Case 6: Save with invalid form doesn't call caller and returns error
func TestConsumer_SaveInvalidForm(t *testing.T) {
	caller := &fakeCaller{}
	rec := &Device{Id: "12", Name: "Original Name", Ip: "192.168.1.1"}
	cfg := Config{
		ParentID: "my-id",
		Caller:   caller,
		Record:   rec,
		ListOp:   "list_devices",
		SaveOp:   "save_device",
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	// Clear Init calls to focus on OnSave operations
	caller.ops = nil
	caller.args = nil

	f := v.Form.(*form.Form)
	f.SetValues("id", "12")
	f.SetValues("name", "") // empty name violates NotNull
	f.SetValues("ip", "10.0.0.1")

	var saveDoneCalled bool
	var saveErr error
	v.OnSave(func(err error) {
		saveDoneCalled = true
		saveErr = err
	})

	if !saveDoneCalled {
		t.Error("expected save callback (done) to be called")
	}
	if saveErr == nil {
		t.Error("expected validation error, got nil")
	}
	if len(caller.ops) != 0 {
		t.Errorf("expected no caller operations, but got %v", caller.ops)
	}
}

// Case 7: Delete calls DeleteOp with selected record
func TestConsumer_DeleteSelected(t *testing.T) {
	caller := &fakeCaller{}
	devices := map[string]*Device{
		"123": {Id: "123", Name: "My Device", Ip: "10.10.10.10"},
	}
	cfg := Config{
		ParentID: "my-id",
		Caller:   caller,
		Record:   &Device{},
		ListOp:   "list_devices",
		DeleteOp: "delete_device",
		Fill: func(id string) model.Model {
			return devices[id]
		},
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	// Clear Init calls to focus on OnDelete operations
	caller.ops = nil
	caller.args = nil

	var deleteDoneCalled bool
	var deleteErr error
	v.OnDelete("123", func(err error) {
		deleteDoneCalled = true
		deleteErr = err
	})

	if !deleteDoneCalled {
		t.Error("expected delete callback to be called")
	}
	if deleteErr != nil {
		t.Errorf("expected no delete error, got: %v", deleteErr)
	}

	if len(caller.ops) != 1 || caller.ops[0] != "delete_device" {
		t.Fatalf("expected DeleteOp 'delete_device' to be called, got %v", caller.ops)
	}

	sentDev, ok := caller.args[0].(*Device)
	if !ok {
		t.Fatalf("expected sent arg to be *Device, got %T", caller.args[0])
	}
	if sentDev.Id != "123" || sentDev.Name != "My Device" {
		t.Errorf("unexpected sent device fields: %+v", sentDev)
	}
}

// Case 8: Delete without selection returns error without calling caller
func TestConsumer_DeleteNoSelection(t *testing.T) {
	caller := &fakeCaller{}
	cfg := Config{
		ParentID: "my-id",
		Caller:   caller,
		Record:   &Device{},
		ListOp:   "list_devices",
		DeleteOp: "delete_device",
		Fill: func(id string) model.Model {
			return nil
		},
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	// Clear Init calls to focus on OnDelete operations
	caller.ops = nil
	caller.args = nil

	var deleteDoneCalled bool
	var deleteErr error
	v.OnDelete("non-existent", func(err error) {
		deleteDoneCalled = true
		deleteErr = err
	})

	if !deleteDoneCalled {
		t.Error("expected delete callback to be called")
	}
	if deleteErr == nil {
		t.Error("expected delete error because of no selection, got nil")
	}
	if len(caller.ops) != 0 {
		t.Errorf("expected no caller operations, but got %v", caller.ops)
	}
}

// Case 9: Caller error on list is propagated to OnError
func TestConsumer_ListErrorPropagated(t *testing.T) {
	expectedErr := fmt.Errf("network connection failed")
	caller := &fakeCaller{err: expectedErr}
	var receivedErr error
	cfg := Config{
		ParentID: "my-id",
		Caller:   caller,
		Record:   &Device{},
		ListOp:   "list_devices",
		OnError: func(err error) {
			receivedErr = err
		},
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	if receivedErr == nil {
		t.Error("expected OnError to receive the caller error, but got nil")
	} else if receivedErr.Error() != expectedErr.Error() {
		t.Errorf("expected error '%s', got '%s'", expectedErr.Error(), receivedErr.Error())
	}
}

// Case 10: Search filters cards by Label and Description (case-insensitive)
func TestConsumer_SearchFiltering(t *testing.T) {
	caller := &fakeCaller{}
	decoded := []Item{
		{ID: "12", Label: "Frontend Device", Description: "192.168.1.10"},
		{ID: "23", Label: "Backend Server", Description: "10.0.0.5"},
		{ID: "34", Label: "Database Instance", Description: "mysql-production"},
	}
	cfg := Config{
		ParentID: "my-id",
		Caller:   caller,
		Record:   &Device{},
		ListOp:   "list_devices",
		Decode: func(raw []byte) ([]Item, error) {
			return decoded, nil
		},
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})
	v.Reload() // fetch and populate items

	// 1. Initial state (no search term) - expect all 3 items
	if len(v.items.Get()) != 3 {
		t.Errorf("expected 3 items initially, got %d", len(v.items.Get()))
	}

	// 2. Search by Label ("backend") case-insensitive - expect only 1 item
	v.search.Set("backend")
	v.filter()
	if len(v.items.Get()) != 1 {
		t.Errorf("expected 1 item for 'backend', got %d", len(v.items.Get()))
	}

	// 3. Search by Description ("mysql") case-insensitive - expect only 1 item
	v.search.Set("MYSQL")
	v.filter()
	if len(v.items.Get()) != 1 {
		t.Errorf("expected 1 item for 'MYSQL', got %d", len(v.items.Get()))
	}

	// 4. Search with no matches ("invalid-term") - expect 0 items
	v.search.Set("invalid-term")
	v.filter()
	if len(v.items.Get()) != 0 {
		t.Errorf("expected 0 items for 'invalid-term', got %d", len(v.items.Get()))
	}
}
