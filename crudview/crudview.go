package crudview

import (
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/html"

	"github.com/tinywasm/components/modaldialog"
	"github.com/tinywasm/components/targetlist"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/fmt/lang"
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
	clsBtnCrudDelete       = NameCrudView.Class("action-delete")
	clsBtnCrudEdit         = NameCrudView.Class("action-edit")
	clsBtnCrudCount        = NameCrudView.Class("action-count")
	clsFooter              = NameCrudView.Class("footer")
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
	iconCrudDelete = svg.Icon("icon-crud-delete") // "🗑" — bulk delete
	iconCrudEdit   = svg.Icon("icon-crud-edit")   // "✏" — bulk edit
)

// crudMode is the single source of truth for which chrome the footer shows and
// what a tap on a row means. A closed set, in one signal: three booleans would
// make "menu open AND selecting" representable, and it is not a state — it is
// a bug waiting for a race between two clicks.
type crudMode string

const (
	modeNormal   crudMode = ""       // "+" plus the two bulk entries, 🗑 and ✏
	modeDeleting crudMode = "delete" // ↺ + 🗑 N, rows show checks
	modeEditing  crudMode = "edit"   // ↺ + ✏ N, rows show checks
)

// ListView is the row-rendering half of a CrudView. targetlist.TargetList
// (plain label rows) and targetdate.TargetDate (a leading date/time badge
// instead of a plain label — see view.Item's LeadTop/Main/Bottom) both
// satisfy it, so Config.List can swap one for the other without CrudView
// ever knowing which it got.
type ListView interface {
	Component
	SetItems(items []view.Item)
	Items() []view.Item
	Count() int
	// Selection mode — targetlist and targetdate both implement it by
	// assembling the components/listselect lego piece.
	SetSelectMode(on bool)
	CheckedIDs() []string
	OnCheckedChange(fn func(n int))
}

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

	// List builds the row-rendering widget, given the Selected signal and the
	// OnSelect/OnDelete callbacks CrudView owns — Init calls it once. Set by
	// New from Config.List; nil there resolves to a targetlist.TargetList
	// factory, same "ergonomic default, not a decision imposed" as Filter.
	List func(selected *SignalString, onSelect func(view.Item)) ListView

	// Additive user hooks — called AFTER the built-in behavior. Assigning them
	// can never disable list→form fill, save or delete wiring.
	OnSelect  func(it view.Item)
	OnNew     func()
	OnSaved   func(err error)
	OnDeleted func(id string, err error)
	OnCancel  func()

	// internal
	form          *form.Form             // typed handle set by New; nil when standalone
	list          ListView               // owns the row rendering + ⋮ menu
	panel         *rightpanel.RightPanel // the skeleton this controller fills
	confirmDelete *modaldialog.ModalDialog
	selected      *SignalString
	search        *SignalString
	canDelete     *SignalBool
	deleteID      *SignalString // record pending confirmation (⋮ → Eliminar)
	deleteLabel   *SignalString // its label, for the confirmation message
	composing     *SignalBool   // "+" was pressed and nothing saved/cancelled yet (see active())
	mode          *SignalString // single source of truth for mode (crudMode)
	checkedCount  *SignalString // count of checked items in selection mode
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
	v.mode = NewString(string(modeNormal))
	v.checkedCount = NewString("0")

	// The list owns row rendering + the ⋮ menu and shares the selected signal
	// so its highlight follows the form. New resolves Config.List's default
	// before this runs, but Init must not assume it always went through
	// New — direct struct construction (&CrudView{...}; v.Init(ctx)) is an
	// established pattern in this package's own tests, same as Filter being
	// left nil is fine. Default here mirrors New's own default exactly.
	buildList := v.List
	if buildList == nil {
		buildList = func(selected *SignalString, onSelect func(view.Item)) ListView {
			return &targetlist.TargetList{Selected: selected, OnSelect: onSelect}
		}
	}
	v.list = buildList(v.selected, v.selectAction)
	v.list.OnCheckedChange(func(n int) {
		v.checkedCount.Set(fmt.Sprintf("%d", n))
	})

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
		Title:     lang.Translate("Confirm").String(),
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
	// seeing the verb and the name of the thing it is about to happen to. Each
	// word is an independent dictionary key — joined with spaces at read time —
	// so consumers reuse the same entries ("Delete" is already the confirm
	// button's key).
	msg := P().BindTextFunc(func() string {
		tmpl := lang.Translate("Delete", "%s?", "This", "action", "cannot", "be", "undone.").String()
		return fmt.Sprintf(tmpl, "«"+v.deleteLabel.Get()+"»")
	})

	cancel := Button().Set(clsDelConfirmBtn.AsAttr()).Text(lang.Translate("Cancel").String()).
		On("click", func(Event) { v.confirmDelete.Close() })

	confirm := Button().Set(clsDelConfirmBtn.AsAttr(), clsDelConfirmBtnDanger.AsAttr()).Text(lang.Translate("Delete").String()).
		On("click", func(Event) { v.confirmDeleteAction() })

	actions := Div().Set(clsDelConfirmActions.AsAttr()).Child(cancel, confirm)

	// Classed so the gap between the message and the actions is declared, and
	// declared as the same step the dialog puts between its header and its
	// body — otherwise the question sits far from the title and right on top of
	// the buttons.
	return Div().Set(clsDelConfirmBody.AsAttr()).Child(msg, actions)
}

