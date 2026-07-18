package crudview

import (
	"testing"

	"github.com/tinywasm/fmt"
	"github.com/tinywasm/form/input"
	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
	"github.com/tinywasm/view/conformance"
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

func (d *Device) Item() view.Item {
	return view.Item{ID: d.Id, Label: d.Name, Description: d.Ip}
}

var _ model.Model = (*Device)(nil) // Verify compile-time implementation
var _ view.Itemizer = (*Device)(nil)

type DeviceList struct {
	Items []*Device
}

func (l *DeviceList) IsNil() bool { return l == nil }
func (l *DeviceList) DecodeFields(r model.FieldReader) {}
func (l *DeviceList) Schema() []model.Field { return nil }
func (l *DeviceList) Pointers() []any { return nil }
func (l *DeviceList) Len() int { return len(l.Items) }
func (l *DeviceList) At(i int) model.Fielder { return l.Items[i] }
func (l *DeviceList) Append() model.Fielder {
	d := &Device{}
	l.Items = append(l.Items, d)
	return d
}

var _ model.ModelSlice = (*DeviceList)(nil)

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

type fakeNoWidgetsPresenter struct {
	record model.Model
}

func (f *fakeNoWidgetsPresenter) Title() string             { return "No Widgets" }
func (f *fakeNoWidgetsPresenter) SearchPlaceholder() string { return "Search" }
func (f *fakeNoWidgetsPresenter) Record() model.Model       { return f.record }
func (f *fakeNoWidgetsPresenter) Items() []view.Item        { return nil }
func (f *fakeNoWidgetsPresenter) Filter(term string) []view.Item { return nil }
func (f *fakeNoWidgetsPresenter) Reload() error             { return nil }
func (f *fakeNoWidgetsPresenter) Selected() string          { return "" }
func (f *fakeNoWidgetsPresenter) Select(id string) model.Model { return nil }
func (f *fakeNoWidgetsPresenter) Deselect()                 {}

// Case 1: New with a model without widgets fails
func TestConsumer_NewNoWidgets(t *testing.T) {
	p := &fakeNoWidgetsPresenter{record: &DeviceNoWidgets{}}
	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err == nil {
		t.Error("expected New with widget-less record to fail, but it succeeded")
	}
	if v != nil {
		t.Error("expected returned view to be nil on error")
	}
}

// Case 2: The list operation on wiring is ListOp (reloaded on Init)
func TestConsumer_ListOp(t *testing.T) {
	var reloaded bool
	caller := &conformance.FakeCaller{
		Reply: func(op string, into model.Decodable) {
			reloaded = true
		},
	}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} })

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	if !reloaded {
		t.Errorf("expected Presenter to be reloaded on Init")
	}
}

// Case 3: The list renders cards returned by Items
func TestConsumer_ListRendersCards(t *testing.T) {
	caller := &conformance.FakeCaller{
		Reply: func(op string, into model.Decodable) {
			dl := into.(*DeviceList)
			d1 := dl.Append().(*Device)
			d1.Id = "12"
			d1.Name = "Device One"
			d1.Ip = "192.168.1.1"

			d2 := dl.Append().(*Device)
			d2.Id = "23"
			d2.Name = "Device Two"
			d2.Ip = "192.168.1.2"
		},
	}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} })

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	if len(v.Presenter.Items()) != 2 {
		t.Errorf("expected 2 items, got %d", len(v.Presenter.Items()))
	}
	if len(v.items.Get()) != 2 {
		t.Errorf("expected 2 rendered cards, got %d", len(v.items.Get()))
	}
}

// Case 4: Selecting a card populates the form using form.LoadValues
func TestConsumer_SelectPopulatesForm(t *testing.T) {
	caller := &conformance.FakeCaller{
		Reply: func(op string, into model.Decodable) {
			dl := into.(*DeviceList)
			d1 := dl.Append().(*Device)
			d1.Id = "12"
			d1.Name = "Device One"
			d1.Ip = "192.168.1.1"

			d2 := dl.Append().(*Device)
			d2.Id = "23"
			d2.Name = "Device Two"
			d2.Ip = "192.168.1.2"
		},
	}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} })

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	v.selectAction(view.Item{ID: "12"})

	// Check form values
	f := v.form
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
	v.selectAction(view.Item{ID: "unknown"})
	if len(idInput.GetValues()) != 0 && idInput.GetValues()[0] != "" {
		t.Errorf("expected id input to be reset/empty, got %v", idInput.GetValues())
	}
}

