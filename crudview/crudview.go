package crudview

import (
	. "github.com/tinywasm/css"
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/html"

	"github.com/tinywasm/components/modaldialog"
	"github.com/tinywasm/components/targetlist"
	"github.com/tinywasm/form"
	"github.com/tinywasm/svg"
	"github.com/tinywasm/view"
)

var (
	clsModuleContent          Class = "cv-module-content"
	clsArticleContend         Class = "cv-article-contend"
	clsArticleContendFullPage Class = "cv-article-contend-full-page"
	clsAsideContend           Class = "cv-aside-contend"
	clsTitleContainer         Class = "cv-title-container"
	clsTitle                  Class = "cv-title"
	clsBoxContent             Class = "cv-box-content"
	clsAsideActions           Class = "cv-aside-actions"
	clsAsideWrap              Class = "cv-aside-wrap"
	clsBtnCrud                Class = "cv-btn-crud"
	clsBtnCrudIconHidden      Class = "cv-btn-crud-icon-hidden"
	clsListaBox               Class = "cv-lista-box"
	clsAsideSearch            Class = "cv-aside-search"
	clsIcon16                 Class = "cv-icon-16"
	clsDelConfirmActions      Class = "cv-delconfirm-actions"
	clsDelConfirmBtn          Class = "cv-delconfirm-btn"
	clsDelConfirmBtnDanger    Class = "cv-delconfirm-btn-danger"
	clsDelConfirmMount        Class = "cv-delconfirm-mount"
	clsBackBtn                Class = "cv-back"
)

const (
	// The single toggle button swaps between these two icons reactively — see
	// the "toggle" block in Render().
	iconCrudNew              = svg.Icon("icon-crud-new")    // "+"  — nothing selected
	iconCrudCancel           = svg.Icon("icon-crud-cancel") // "↺" — a row is selected (undo)
	iconCrudSearch           = svg.Icon("icon-crud-search")
	iconCrudBack             = svg.Icon("icon-crud-back") // "‹" — mobile-only back-to-list button
	defaultSearchPlaceholder = "Search…"
)

