package crudview

import (
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/html"

	"github.com/tinywasm/components/modaldialog"
	"github.com/tinywasm/components/targetlist"
	"github.com/tinywasm/form"
	"github.com/tinywasm/layout/rightpanel"
	"github.com/tinywasm/svg"
	"github.com/tinywasm/view"
	"github.com/tinywasm/widget"
)

const NameCrudView widget.Name = "crudview"

var (
	clsBoxContent          = NameCrudView.Class("fields")
	clsBtnCrud             = NameCrudView.Class("action")
	clsBtnCrudIconHidden   = NameCrudView.Class("action-hidden")
	clsListaBox            = NameCrudView.Class("list")
	clsDelConfirmBody      = NameCrudView.Class("delconfirm-body")
	clsDelConfirmActions   = NameCrudView.Class("delconfirm-actions")
	clsDelConfirmBtn       = NameCrudView.Class("delconfirm-btn")
	clsDelConfirmBtnDanger = NameCrudView.Class("delconfirm-btn-danger")
	clsDelConfirmMount     = NameCrudView.Class("delconfirm-mount")
)

const (
	// The single toggle button swaps between these two icons reactively — see
	// the "toggle" block in Render().
	iconCrudNew    = svg.Icon("icon-crud-new")    // "+"  — nothing selected
	iconCrudCancel = svg.Icon("icon-crud-cancel") // "↺" — a row is selected (undo)
)

func (v *CrudView) WidgetName() widget.Name { return NameCrudView }
func (v *CrudView) WidgetKind() widget.Kind { return widget.Disclosure }

type CrudView struct {
	Element // value embed — NEVER *dom.Element

	Title     string
	Form      Component // what Render paints (may stay nil in standalone mode)
	Presenter view.Presenter

	// Filter is the control that narrows the list — a searchbar.SearchBar, a
	// calendar, a select. nil paints no controls band at all. If it implements
	// widget.Filterable, crudview wires it to the list filter in Init.
	//
	// The type is Component, not widget.Filterable, on purpose: a control with
	// no live output (a static legend, a chip strip that navigates elsewhere) is
	// a legitimate occupant and must not be forced to implement a callback it
	// would never fire.
	Filter Component

	// Additive user hooks — called AFTER the built-in behavior. Assigning them
	// can never disable list→form fill, save or delete wiring.
	OnSelect  func(it view.Item)
	OnNew     func()
	OnSaved   func(err error)
	OnDeleted func(id string, err error)
	OnCancel  func()

	// internal
	form          *form.Form             // typed handle set by New; nil when standalone
	list          *targetlist.TargetList // owns the row rendering + ⋮ menu
	panel         *rightpanel.RightPanel // the skeleton this controller fills
	confirmDelete *modaldialog.ModalDialog
	selected      *SignalString
	search        *SignalString
	canDelete     *SignalBool
	deleteID      *SignalString // record pending confirmation (⋮ → Eliminar)
	deleteLabel   *SignalString // its label, for the confirmation message
	composing     *SignalBool   // "+" was pressed and nothing saved/cancelled yet (see active())
}

// active reports whether the toggle button should show "↺" (cancel/undo):
// either an existing row is selected, or the user is mid-composing a new one
// (pressed "+", hasn't saved or cancelled). Nothing else should read
// v.selected/v.composing directly to decide the icon/branch — this is the
// single place that combines them.
func (v *CrudView) active() bool {
	return v.selected.Get() != "" || v.composing.Get()
}