// Case 5: Save calls Save with form data, not original Record
func TestConsumer_SaveWithFormData(t *testing.T) {
	var savedDevice *Device
	caller := &conformance.FakeCaller{
		Reply: func(op string, into model.Decodable) {
			if op == "device_save" {
				// Handled in WithSaveOp or similar, but Caller gets Called
				// and we want to capture what was sent. FakeCaller.Calls can capture this.
			}
		},
	}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} },
		view.WithSaveOp("device_save"))

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	f := v.form
	f.SetValues("id", "12")
	f.SetValues("name", "New Name")
	f.SetValues("ip", "10.0.0.1")

	var saveDoneCalled bool
	var saveErr error
	v.OnSaved = func(err error) {
		saveDoneCalled = true
		saveErr = err
	}

	saver, ok := p.(view.Saver)
	if !ok {
		t.Fatal("expected view.Presenter to implement view.Saver")
	}
	v.saveAction(saver)

	if !saveDoneCalled {
		t.Fatal("expected OnSaved hook to be called")
	}
	if saveErr != nil {
		t.Fatalf("expected no save error, got: %v", saveErr)
	}

	// Read from FakeCaller.Calls
	if len(caller.Calls) == 0 {
		t.Fatal("expected save call on fake caller")
	}
	// The first call should be "device_list" on Init, the second should be "device_save"
	var saveCall *conformance.FakeCall
	for _, c := range caller.Calls {
		if c.Op == "device_save" {
			saveCall = &c
			break
		}
	}
	if saveCall == nil {
		t.Fatal("expected device_save call to be recorded")
	}
	savedDevice = saveCall.Args.(*Device)
	if savedDevice.Name != "New Name" {
		t.Errorf("expected sent device name to be 'New Name', got '%s'", savedDevice.Name)
	}
	if savedDevice.Ip != "10.0.0.1" {
		t.Errorf("expected sent device ip to be '10.0.0.1', got '%s'", savedDevice.Ip)
	}
}

// Case 6: Save with invalid form doesn't call presenter and returns error
func TestConsumer_SaveInvalidForm(t *testing.T) {
	caller := &conformance.FakeCaller{}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} },
		view.WithSaveOp("device_save"))

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	f := v.form
	f.SetValues("id", "12")
	f.SetValues("name", "") // empty name violates NotNull
	f.SetValues("ip", "10.0.0.1")

	var saveDoneCalled bool
	var saveErr error
	v.OnSaved = func(err error) {
		saveDoneCalled = true
		saveErr = err
	}

	saver, ok := p.(view.Saver)
	if !ok {
		t.Fatal("expected view.Presenter to implement view.Saver")
	}
	v.saveAction(saver)

	if !saveDoneCalled {
		t.Error("expected OnSaved hook to be called")
	}
	if saveErr == nil {
		t.Error("expected validation error, got nil")
	}
	// Verify device_save was never called on caller
	for _, c := range caller.Calls {
		if c.Op == "device_save" {
			t.Error("device_save was called on fake caller but form was invalid")
		}
	}
}