type CrudView struct {
	Element // value embed — NEVER *dom.Element

	Title             string
	Form              Component // what Render paints (may stay nil in standalone mode)
	Presenter         view.Presenter
	SearchPlaceholder string

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

	// Delete confirmation modal — Content is built once; its message reacts to
	// v.deleteLabel so the same instance is reused across every ⋮ → Eliminar.
	v.confirmDelete = &modaldialog.ModalDialog{
		Title:   "Eliminar",
		Content: v.renderDeleteConfirm(),
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
	msg := P().BindTextFunc(func() string {
		return "¿Eliminar «" + v.deleteLabel.Get() + "»? Esta acción no se puede deshacer."
	})

	cancel := Button().Set(clsDelConfirmBtn.AsAttr()).Text("Cancelar").
		On("click", func(Event) { v.confirmDelete.Close() })

	confirm := Button().Set(clsDelConfirmBtn.AsAttr(), clsDelConfirmBtnDanger.AsAttr()).Text("Eliminar").
		On("click", func(Event) { v.confirmDeleteAction() })

	actions := Div().Set(clsDelConfirmActions.AsAttr()).Child(cancel, confirm)

	return Div().Child(msg, actions)
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

// detailPanelID identifies the form/article panel — the mobile scroll-snap
// target (see css.go's "(max-width: 640px)" block and selectAction below).
func (v *CrudView) detailPanelID() string {
	return v.GetID() + ".detail"
}

// listPanelID identifies the list/aside panel — the mobile "‹ back" button's
// scroll target (see css.go's "(max-width: 640px)" block and the back
// button's click handler in Render() below).
func (v *CrudView) listPanelID() string {
	return v.GetID() + ".list"
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
		if el, ok := Get(v.detailPanelID()); ok {
			el.ScrollIntoView()
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
func (v *CrudView) saveAction(saver view.Saver) {
	err := v.form.Validate()
	if err == nil {
		record := v.Presenter.Record()
		if err = v.form.SyncValues(record); err == nil {
			if err = saver.Save(record); err == nil {
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

	root := Div().Set(clsModuleContent.AsAttr())

	// ── Left Column ──────────────────────────────────────────────────────────
	articleContCls := clsArticleContend
	if !hasSource {
		articleContCls = clsArticleContendFullPage
	}

	articleCont := Div().Set(articleContCls.AsAttr()).ID(v.detailPanelID())

	// Title container. "‹ back" (mobile-only, see clsBackBtn's Display(None)
	// base rule) comes first, only when there's a list to go back to
	// (hasSource — the full-page/no-presenter variant has no list panel at
	// all, so a back button there would be a dead click). It only moves the
	// viewport back to the list panel, same effect as a manual swipe — it
	// does NOT call undoAction, so the selection/draft survives a user just
	// glancing back at the list.
	titleContainer := Div().Set(clsTitleContainer.AsAttr())
	if hasSource {
		back := Button().Set(clsBackBtn.AsAttr()).
			Attr("name", "cv-back"). // NOT "btn_..." — see cv-crudtoggle's comment below on the actionbutton name collision
			Child(iconCrudBack.Render(string(clsIcon16)))
		back.On("click", func(Event) {
			if el, ok := Get(v.listPanelID()); ok {
				el.ScrollIntoView()
			}
		})
		titleContainer.Child(back)
	}
	titleContainer.Child(Div().Set(clsTitle.AsAttr()).
		Child(H1().Text(v.Title)))
	articleCont.Child(titleContainer)

	// Article/Form
	boxContent := Div().Set(clsBoxContent.AsAttr())
	if v.Form != nil {
		boxContent.Child(v.Form)
	}
	articleCont.Child(Article().Child(boxContent))

	root.Child(articleCont)

	// ── Right Column ─────────────────────────────────────────────────────────
	// Three independent cards stacked in clsAsideWrap, each its own white
	// "frame" on the blue module background (never seamed together) — search
	// and the toggle button share the same card treatment AND height, list
	// is the big card between them. Matches the reference: its search bar
	// (bottom, in that layout) and its list are each their own bordered
	// piece; here the toggle button takes over the search bar's exact
	// styling since it now occupies that same bottom slot.
	if hasSource {
		// Search — its own small card up top, filtering the list below. The
		// label+input pill sits directly in the card (no extra inner layer —
		// see the RenderCSS comment on clsAsideSearch).
		searchCard := Div().Set(clsAsideSearch.AsAttr())
		searchCard.Child(Label().Child(renderIcon(iconCrudSearch)))

		placeholder := v.SearchPlaceholder
		if placeholder == "" {
			placeholder = defaultSearchPlaceholder
		}
		input := Input("search").Attr("placeholder", placeholder)
		input.On("input", func(e Event) {
			v.search.Set(e.TargetValue())
			v.filter()
		})
		searchCard.Child(input)

		// List — its own big card. The targetlist component owns rows + the
		// ⋮ menu; clsListaBox is the gray inset (unchanged).
		listCard := Aside().Set(clsAsideContend.AsAttr()).
			Child(Div().Set(clsListaBox.AsAttr()).Child(v.list))

		// Single toggle button — "+" when nothing is selected, "↺" when a row
		// is; Editar/Eliminar live in the targetlist row's ⋮ menu instead.
		// Its own card (clsAsideActions), same white-frame treatment and
		// height as searchCard — it occupies the bottom slot the search bar
		// has in the reference, so it gets that same integration, not a bare
		// color block that gets lost against the blue background.
		actionsCard := Div().Set(clsAsideActions.AsAttr())
		toggle := Button().Set(clsBtnCrud.AsAttr()).
			// NOT "btn_..." — actionbutton's global `button[name*="btn"]` rule
			// matches any button whose name contains that substring and, being
			// a type+attribute selector, outranks .cv-btn-crud's specificity;
			// it was silently injecting a stray margin that shrank this button
			// below clsAsideSearch's height. This button is crudview-owned
			// (ROADMAP.md), so its name must not accidentally opt back in.
			Attr("name", "cv-crudtoggle").
			Child(
				iconCrudNew.Render(string(clsIcon16)).
					BindClass(string(clsBtnCrudIconHidden), DeriveBool(func() bool {
						return v.active()
					})),
				iconCrudCancel.Render(string(clsIcon16)).
					BindClass(string(clsBtnCrudIconHidden), DeriveBool(func() bool {
						return !v.active()
					})),
			)
		toggle.On("click", func(Event) { v.toggleAction() })
		actionsCard.Child(toggle)

		asideWrap := Div().Set(clsAsideWrap.AsAttr()).ID(v.listPanelID()).Child(searchCard, listCard, actionsCard)
		root.Child(asideWrap)
		// v.confirmDelete's Show() wraps its content in a bare, class-less div
		// even while hidden (its dom/Show() anchor). As a 3rd, un-placed child
		// of this 2-column/1-row grid, that anchor got auto-placed into an
		// implicit 2nd row — and the shared row/column `gap` then added an
		// extra gap below row 1, doubling the bottom gutter versus the other
		// 3 sides. clsDelConfirmMount is `position: fixed`, which removes it
		// from grid item participation entirely (same as how the dialog
		// itself already positions when open), regardless of visibility.
		root.Child(Div().Set(clsDelConfirmMount.AsAttr()).Child(v.confirmDelete))
	}

	return root
}

func renderIcon(icon svg.Icon) *Element {
	return icon.Render(string(clsIcon16))
}