func (v *CrudView) Init(ctx Ctx) {
	v.selected = NewString("")
	v.search = NewString("")
	v.canDelete = NewBool(false)
	v.deleteID = NewString("")
	v.deleteLabel = NewString("")
	v.composing = NewBool(false)

	// The list is a targetlist component: it owns row rendering + the ⋮ menu and
	// shares the selected signal so its highlight follows the form.
	v.list = &targetlist.TargetList{
		Selected: v.selected,
		OnSelect: func(it targetlist.Item) {
			v.selectAction(view.Item{ID: it.ID, Label: it.Label, Description: it.Description})
		},
		OnEdit:   func(id string) { v.editAction(id) },
		OnDelete: func(id string) { v.deleteRequest(id) },
	}

	// The filter control reports terms; crudview owns what a term means.
	// Assigned here, not in Render, so it survives a re-render and so a host
	// that never renders — the conformance driver — still filters.
	if src, ok := v.Filter.(widget.Filterable); ok {
		src.OnFilterChange(func(term string) {
			v.search.Set(term)
			v.filter()
		})
	}

	// Delete confirmation modal — Content is built once; its message reacts to
	// v.deleteLabel so the same instance is reused across every ⋮ → Eliminar.
	// No "×": the body carries Cancelar and Eliminar, and a third way out says
	// nothing Cancelar does not. A destructive confirmation wants exactly two
	// exits, both of them explicit.
	v.confirmDelete = &modaldialog.ModalDialog{
		Title:     "Confirmar",
		HideClose: true,
		Content:   v.renderDeleteConfirm(),
	}

	if v.Presenter != nil {
		if err := v.Reload(); err != nil {
			Log(err.Error())
		}
	}
}

// renderDeleteConfirm builds the confirmation modal's body: a message naming
// the record, plus Cancelar/Eliminar actions. Built once in Init and reused —
// see the confirmDelete field.
func (v *CrudView) renderDeleteConfirm() *Element {
	// The question names the action and the record. "¿Desea continuar?" would
	// read as safe and be signed without thought — what makes someone stop is
	// seeing the verb and the name of the thing it is about to happen to.
	msg := P().BindTextFunc(func() string {
		return "¿Eliminar «" + v.deleteLabel.Get() + "»? Esta acción no se puede deshacer."
	})

	cancel := Button().Set(clsDelConfirmBtn.AsAttr()).Text("Cancelar").
		On("click", func(Event) { v.confirmDelete.Close() })

	confirm := Button().Set(clsDelConfirmBtn.AsAttr(), clsDelConfirmBtnDanger.AsAttr()).Text("Eliminar").
		On("click", func(Event) { v.confirmDeleteAction() })

	actions := Div().Set(clsDelConfirmActions.AsAttr()).Child(cancel, confirm)

	// Classed so the gap between the message and the actions is declared, and
	// declared as the same step the dialog puts between its header and its
	// body — otherwise the question sits far from the title and right on top of
	// the buttons.
	return Div().Set(clsDelConfirmBody.AsAttr()).Child(msg, actions)
}

// editAction (⋮ → Editar): load the record and unlock the form for editing.
// selectAction already loaded+locked it (read-only) if this wasn't already the
// selected row. Focuses the first field so the user can start typing right
// away — standard behavior, see the view/conformance "edit_focuses_first_field"
// clause.
func (v *CrudView) editAction(id string) {
	v.selectAction(view.Item{ID: id})
	if v.form != nil {
		v.form.SetLocked(false)
		v.form.Focus()
	}
}

// deleteRequest (⋮ → Eliminar): opens the confirmation modal instead of
// deleting immediately. confirmDeleteAction performs the actual delete.
func (v *CrudView) deleteRequest(id string) {
	if _, ok := v.Presenter.(view.Deleter); !ok || id == "" {
		return
	}
	label := id
	for _, it := range v.list.Items() {
		if it.ID == id {
			label = it.Label
			break
		}
	}
	v.deleteID.Set(id)
	v.deleteLabel.Set(label)
	v.confirmDelete.Open()
}

// confirmDeleteAction: the modal's "Eliminar" button. Deletes the record
// pending confirmation (set by deleteRequest) and closes the modal.
func (v *CrudView) confirmDeleteAction() {
	id := v.deleteID.Get()
	if deleter, ok := v.Presenter.(view.Deleter); ok && id != "" {
		v.deleteAction(deleter, id)
	}
	v.confirmDelete.Close()
}

func (v *CrudView) Reload() error {
	if v.Presenter == nil {
		return nil
	}
	if err := v.Presenter.Reload(); err != nil {
		return err
	}
	v.filter()
	return nil
}