// editAction: load the record and focus the first field so the user can start
// typing right away. Kept as a method although no UI path calls it anymore —
// the ⋮ menu lost Editar with the lock it existed to undo — because the
// view/conformance suite's "edit_focuses_first_field" clause drives it
// directly (see conformance_test.go). Focus is synchronous, on purpose — see
// newAction below for why it must never be deferred.
func (v *CrudView) editAction(id string) {
	v.selectAction(view.Item{ID: id})
	if v.form != nil {
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

// confirmDeleteAction: the modal's "Eliminar" button. Deletes the record(s)
// pending confirmation and closes the modal.
func (v *CrudView) confirmDeleteAction() {
	if v.mode.Get() == string(modeDeleting) {
		if deleter, ok := v.Presenter.(view.Deleter); ok && v.list != nil {
			ids := v.list.CheckedIDs()
			if len(ids) > 0 {
				err := deleter.Delete(ids...)
				if err == nil {
					v.setMode(modeNormal)
					_ = v.Reload()
				} else {
					Log(err.Error())
				}
				if v.OnDeleted != nil {
					v.OnDeleted("", err)
				}
			}
		}
	} else {
		id := v.deleteID.Get()
		if deleter, ok := v.Presenter.(view.Deleter); ok && id != "" {
			v.deleteAction(deleter, id)
		}
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

// selectAction: card click / driver Select. Selecting an existing row loads it
// into the form — immediately editable, because the form lock is gone: there
// is no Save button, every field commit persists on its own (autoSaveAction),
// so nothing needs protecting from edits. On mobile (horizontal scroll-snap
// strip) it also snaps the viewport to the form panel; a no-op on desktop,
// where both columns are already visible.
func (v *CrudView) selectAction(it view.Item) {
	v.selected.Set(it.ID)
	v.canDelete.Set(it.ID != "")
	if v.list != nil {
	}
	if it.ID != "" {
		v.composing.Set(false) // picking an existing row abandons any new-record draft
	}
	rec := v.Presenter.Select(it.ID)
	if v.form != nil {
		_ = v.form.LoadValues(rec) // nil record → LoadValues resets; not an error
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
	if v.list != nil {
	}
	v.Presenter.Deselect()
	if v.form != nil {
		v.form.Reset()
		v.form.Focus()
	}
	v.composing.Set(true)
	// Focus before ShowMain, both synchronous, is deliberate: iOS Safari only
	// opens the keyboard for a focus() that happens inside the call stack of
	// the user gesture, so deferring it by any amount (tried at several
	// delays) moves DOM focus but leaves the keyboard shut. Order matters too
	// — Focus must not be the last scroll-affecting thing to run, which is why
	// ShowMain follows it and dom's Focus suppresses the browser's own
	// focus-scroll (see elementWasm.Focus): otherwise the two scrolls race and
	// the strip lands between snap points.
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
	v.clearSelection()
	if v.OnCancel != nil {
		v.OnCancel()
	}
}

// clearSelection puts the controller back in its resting state: nothing
// selected, no draft, no delete armed, an empty form, and — on a phone, where
// the two columns are a scroll-snap strip — the list back on screen instead of
// a form with nothing left in it.
//
// Two callers, deliberately: undoAction (the user pressed "↺") and
// dropSelectionOutOfScope (the filter moved and took the record with it). They
// are the same state change; only undoAction is a user-initiated cancel, so
// only undoAction fires OnCancel.
func (v *CrudView) setMode(m crudMode) {
	if crudMode(v.mode.Get()) == m {
		return
	}
	v.mode.Set(string(m))
	if m == modeDeleting || m == modeEditing {
		if v.list != nil {
			v.list.SetSelectMode(true)
		}
		if m == modeEditing && v.form != nil {
			v.form.Reset()
			v.form.MarkPristine()
		}
	} else {
		if v.list != nil {
			v.list.SetSelectMode(false)
		}
	}
}

func (v *CrudView) clearSelection() {
	v.setMode(modeNormal)
	v.selected.Set("")
	v.canDelete.Set(false)
	v.composing.Set(false)
	if v.list != nil {
	}
	if v.Presenter != nil {
		v.Presenter.Deselect()
	}
	if v.form != nil {
		v.form.Reset() // also clears the tracked FocusedFieldID()
	}
	if v.panel != nil {
		v.panel.ShowAside()
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
// presenter cannot save (standalone/read-only views) or when in bulk editing mode.
func (v *CrudView) autoSaveAction() {
	if v.mode != nil && v.mode.Get() == string(modeEditing) {
		return // Bulk edit mode: changes are queued until explicit bulk apply
	}
	if saver, ok := v.Presenter.(view.Saver); ok && v.form != nil {
		v.saveAction(saver)
	}
}

// deleteAction: delete button / driver Delete. Only reachable when deleter != nil.
func (v *CrudView) deleteAction(deleter view.Deleter, id string) {
	err := deleter.Delete(id)  // plural contract; the bulk path arrives with this repo's own plan
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

// filter repopulates the list for the current term and then enforces the one
// invariant a list-detail controller cannot let slip: what the form holds must
// be something the list still shows.
func (v *CrudView) filter() {
	v.list.SetItems(v.Presenter.Filter(v.search.Get()))
	v.dropSelectionOutOfScope()
}

// bulkDeleteAction commits the marked rows. One call, not a loop: view.Deleter
// is variadic precisely so the whole batch is one statement, and a loop would
// bring back the half-applied failure the plural contract exists to prevent.
func (v *CrudView) bulkDeleteAction() {
	if v.list == nil {
		return
	}
	ids := v.list.CheckedIDs()
	if len(ids) == 0 {
		return
	}
	var label string
	if len(ids) == 1 {
		label = ids[0]
		for _, it := range v.list.Items() {
			if it.ID == ids[0] {
				label = it.Label
				break
			}
		}
	} else {
		label = fmt.Sprintf("%d", len(ids))
	}
	v.deleteID.Set("")
	v.deleteLabel.Set(label)
	v.confirmDelete.Open()
}

// bulkEditAction patches the marked rows with ONLY the fields the user
// touched. Sending whole records instead would silently revert every column
// someone else changed since this client last reloaded — see the master plan's
// "por qué ids + delta".
func (v *CrudView) bulkEditAction() {
	if v.list == nil || v.form == nil {
		return
	}
	ids := v.list.CheckedIDs()
	if len(ids) == 0 {
		return
	}
	fields := v.form.DirtyFields()
	if len(fields) == 0 {
		return
	}
	updater, ok := v.Presenter.(view.Updater)
	if !ok {
		return
	}
	record := v.Presenter.Record()
	if err := v.form.SyncValues(record); err != nil {
		Log(err.Error())
		return
	}
	err := updater.Update(ids, record, fields)
	if err == nil {
		v.setMode(modeNormal)
		v.form.Reset()
		_ = v.Reload()
	} else {
		Log(err.Error())
	}
}

// dropSelectionOutOfScope clears the selection when the freshly filtered list
// no longer contains it. A picker that changes the SCOPE (a patient selector
// narrowing to that patient's records) leaves the previously loaded record
// orphaned: still in the form, still editable, still saveable — against a list
// that no longer shows it. That is a data-corruption path, not a cosmetic one.
//
// Unconditional, with no config flag to turn it off: a form holding a record
// outside the visible list is an illegal state, not a preference. A composing
// draft goes too — it belongs to the scope that just left.
func (v *CrudView) dropSelectionOutOfScope() {
	id := v.selected.Get()
	if id == "" && !v.composing.Get() {
		return // nothing loaded; nothing to orphan
	}
	if id != "" && v.list != nil {
		for _, it := range v.list.Items() {
			if it.ID == id {
				return // still in scope
			}
		}
	}
	v.clearSelection()
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

		toggle := Button().Set(clsBtnCrud.AsAttr()).
			Attr("name", "cv-crudtoggle").
			BindStateFunc(widget.Open, func() bool { return v.mode.Get() != string(modeNormal) || v.active() }).
			Child(
				iconCrudNew.Render(string(NameCrudView.Class("action-new"))).
					BindStateFunc(widget.Open, func() bool { return v.mode.Get() != string(modeNormal) || v.active() }),
				iconCrudCancel.Render(string(NameCrudView.Class("action-cancel"))).
					BindStateFunc(widget.Open, func() bool { return v.mode.Get() != string(modeNormal) || v.active() }),
			)
		toggle.On("click", func(Event) {
			if v.mode.Get() != string(modeNormal) {
				v.setMode(modeNormal)
			} else {
				v.toggleAction()
			}
		})

		footer := Div().Set(clsFooter.AsAttr()).Child(toggle)

		if _, ok := v.Presenter.(view.Deleter); ok {
			btnDelete := Button().Set(clsBtnCrudDelete.AsAttr()).
				Attr("name", "cv-cruddelete").
				BindStateFunc(widget.Open, func() bool {
					return v.mode.Get() == string(modeNormal) || v.mode.Get() == string(modeDeleting)
				}).
				BindAttrBool("disabled", DeriveBool(func() bool {
					if v.mode.Get() == string(modeNormal) {
						return v.active()
					}
					return v.checkedCount.Get() == "0" || v.checkedCount.Get() == ""
				})).
				Child(
					iconCrudDelete.Render(""),
					Span().Set(clsBtnCrudCount.AsAttr()).
						BindText(v.checkedCount).
						BindStateFunc(widget.Open, func() bool { return v.mode.Get() == string(modeDeleting) }),
				)
			btnDelete.On("click", func(Event) {
				if v.mode.Get() == string(modeNormal) {
					if !v.active() {
						v.setMode(modeDeleting)
					}
				} else if v.mode.Get() == string(modeDeleting) {
					v.bulkDeleteAction()
				}
			})
			footer.Child(btnDelete)
		}

		if _, ok := v.Presenter.(view.Updater); ok {
			btnEdit := Button().Set(clsBtnCrudEdit.AsAttr()).
				Attr("name", "cv-crudedit").
				BindStateFunc(widget.Open, func() bool {
					return v.mode.Get() == string(modeNormal) || v.mode.Get() == string(modeEditing)
				}).
				BindAttrBool("disabled", DeriveBool(func() bool {
					if v.mode.Get() == string(modeNormal) {
						return v.active()
					}
					return v.checkedCount.Get() == "0" || v.checkedCount.Get() == ""
				})).
				Child(
					iconCrudEdit.Render(""),
					Span().Set(clsBtnCrudCount.AsAttr()).
						BindText(v.checkedCount).
						BindStateFunc(widget.Open, func() bool { return v.mode.Get() == string(modeEditing) }),
				)
			btnEdit.On("click", func(Event) {
				if v.mode.Get() == string(modeNormal) {
					if !v.active() {
						v.setMode(modeEditing)
					}
				} else if v.mode.Get() == string(modeEditing) {
					v.bulkEditAction()
				}
			})
			footer.Child(btnEdit)
		}

		v.panel.AsideFooter = footer
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