// Case 7: Delete calls Delete on presenter
func TestConsumer_DeleteSelected(t *testing.T) {
	caller := &conformance.FakeCaller{
		Reply: func(op string, into model.Decodable) {
			if op == "device_list" {
				dl := into.(*DeviceList)
				d1 := dl.Append().(*Device)
				d1.Id = "123"
				d1.Name = "Device One"
				d1.Ip = "192.168.1.1"
			}
		},
	}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} },
		view.WithDeleteOp("device_delete"))

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	// Select the loaded device so the presenter indexes and registers selection
	v.selectAction(view.Item{ID: "123"})

	var deleteDoneCalled bool
	var deleteErr error
	var deletedID string
	v.OnDeleted = func(id string, err error) {
		deleteDoneCalled = true
		deleteErr = err
		deletedID = id
	}

	deleter, ok := p.(view.Deleter)
	if !ok {
		t.Fatal("expected view.Presenter to implement view.Deleter")
	}
	v.deleteAction(deleter, "123")

	if !deleteDoneCalled {
		t.Error("expected OnDeleted hook to be called")
	}
	if deleteErr != nil {
		t.Errorf("expected no delete error, got: %v", deleteErr)
	}
	if deletedID != "123" {
		t.Errorf("expected hook deleted id to be '123', got '%s'", deletedID)
	}

	// Verify delete op was called on caller
	var deleteCall *conformance.FakeCall
	for _, c := range caller.Calls {
		if c.Op == "device_delete" {
			deleteCall = &c
			break
		}
	}
	if deleteCall == nil {
		t.Fatal("expected device_delete call to be recorded")
	}
	// Check identity field of the passed model
	argsDev := deleteCall.Args.(*Device)
	if argsDev.Id != "123" {
		t.Errorf("expected deleted id on presenter to be '123', got '%s'", argsDev.Id)
	}
}

// Case 8: Delete can return error from presenter
func TestConsumer_DeleteNoSelection(t *testing.T) {
	caller := &conformance.FakeCaller{
		Reply: func(op string, into model.Decodable) {
			if op == "device_list" {
				dl := into.(*DeviceList)
				d1 := dl.Append().(*Device)
				d1.Id = "non-existent"
				d1.Name = "Device One"
				d1.Ip = "192.168.1.1"
			}
		},
	}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} },
		view.WithDeleteOp("device_delete"))

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	// Select the item so it is registered
	v.selectAction(view.Item{ID: "non-existent"})

	// Set caller.Err *after* successful Init/Reload
	expectedErr := fmt.Errf("no selection")
	caller.Err = expectedErr

	var deleteDoneCalled bool
	var deleteErr error
	v.OnDeleted = func(id string, err error) {
		deleteDoneCalled = true
		deleteErr = err
	}

	deleter, ok := p.(view.Deleter)
	if !ok {
		t.Fatal("expected view.Presenter to implement view.Deleter")
	}
	v.deleteAction(deleter, "non-existent")

	if !deleteDoneCalled {
		t.Error("expected OnDeleted hook to be called")
	}
	if deleteErr == nil {
		t.Error("expected delete error because presenter/caller returned error, got nil")
	}
}

// Case 9: Presenter error on list is propagated to Reload caller
func TestConsumer_ListErrorPropagated(t *testing.T) {
	expectedErr := fmt.Errf("network connection failed")
	caller := &conformance.FakeCaller{
		Err: expectedErr,
	}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} })

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	receivedErr := v.Reload()
	if receivedErr == nil {
		t.Error("expected Reload to return the presenter error, but got nil")
	} else if receivedErr.Error() != expectedErr.Error() {
		t.Errorf("expected error '%s', got '%s'", expectedErr.Error(), receivedErr.Error())
	}
}

// Case 10: Search filters cards by Label and Description (case-insensitive)
func TestConsumer_SearchFiltering(t *testing.T) {
	caller := &conformance.FakeCaller{
		Reply: func(op string, into model.Decodable) {
			dl := into.(*DeviceList)
			d1 := dl.Append().(*Device)
			d1.Id = "12"
			d1.Name = "Frontend Device"
			d1.Ip = "192.168.1.10"

			d2 := dl.Append().(*Device)
			d2.Id = "23"
			d2.Name = "Backend Server"
			d2.Ip = "10.0.0.5"

			d3 := dl.Append().(*Device)
			d3.Id = "34"
			d3.Name = "Database Instance"
			d3.Ip = "mysql-production"
		},
	}
	p := view.New(caller, &Device{}, "device_list",
		func() model.ModelSlice { return &DeviceList{} })

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})
	_ = v.Reload() // fetch and populate items

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