// selectAction: card click / driver Select. Selecting an existing row shows it
// read-only — only the ⋮ → Editar path (editAction) unlocks it. On mobile
// (horizontal scroll-snap strip) it also snaps the viewport to the form panel;
// a no-op on desktop, where both columns are already visible.
func (v *CrudView) selectAction(it view.Item) {
	v.selected.Set(it.ID)
	v.canDelete.Set(it.ID != "")
	if it.ID != "" {
		v.composing.Set(false) // picking an existing row abandons any new-record draft
	}
	rec := v.Presenter.Select(it.ID)
	if v.form != nil {
		_ = v.form.LoadValues(rec) // nil record → LoadValues resets; not an error
		v.form.SetLocked(it.ID != "")
	}
	if it.ID != "" {
		if v.panel != nil {
			v.panel.ShowMain()
		}
	}
	if v.OnSelect != nil {
		v.OnSelect(it)
	}
}

// newAction: the toggle button in its "+" state (nothing selected). Idempotent —
// safe to call even when already in the "new" state. Focuses the first field —
// standard behavior, see the view/conformance "new_focuses_first_field" clause.
// Marks composing=true: the button switches to "↺" for the rest of the draft
// (until saved-and-cancelled or explicitly undone), the same as selecting an
// existing row does — see active().
func (v *CrudView) newAction() {
	v.selected.Set("")
	v.canDelete.Set(false)
	v.Presenter.Deselect()
	if v.form != nil {
		v.form.Reset()
		v.form.SetLocked(false)
		v.form.Focus()
	}
	v.composing.Set(true)
	// It focused the first field; on a phone that field is on the panel next
	// door, so bring the panel with it.
	if v.panel != nil {
		v.panel.ShowMain()
	}
	if v.OnNew != nil {
		v.OnNew()
	}
}

// undoAction: the toggle button in its "↺" state (active() — a row is
// selected, or a new-record draft is in progress). Undoes everything —
// deselects, clears the form, drops the draft — and returns the button to
// "+". Deliberately does NOT call Form.Focus(): cancelling must leave nothing
// selected/focused (standard behavior, see the view/conformance
// "cancel_clears_focus" clause) — unlike newAction/editAction, which enter an
// editable state and focus the first field on purpose.
func (v *CrudView) undoAction() {
	v.selected.Set("")
	v.canDelete.Set(false)
	v.composing.Set(false)
	v.Presenter.Deselect()
	if v.form != nil {
		v.form.Reset() // also clears the tracked FocusedFieldID()
		v.form.SetLocked(false)
	}
	// The list is the resting view: cancelling has to put the user back on it,
	// or on a phone the strip stays parked on an empty form with nothing left
	// to cancel.
	if v.panel != nil {
		v.panel.ShowAside()
	}
	if v.OnCancel != nil {
		v.OnCancel()
	}
}

// toggleAction: the single crud button's click handler. "↺"→undo whenever
// active() (a row is selected, or a new-record draft is in progress);
// "+"→create otherwise.
func (v *CrudView) toggleAction() {
	if v.active() {
		v.undoAction()
	} else {
		v.newAction()
	}
}

// saveAction persists the current form values. Only reachable when saver != nil.
// A silent no-op when the form isn't dirty (see Form.IsDirty) — moving focus
// through a field without changing it must never persist or fire OnSaved
// (a host's "Guardado" toast on an untouched field would be pure noise, and
// on mobile — where the module fills the screen — actively in the way).
func (v *CrudView) saveAction(saver view.Saver) {
	if !v.form.IsDirty() {
		return
	}
	err := v.form.Validate()
	if err == nil {
		record := v.Presenter.Record()
		if err = v.form.SyncValues(record); err == nil {
			if err = saver.Save(record); err == nil {
				v.form.MarkPristine() // a later untouched commit isn't dirty again
				if v.composing.Get() {
					// A new-record draft just saved successfully: the draft is
					// done, return to the "+" ready state. Not undoAction — that
					// fires OnCancel ("Cancelado"), wrong after a real save; this
					// is a silent reset, same shape as undoAction minus the
					// callback. Only for composing: editing an EXISTING record
					// (selected≠"", composing=false) must NOT reset here, or
					// every auto-save on blur would kick the user out of the
					// record they're still editing.
					v.composing.Set(false)
					v.selected.Set("")
					v.canDelete.Set(false)
					v.Presenter.Deselect()
					v.form.Reset()
					v.form.SetLocked(false)
				}
				_ = v.Reload()
			}
		}
	}
	if err != nil {
		Log(err.Error())
	}
	if v.OnSaved != nil {
		v.OnSaved(err)
	}
}

