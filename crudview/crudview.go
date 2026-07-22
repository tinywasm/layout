package crudview

import (
	. "github.com/tinywasm/css"
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/html"

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
	clsBtnCrud                Class = "cv-btn-crud"
	clsBtnCrudIconHidden      Class = "cv-btn-crud-icon-hidden"
	clsAsideList              Class = "cv-aside-list"
	clsListaBox               Class = "cv-lista-box"
	clsAsideSearch            Class = "cv-aside-search"
	clsIcon16                 Class = "cv-icon-16"
)

const (
	// The single toggle button swaps between these two icons reactively — see
	// the "toggle" block in Render().
	iconCrudNew              = svg.Icon("icon-crud-new")    // "+"  — nothing selected
	iconCrudCancel           = svg.Icon("icon-crud-cancel") // "↺" — a row is selected (undo)
	iconCrudSearch           = svg.Icon("icon-crud-search")
	defaultSearchPlaceholder = "Search…"
)

type CrudView struct {
	Element // value embed — NEVER *dom.Element

	Title             string
	Form              Component      // what Render paints (may stay nil in standalone mode)
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
	form      *form.Form             // typed handle set by New; nil when standalone
	list      *targetlist.TargetList // owns the row rendering + ⋮ menu
	selected  *SignalString
	search    *SignalString
	canDelete *SignalBool
}

func (v *CrudView) Init(ctx Ctx) {
	v.selected = NewString("")
	v.search = NewString("")
	v.canDelete = NewBool(false)

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

	if v.Presenter != nil {
		if err := v.Reload(); err != nil {
			Log(err.Error())
		}
	}
}

// editAction (⋮ → Editar): load the record and unlock the form for editing.
// selectAction already loaded+locked it (read-only) if this wasn't already the
// selected row.
func (v *CrudView) editAction(id string) {
	v.selectAction(view.Item{ID: id})
	if v.form != nil {
		v.form.SetLocked(false)
	}
}

// deleteRequest (⋮ → Eliminar): delete the record. Modal confirmation is a
// follow-up step; this performs the delete when the presenter supports it.
func (v *CrudView) deleteRequest(id string) {
	if deleter, ok := v.Presenter.(view.Deleter); ok && id != "" {
		v.deleteAction(deleter, id)
	}
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
// read-only — only the ⋮ → Editar path (editAction) unlocks it.
func (v *CrudView) selectAction(it view.Item) {
	v.selected.Set(it.ID)
	v.canDelete.Set(it.ID != "")
	rec := v.Presenter.Select(it.ID)
	if v.form != nil {
		_ = v.form.LoadValues(rec) // nil record → LoadValues resets; not an error
		v.form.SetLocked(it.ID != "")
	}
	if v.OnSelect != nil {
		v.OnSelect(it)
	}
}

// newAction: the toggle button in its "+" state (nothing selected). Idempotent —
// safe to call even when already in the "new" state.
func (v *CrudView) newAction() {
	v.selected.Set("")
	v.canDelete.Set(false)
	v.Presenter.Deselect()
	if v.form != nil {
		v.form.Reset()
		v.form.SetLocked(false)
	}
	if v.OnNew != nil {
		v.OnNew()
	}
}

// undoAction: the toggle button in its "↺" state (a row is selected). Undoes
// everything — deselects, clears the form — and returns the button to "+".
func (v *CrudView) undoAction() {
	v.selected.Set("")
	v.canDelete.Set(false)
	v.Presenter.Deselect()
	if v.form != nil {
		v.form.Reset()
		v.form.SetLocked(false)
	}
	if v.OnCancel != nil {
		v.OnCancel()
	}
}

// toggleAction: the single crud button's click handler. "+"→create nothing
// selected; "↺"→undo when a row is selected.
func (v *CrudView) toggleAction() {
	if v.selected.Get() != "" {
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

	articleCont := Div().Set(articleContCls.AsAttr())

	// Title
	articleCont.Child(Div().Set(clsTitleContainer.AsAttr()).
		Child(Div().Set(clsTitle.AsAttr()).
			Child(H1().Text(v.Title))))

	// Article/Form
	boxContent := Div().Set(clsBoxContent.AsAttr())
	if v.Form != nil {
		boxContent.Child(v.Form)
	}
	articleCont.Child(Article().Child(boxContent))

	root.Child(articleCont)

	// ── Right Column ─────────────────────────────────────────────────────────
	if hasSource {
		asideCont := Aside().Set(clsAsideContend.AsAttr())

		// Search (TOP) — filters the targetlist below.
		asideSearch := Div().Set(clsAsideSearch.AsAttr())
		asideSearch.Child(Label().Child(renderIcon(iconCrudSearch)))

		placeholder := v.SearchPlaceholder
		if placeholder == "" {
			placeholder = defaultSearchPlaceholder
		}
		input := Input("search").Attr("placeholder", placeholder)
		input.On("input", func(e Event) {
			v.search.Set(e.TargetValue())
			v.filter()
		})
		asideSearch.Child(input)
		asideCont.Child(asideSearch)

		// List — the targetlist component owns rows + the ⋮ menu.
		asideList := Div().Set(clsAsideList.AsAttr()).
			Child(Div().Set(clsListaBox.AsAttr()).Child(v.list))
		asideCont.Child(asideList)

		// Single toggle button (BOTTOM) — "+" when nothing is selected, "↺" when a
		// row is; Editar/Eliminar live in the targetlist row's ⋮ menu instead.
		actions := Div().Set(clsAsideActions.AsAttr())
		toggle := Button().Set(clsBtnCrud.AsAttr()).
			Attr("name", "btn_crudtoggle").
			Child(
				iconCrudNew.Render(string(clsIcon16)).
					BindClass(string(clsBtnCrudIconHidden), DeriveBool(func() bool {
						return v.selected.Get() != ""
					})),
				iconCrudCancel.Render(string(clsIcon16)).
					BindClass(string(clsBtnCrudIconHidden), DeriveBool(func() bool {
						return v.selected.Get() == ""
					})),
			)
		toggle.On("click", func(Event) { v.toggleAction() })
		actions.Child(toggle)
		asideCont.Child(actions)

		root.Child(asideCont)
	}

	return root
}

func renderIcon(icon svg.Icon) *Element {
	return icon.Render(string(clsIcon16))
}
