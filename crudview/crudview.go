package crudview

import (
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/html"
	"github.com/tinywasm/svg"
	. "github.com/tinywasm/css"
	"github.com/tinywasm/view"
	"github.com/tinywasm/form"
)

var (
	clsModuleContent          Class = "cv-module-content"
	clsArticleContend         Class = "cv-article-contend"
	clsArticleContendFullPage Class = "cv-article-contend-full-page"
	clsAsideContend           Class = "cv-aside-contend"
	clsTitleContainer         Class = "cv-title-container"
	clsTitle                  Class = "cv-title"
	clsBoxContent             Class = "cv-box-content"
	clsControls               Class = "cv-controls"
	clsContebuton             Class = "cv-contebuton"
	clsBtnCrud                Class = "cv-btn-crud"
	clsAsideList              Class = "cv-aside-list"
	clsListaBox               Class = "cv-lista-box"
	clsLista                  Class = "cv-lista"
	clsTargetLi               Class = "cv-target-li"
	clsTargetLiOn             Class = "cv-target-li-on"
	clsDescriptionTarget      Class = "cv-description-target"
	clsAsideSearch            Class = "cv-aside-search"
	clsIcon16                 Class = "cv-icon-16"
)

const (
	iconCrudNew              = svg.Icon("icon-crud-new")
	iconCrudDel              = svg.Icon("icon-crud-del")
	iconCrudCancel           = svg.Icon("icon-crud-cancel")
	iconCrudSave             = svg.Icon("icon-crud-save")
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
	form      *form.Form // typed handle set by New; nil when standalone
	items     *SignalNodes
	selected  *SignalString
	search    *SignalString
	canSave   *SignalBool
	canDelete *SignalBool
}

func (v *CrudView) Init(ctx Ctx) {
	v.items = NewNodes()
	v.selected = NewString("")
	v.search = NewString("")
	v.canSave = NewBool(true)   // enabled by default in Pa100T capture
	v.canDelete = NewBool(false)

	if v.Presenter != nil {
		if err := v.Reload(); err != nil {
			Log(err.Error())
		}
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

// selectAction: card click / driver Select
func (v *CrudView) selectAction(it view.Item) {
	v.selected.Set(it.ID)
	v.canDelete.Set(it.ID != "")
	rec := v.Presenter.Select(it.ID)
	if v.form != nil {
		_ = v.form.LoadValues(rec) // nil record → LoadValues resets; not an error
	}
	if v.OnSelect != nil {
		v.OnSelect(it)
	}
}

// newAction: "+" button
func (v *CrudView) newAction() {
	v.selected.Set("")
	v.canDelete.Set(false)
	v.Presenter.Deselect()
	if v.form != nil {
		v.form.Reset()
	}
	if v.OnNew != nil {
		v.OnNew()
	}
}

// cancelAction: "↺" button
func (v *CrudView) cancelAction() {
	if v.form != nil {
		v.form.Reset()
	}
	if v.OnCancel != nil {
		v.OnCancel()
	}
}

// saveAction: save button / driver Save. Only reachable when saver != nil.
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
	nodes := make([]*Element, 0)

	for _, it := range v.Presenter.Filter(v.search.Get()) {
		it := it
		id := it.ID
		card := Li().Set(clsTargetLi.AsAttr()).
			BindClass(string(clsTargetLiOn), DeriveBool(func() bool {
				return v.selected.Get() == id
			})).
			Text(it.Label).
			Child(Span().Set(clsDescriptionTarget.AsAttr()).Text(it.Description))

		card.On("click", func(Event) {
			v.selectAction(it)
		})

		nodes = append(nodes, card)
	}

	v.items.Set(nodes)
}

func (v *CrudView) Render() *Element {
	hasSource := v.Presenter != nil

	var saver view.Saver
	var deleter view.Deleter
	if hasSource {
		saver, _ = v.Presenter.(view.Saver)
		deleter, _ = v.Presenter.(view.Deleter)
	}

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

	// Controls
	if hasSource {
		controls := Div().Set(clsControls.AsAttr())
		contebuton := Div().Set(clsContebuton.AsAttr())

		// Delete (−)
		if deleter != nil {
			btn := Button().Set(clsBtnCrud.AsAttr()).
				Attr("name", "btn_cruddel").
				BindAttrBool("disabled", DeriveBool(func() bool { return !v.canDelete.Get() })).
				Child(renderIcon(iconCrudDel))
			btn.On("click", func(Event) {
				id := v.selected.Get()
				if id != "" {
					v.deleteAction(deleter, id)
				}
			})
			contebuton.Child(btn)
		}

		// Cancel (↺)
		btnCancel := Button().Set(clsBtnCrud.AsAttr()).
			Attr("name", "btn_crudcancel").
			Child(renderIcon(iconCrudCancel))
		btnCancel.On("click", func(Event) {
			v.cancelAction()
		})
		contebuton.Child(btnCancel)

		// New (+)
		btnNew := Button().Set(clsBtnCrud.AsAttr()).
			Attr("name", "btn_crudnew").
			Child(renderIcon(iconCrudNew))
		btnNew.On("click", func(Event) {
			v.newAction()
		})
		contebuton.Child(btnNew)

		// Save (💾)
		if saver != nil {
			btn := Button().Set(clsBtnCrud.AsAttr()).
				Attr("name", "btn_crudsave").
				BindAttrBool("disabled", DeriveBool(func() bool { return !v.canSave.Get() })).
				Child(renderIcon(iconCrudSave))
			btn.On("click", func(Event) {
				v.saveAction(saver)
			})
			contebuton.Child(btn)
		}

		controls.Child(contebuton)
		articleCont.Child(controls)
	}

	root.Child(articleCont)

	// ── Right Column ─────────────────────────────────────────────────────────
	if hasSource {
		asideCont := Aside().Set(clsAsideContend.AsAttr())

		// List
		asideList := Div().Set(clsAsideList.AsAttr()).
			Child(Div().Set(clsListaBox.AsAttr()).
				Child(Ul().Set(clsLista.AsAttr()).BindChildren(v.items)))
		asideCont.Child(asideList)

		// Search
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

		root.Child(asideCont)
	}

	return root
}

func renderIcon(icon svg.Icon) *Element {
	return icon.Render(string(clsIcon16))
}