// autoSaveAction is the Form.OnFieldChange hook: every field commit (blur/change)
// persists immediately — there is no explicit Save button. A no-op when the
// presenter cannot save (standalone/read-only views).
func (v *CrudView) autoSaveAction() {
	if saver, ok := v.Presenter.(view.Saver); ok && v.form != nil {
		v.saveAction(saver)
	}
}

// deleteAction: delete button / driver Delete. Only reachable when deleter != nil.
func (v *CrudView) deleteAction(deleter view.Deleter, id string) {
	err := deleter.Delete(id)
	if err == nil {
		v.selected.Set("")
		v.canDelete.Set(false)
		_ = v.Reload()
	} else {
		Log(err.Error())
	}
	if v.OnDeleted != nil {
		v.OnDeleted(id, err)
	}
}

func (v *CrudView) filter() {
	src := v.Presenter.Filter(v.search.Get())
	items := make([]targetlist.Item, 0, len(src))
	for _, it := range src {
		items = append(items, targetlist.Item{ID: it.ID, Label: it.Label, Description: it.Description})
	}
	v.list.SetItems(items)
}

func (v *CrudView) Render() *Element {
	hasSource := v.Presenter != nil

	// The form's inset. The card around it is rightpanel's own `rp__article`
	// (the skeleton's Part("article")) — crudview paints no frame, so the form
	// area is exactly one layer, not a card nested in a card.
	boxContent := Div().Set(clsBoxContent.AsAttr())
	if v.Form != nil {
		boxContent.Child(v.Form)
	}

	v.panel = &rightpanel.RightPanel{
		Title:   v.Title,
		Article: boxContent,
	}

	if hasSource {
		// The list — its own inset card inside the aside's content band.
		v.panel.Aside = Div().Set(clsListaBox.AsAttr()).Child(v.list)

		// The filter is the consumer's control. crudview supplies no card
		// around it: rightpanel's controls band already keeps its size, and a
		// second frame around a control that has one reads as a box in a box.
		v.panel.AsideControls = v.Filter

		// Single toggle button — "+" when nothing is selected, "↺" when a row
		// is; Editar/Eliminar live in the targetlist row's ⋮ menu instead.
		toggle := Button().Set(clsBtnCrud.AsAttr()).
			// NOT "btn_..." — actionbutton's global `button[name*="btn"]` rule
			// matches any button whose name contains that substring and, being
			// a type+attribute selector, outranks this class; it was silently
			// injecting a stray margin. This button is crudview-owned, so its
			// name must not accidentally opt back in.
			Attr("name", "cv-crudtoggle").
			BindStateFunc(widget.Open, v.active).
			Child(
				iconCrudNew.Render(string(NameCrudView.Class("action-new"))).
					BindStateFunc(widget.Open, v.active),
				iconCrudCancel.Render(string(NameCrudView.Class("action-cancel"))).
					BindStateFunc(widget.Open, v.active),
			)
		toggle.On("click", func(Event) { v.toggleAction() })
		v.panel.AsideFooter = toggle
	}

	root := v.panel.Render()

	if hasSource {
		// v.confirmDelete's Show() wraps its content in a bare, class-less div
		// even while hidden. As an un-placed child of the frame's grid it got
		// auto-placed into an implicit second row and doubled the bottom gutter.
		// clsDelConfirmMount is position:fixed, which removes it from grid item
		// participation entirely, regardless of visibility.
		root.Child(Div().Set(clsDelConfirmMount.AsAttr()).Child(v.confirmDelete))
	}

	return root
}
