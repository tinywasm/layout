package crudview

import (
	. "github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
	. "github.com/tinywasm/html"
	"github.com/tinywasm/svg"
	"github.com/tinywasm/model"
)

// Item is one record in the right-hand list.
type Item struct {
	ID          string // selection key
	Label       string // main text of the card
	Description string // small chip at the card's bottom-right (e.g. an IP)
}

// Caller is a placeholder for tinywasm/router.Caller
type Caller interface {
	Call(op string, args model.Encodable) ([]byte, error)
}

// Source is the data seam: a fake in tests, a router.Caller adapter in prod.
type Source struct {
	Caller Caller
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

	raw, err := v.Source.Caller.Call(v.Source.ListOp, args)
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
}

func (v *CrudView) filter() {
	term := Convert(v.search.Get()).ToLower().String()
	nodes := make([]*Element, 0)

	for _, it := range v.allItems {
		if term != "" {
			label := Convert(it.Label).ToLower().String()
			desc := Convert(it.Description).ToLower().String()
			if !contains(label, term) && !contains(desc, term) {
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
				Child(renderIcon("icon-crud-del"))
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
				Child(renderIcon("icon-crud-cancel"))
			btn.On("click", func(Event) {
				v.OnCancel()
			})
			contebuton.Child(btn)
		}

		// New (+)
		if v.OnNew != nil {
			btn := Button().Set(clsBtnCrud.AsAttr()).
				Attr("name", "btn_crudnew").
				Child(renderIcon("icon-crud-new"))
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
				Child(renderIcon("icon-crud-save"))
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
		asideSearch.Child(Label().Child(renderIcon("icon-crud-search")))

		input := Input("search")
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

func contains(s, substr string) bool {
	if len(substr) == 0 { return true }
	if len(s) < len(substr) { return false }
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr { return true }
	}
	return false
}

func renderIcon(id string) *Element {
	return svg.Svg().
		Attr("aria-hidden", "true").
		Attr("focusable", "false").
		Class("icon-16").
		Child(svg.Use().Attr("href", "#"+id))
}
