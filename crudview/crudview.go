package crudview

import (
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	. "github.com/tinywasm/html"
	"github.com/tinywasm/svg"
	. "github.com/tinywasm/css"
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
	Form              Component
	Presenter         view.Presenter

	OnSelect          func(it view.Item)
	OnNew             func()
	OnSave            func(done func(err error))
	OnDelete          func(id string, done func(err error))
	OnCancel          func()

	SearchPlaceholder string

	// internal state
	items             *SignalNodes
	selected          *SignalString
	search            *SignalString
	canSave           *SignalBool
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

func (v *CrudView) filter() {
	term := Convert(v.search.Get()).ToLower().String()
	nodes := make([]*Element, 0)

	for _, it := range v.Presenter.Items() {
		if term != "" {
			label := Convert(it.Label).ToLower().String()
			desc := Convert(it.Description).ToLower().String()
			if !Contains(label, term) && !Contains(desc, term) {
				continue
			}
		}

		it := it
		id := it.ID
		card := Li().Set(clsTargetLi.AsAttr()).
			BindClass(string(clsTargetLiOn), DeriveBool(func() bool {
				return v.selected.Get() == id
			})).
			Text(it.Label).
			Child(Span().Set(clsDescriptionTarget.AsAttr()).Text(it.Description))

		card.On("click", func(Event) {
			v.Select(id)
			if v.OnSelect != nil {
				v.OnSelect(it)
			}
		})

		nodes = append(nodes, card)
	}

	v.items.Set(nodes)
}

func (v *CrudView) Select(id string) {
	v.selected.Set(id)
	v.canDelete.Set(id != "")
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

	// Controls
	if hasSource {
		controls := Div().Set(clsControls.AsAttr())
		contebuton := Div().Set(clsContebuton.AsAttr())

		// Delete (−)
		if v.OnDelete != nil {
			btn := Button().Set(clsBtnCrud.AsAttr()).
				Attr("name", "btn_cruddel").
				BindAttrBool("disabled", DeriveBool(func() bool { return !v.canDelete.Get() })).
				Child(renderIcon(iconCrudDel))
			btn.On("click", func(Event) {
				id := v.selected.Get()
				if id != "" {
					v.OnDelete(id, func(err error) {
						if err == nil {
							v.Select("")
							_ = v.Reload()
						} else {
							Log(err.Error())
						}
					})
				}
			})
			contebuton.Child(btn)
		}

		// Cancel (↺)
		if v.OnCancel != nil {
			btn := Button().Set(clsBtnCrud.AsAttr()).
				Attr("name", "btn_crudcancel").
				Child(renderIcon(iconCrudCancel))
			btn.On("click", func(Event) {
				v.OnCancel()
			})
			contebuton.Child(btn)
		}

		// New (+)
		if v.OnNew != nil {
			btn := Button().Set(clsBtnCrud.AsAttr()).
				Attr("name", "btn_crudnew").
				Child(renderIcon(iconCrudNew))
			btn.On("click", func(Event) {
				v.Select("")
				v.OnNew()
			})
			contebuton.Child(btn)
		}

		// Save (💾)
		if v.OnSave != nil {
			btn := Button().Set(clsBtnCrud.AsAttr()).
				Attr("name", "btn_crudsave").
				BindAttrBool("disabled", DeriveBool(func() bool { return !v.canSave.Get() })).
				Child(renderIcon(iconCrudSave))
			btn.On("click", func(Event) {
				v.OnSave(func(err error) {
					if err == nil {
						_ = v.Reload()
					} else {
						Log(err.Error())
					}
				})
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
