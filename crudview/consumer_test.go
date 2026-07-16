package crudview

import (
	"testing"

	"github.com/tinywasm/fmt"
	"github.com/tinywasm/form"
	"github.com/tinywasm/form/input"
	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
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

type fakePresenter struct {
	title             string
	searchPlaceholder string
	record            model.Model
	items             []view.Item
	selected          string
	canSave           bool
	canDelete         bool

	onReload func() error
	onSelect func(id string) model.Model
	onSave   func(m model.Model) error
	onDelete func(id string) error

	reloaded bool
}

func (f *fakePresenter) Title() string             { return f.title }
func (f *fakePresenter) SearchPlaceholder() string { return f.searchPlaceholder }
func (f *fakePresenter) Record() model.Model       { return f.record }
func (f *fakePresenter) Items() []view.Item        { return f.items }
func (f *fakePresenter) Reload() error {
	f.reloaded = true
	if f.onReload != nil {
		return f.onReload()
	}
	return nil
}
func (f *fakePresenter) Selected() string { return f.selected }
func (f *fakePresenter) Select(id string) model.Model {
	f.selected = id
	if f.onSelect != nil {
		return f.onSelect(id)
	}
	return nil
}
func (f *fakePresenter) CanSave() bool { return f.canSave }
func (f *fakePresenter) Save(payload model.Model) error {
	if f.onSave != nil {
		return f.onSave(payload)
	}
	return nil
}
func (f *fakePresenter) CanDelete() bool { return f.canDelete }
func (f *fakePresenter) Delete(id string) error {
	if f.onDelete != nil {
		return f.onDelete(id)
	}
	return nil
}

// Case 1: New with a model without widgets fails
func TestConsumer_NewNoWidgets(t *testing.T) {
	p := &fakePresenter{record: &DeviceNoWidgets{}}
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
	p := &fakePresenter{record: &Device{}}
	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	if !p.reloaded {
		t.Errorf("expected Presenter to be reloaded on Init")
	}
}

// Case 3: The list renders cards returned by Items
func TestConsumer_ListRendersCards(t *testing.T) {
	decoded := []view.Item{
		{ID: "12", Label: "Device One", Description: "192.168.1.1"},
		{ID: "23", Label: "Device Two", Description: "192.168.1.2"},
	}
	p := &fakePresenter{record: &Device{}, items: decoded}

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
	devices := map[string]*Device{
		"12": {Id: "12", Name: "Device One", Ip: "192.168.1.1"},
		"23": {Id: "23", Name: "Device Two", Ip: "192.168.1.2"},
	}
	p := &fakePresenter{record: &Device{}}
	p.onSelect = func(id string) model.Model {
		dev := devices[id]
		if dev == nil {
			return nil
		}
		return dev
	}

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

	v.OnSelect(view.Item{ID: "12"})

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
	v.OnSelect(view.Item{ID: "unknown"})
	if len(idInput.GetValues()) != 0 && idInput.GetValues()[0] != "" {
		t.Errorf("expected id input to be reset/empty, got %v", idInput.GetValues())
	}
}

// Case 5: Save calls Save with form data, not original Record
func TestConsumer_SaveWithFormData(t *testing.T) {
	rec := &Device{Id: "12", Name: "Original Name", Ip: "192.168.1.1"}
	p := &fakePresenter{record: rec, canSave: true}
	var savedDevice *Device
	p.onSave = func(m model.Model) error {
		savedDevice = m.(*Device)
		return nil
	}

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

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

	if savedDevice == nil {
		t.Fatal("expected save to be called on presenter")
	}
	if savedDevice.Name != "New Name" {
		t.Errorf("expected sent device name to be 'New Name', got '%s'", savedDevice.Name)
	}
	if savedDevice.Ip != "10.0.0.1" {
		t.Errorf("expected sent device ip to be '10.0.0.1', got '%s'", savedDevice.Ip)
	}
}

// Case 6: Save with invalid form doesn't call presenter and returns error
func TestConsumer_SaveInvalidForm(t *testing.T) {
	rec := &Device{Id: "12", Name: "Original Name", Ip: "192.168.1.1"}
	p := &fakePresenter{record: rec, canSave: true}
	var saveCalled bool
	p.onSave = func(m model.Model) error {
		saveCalled = true
		return nil
	}

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

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
	if saveCalled {
		t.Errorf("expected no save to be called on presenter because form was invalid")
	}
}

// Case 7: Delete calls Delete on presenter
func TestConsumer_DeleteSelected(t *testing.T) {
	p := &fakePresenter{record: &Device{}, canDelete: true}
	var deletedID string
	p.onDelete = func(id string) error {
		deletedID = id
		return nil
	}

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

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

	if deletedID != "123" {
		t.Errorf("expected deleted id on presenter to be '123', got '%s'", deletedID)
	}
}

// Case 8: Delete can return error from presenter
func TestConsumer_DeleteNoSelection(t *testing.T) {
	p := &fakePresenter{record: &Device{}, canDelete: true}
	p.onDelete = func(id string) error {
		return fmt.Errf("no selection")
	}

	cfg := Config{
		ParentID:  "my-id",
		Presenter: p,
	}
	v, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v.Init(&fakeCtx{})

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
		t.Error("expected delete error because presenter returned error, got nil")
	}
}

// Case 9: Presenter error on list is propagated to Reload caller
func TestConsumer_ListErrorPropagated(t *testing.T) {
	expectedErr := fmt.Errf("network connection failed")
	p := &fakePresenter{record: &Device{}}
	p.onReload = func() error {
		return expectedErr
	}

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
	decoded := []view.Item{
		{ID: "12", Label: "Frontend Device", Description: "192.168.1.10"},
		{ID: "23", Label: "Backend Server", Description: "10.0.0.5"},
		{ID: "34", Label: "Database Instance", Description: "mysql-production"},
	}
	p := &fakePresenter{record: &Device{}, items: decoded}

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
