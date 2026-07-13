package crudview

import (
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	. "github.com/tinywasm/html"
	"github.com/tinywasm/svg"
	"github.com/tinywasm/model"
	"github.com/tinywasm/router"
	. "github.com/tinywasm/css"
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
	iconCrudNew    = svg.Icon("icon-crud-new")
	iconCrudDel    = svg.Icon("icon-crud-del")
	iconCrudCancel = svg.Icon("icon-crud-cancel")
	iconCrudSave   = svg.Icon("icon-crud-save")
	iconCrudSearch = svg.Icon("icon-crud-search")
)

// Item is one record in the right-hand list.
type Item struct {
	ID          string // selection key
	Label       string // main text of the card
	Description string // small chip at the card's bottom-right (e.g. an IP)
}

// Source is the data seam: a fake in tests, a router.Caller adapter in prod.
type Source struct {
	Caller router.Caller
	ListOp string                 // logical operation, e.g. "list_devices"
	Args   func() model.Encodable // list request args (nil → no args)
	Decode func(raw []byte) ([]Item, error)
}

type CrudView struct {
	Element // value embed — NEVER *dom.Element

	Title  string        // h1, top-left corner of the panel (white on accent bar)
	Form   Component     // LEFT slot, the protagonist (typically *form.Form)
	Source Source        // feeds the right-hand list; zero Source = full-page
	                     // variant: no list/search/CRUD bar (title + form only)

	// Interaction callbacks — the composition root wires these to the form
	// and transport ONCE per app; modules never re-implement them.
	OnSelect func(it Item)                       // list card clicked → load into form
	OnNew    func()                              // (+) pressed → reset form for create
	OnSave   func(done func(err error))          // (💾) pressed; done(nil) reloads list
	OnDelete func(id string, done func(err error)) // (−) pressed on selection
	OnCancel func()                              // (↺) pressed → undo current edit
	OnError  func(err error)                     // list load/decode failures; nil = drop

	// internal state
	items     *SignalNodes
	allItems  []Item
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

	if v.Source.Caller != nil {
		v.Reload()
	}
}

func (v *CrudView) Reload() {
	if v.Source.Caller == nil {
		return
	}

	var args model.Encodable
	if v.Source.Args != nil {
		args = v.Source.Args()
	}

	v.Source.Caller.Call(v.Source.ListOp, args, func(raw []byte, err error) {
		if err != nil {
			v.handleError(err)
			return
		}

		if v.Source.Decode == nil {
			return
		}

		items, err := v.Source.Decode(raw)
		if err != nil {
			v.handleError(err)
			return
		}

		v.allItems = items
		v.filter()
	})
}

func (v *CrudView) filter() {
	term := Convert(v.search.Get()).ToLower().String()
	nodes := make([]*Element, 0)

	for _, it := range v.allItems {
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

func (v *CrudView) handleError(err error) {
	if v.OnError != nil {
		v.OnError(err)
	}
}

func (v *CrudView) Render() *Element {
	hasSource := v.Source.Caller != nil

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
							v.Reload()
						} else {
							v.handleError(err)
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
						v.Reload()
					} else {
						v.handleError(err)
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

		input := Input("search").Attr("placeholder", "Buscar...")
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
