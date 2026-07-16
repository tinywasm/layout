---
PLAN: "feat: crudview renders a view.Presenter — Config drops ListOp/SaveOp/Decode"
TAG: v0.0.12
---

> Este plan se despacha vía el flujo CodeJob. Ver skill: **agents-workflow**.
> Orquestado por `tinywasm/app-releases/docs/REUSABLE_MODULES_MASTER_PLAN.md` — **Fase B4**.

# PLAN — `layout/crudview`: renderer de `view.Presenter`

Autocontenido, en español. Eres un agente **sin contexto previo** y **solo tienes este repo**
(`tinywasm/layout`). Todo el contrato y el código exacto van inline.

> Nota: existió un `docs/PLAN.md` anterior, borrado en el commit que introdujo la `Config`/`New`
> actuales (motivo: "un plan cerrado se retira, su contenido pasa a docs permanentes"). El
> `README.md` todavía enlaza a ese archivo borrado (`README.md:11`) — corrígelo en este plan (§3.4).

## 1. Qué cambia y por qué

`tinywasm/view` (nuevo, ya publicado `v0.1.0`) es ahora el motor CRUD agnóstico de UI: un módulo
llama `view.New(caller, record, listOp, newList, project, opts...)` y obtiene un `view.Presenter`
—lista, selección, guardar, eliminar, **todo síncrono, todo ya resuelto contra el `Caller`**—. Hoy
`crudview.Config`/`New` **reimplementan exactamente eso** (`Source`, `OnSelect`/`OnSave`/`OnDelete`
armados a mano, `Caller.Call` invocado directo). Es la duplicación que esta ola cierra:
`crudview` deja de ser el motor y pasa a ser **solo el renderer**: recibe un `view.Presenter` ya
construido, genera el form desde `Record().Schema()`, sincroniza el input del usuario dentro de
`Record()`, y llama `Presenter.Save(record)`/`Delete(id)`. El módulo de dominio construye el
`Presenter` (importando `view`+`model`+`router`, nunca `layout`); el app decide qué renderer usar.

**Rotura adicional que ESTE plan también arregla:** `router.Caller.Call` cambió de firma
(`router@v0.1.13`, ya publicado) — los tres call-sites que `crudview` tiene hoy dejan de compilar.
Al reemplazar `Config`/`New` por el diseño de este plan, **esos tres call-sites desaparecen**: ya no
es `crudview` quien llama `Caller.Call`, es `view.Presenter` (dentro de otro repo, ya resuelto).

## 2. Estado actual exacto (verificado, no supuesto)

Dos archivos, mismo paquete `crudview`:

**`crud.go`** — `Config` (líneas 13-39) y `New` (41-103, cuerpo completo):
```go
type Config struct {
	ParentID string
	Caller router.Caller
	Title  string
	Record model.Model
	ListOp   string
	SaveOp   string
	DeleteOp string
	Args   func() model.Encodable
	Decode func(raw []byte) ([]Item, error)
	Fill   func(id string) model.Model
	OnError func(err error)
	SearchPlaceholder string
}

func New(cfg Config) (*CrudView, error) {
	if cfg.Caller == nil { return nil, fmt.Errf("crudview.New: Caller is required") }
	if cfg.ListOp == "" { return nil, fmt.Errf("crudview.New: ListOp is required") }
	if model.IsNil(cfg.Record) { return nil, fmt.Errf("crudview.New: Record is required") }

	f, err := form.New(cfg.ParentID, cfg.Record)
	if err != nil { return nil, err }
	f.HideSubmit()

	v := &CrudView{
		Title: cfg.Title, Form: f,
		Source: Source{Caller: cfg.Caller, ListOp: cfg.ListOp, Args: cfg.Args, Decode: cfg.Decode},
		OnError: cfg.OnError, SearchPlaceholder: cfg.SearchPlaceholder,
	}

	v.OnSelect = func(it Item) {
		if cfg.Fill == nil { return }
		_ = f.LoadValues(cfg.Fill(it.ID))
	}
	v.OnNew = func() { f.Reset() }
	v.OnCancel = func() { f.Reset() }

	if cfg.SaveOp != "" {
		v.OnSave = func(done func(err error)) {
			if err := f.Validate(); err != nil { done(err); return }
			if err := f.SyncValues(cfg.Record); err != nil { done(err); return }
			cfg.Caller.Call(cfg.SaveOp, cfg.Record, func(_ []byte, err error) { done(err) }) // ← call-site 2
		}
	}
	if cfg.DeleteOp != "" {
		v.OnDelete = func(id string, done func(err error)) {
			rec := cfg.Fill(id)
			if model.IsNil(rec) { done(fmt.Errf("crudview: no record for id %s", id)); return }
			cfg.Caller.Call(cfg.DeleteOp, rec, func(_ []byte, err error) { done(err) }) // ← call-site 3
		}
	}
	return v, nil
}
```

**`crudview.go`** — `CrudView` struct (58-84), `Source` struct (51-56), `Item` (44-48), `Init`
(86-96), `Reload` (98-126, contiene el call-site 1), `filter` (129-162), `Select`/`handleError`
(164-173). El resto (`Render`, botones CRUD) NO se muestra aquí completo por espacio — está intacto
y **no cambia su lógica de pintado**, solo su fuente de datos (§3).

```go
type Item struct { ID, Label, Description string } // idéntico campo a campo a view.Item

type Source struct {
	Caller router.Caller
	ListOp string
	Args   func() model.Encodable
	Decode func(raw []byte) ([]Item, error)
}

type CrudView struct {
	Element
	Title  string
	Form   Component
	Source Source
	OnSelect func(it Item)
	OnNew    func()
	OnSave   func(done func(err error))
	OnDelete func(id string, done func(err error))
	OnCancel func()
	OnError  func(err error)
	SearchPlaceholder string
	items     *SignalNodes
	allItems  []Item
	selected  *SignalString
	search    *SignalString
	canSave   *SignalBool
	canDelete *SignalBool
}

func (v *CrudView) Reload() {
	if v.Source.Caller == nil { return }
	var args model.Encodable
	if v.Source.Args != nil { args = v.Source.Args() }
	v.Source.Caller.Call(v.Source.ListOp, args, func(raw []byte, err error) { // ← call-site 1
		if err != nil { v.handleError(err); return }
		if v.Source.Decode == nil { return }
		items, err := v.Source.Decode(raw)
		if err != nil { v.handleError(err); return }
		v.allItems = items
		v.filter()
	})
}
```

- `go.mod` del módulo `layout` (compartido por `crudview`, no hay `go.mod` propio del subpaquete):
  `github.com/tinywasm/router v0.1.12`, `github.com/tinywasm/model v0.0.15`,
  `github.com/tinywasm/form v0.2.14`. **No** hay `github.com/tinywasm/view` — se añade en este plan.
- Los 4 archivos de test (`crudview_test.go`, `consumer_test.go`, `crudview_wasm_test.go`,
  `consumer_wasm_test.go`) construyen `Config{...}` con la forma vieja y un `fakeCaller` de 3
  argumentos — **todos migran** (§4).

## 3. El cambio exacto

### 3.1 `go.mod`

```
go get github.com/tinywasm/router@v0.1.13 github.com/tinywasm/view@latest
```

### 3.2 `crud.go` — `Config`/`New` nuevos (reemplazan los de §2 completos)

```go
package crudview

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/form"
	"github.com/tinywasm/view"
)

// Config is what a renderer needs to draw a view.Presenter — nothing about ops, transport, or
// codec: that is ALL inside the Presenter already. The module builds the Presenter via
// view.New(...) (importing view+model+router, never layout) and hands it here.
type Config struct {
	// ParentID is the DOM id the form mounts under.
	ParentID string
	// Presenter is built by the module via view.New(...). Required.
	Presenter view.Presenter
}

// New builds the renderer around an already-constructed Presenter. It generates the form from
// Presenter.Record().Schema(), and wires save/delete to sync the form INTO Record() before
// calling Presenter.Save/Delete — crudview never talks to a Caller directly anymore.
func New(cfg Config) (*CrudView, error) {
	if cfg.Presenter == nil {
		return nil, fmt.Errf("crudview.New: Presenter is required")
	}

	f, err := form.New(cfg.ParentID, cfg.Presenter.Record())
	if err != nil {
		return nil, err // a record with no widgets fails HERE, loudly
	}
	f.HideSubmit() // the CRUD bar owns save — the form must not paint its own submit

	v := &CrudView{
		Title:             cfg.Presenter.Title(),
		Form:              f,
		Presenter:         cfg.Presenter,
		SearchPlaceholder: cfg.Presenter.SearchPlaceholder(),
	}

	v.OnSelect = func(it view.Item) {
		_ = f.LoadValues(cfg.Presenter.Select(it.ID)) // nil record → LoadValues resets. Not an error.
	}
	v.OnNew = func() {
		cfg.Presenter.Select("")
		f.Reset()
	}
	v.OnCancel = func() { f.Reset() }

	if cfg.Presenter.CanSave() {
		v.OnSave = func(done func(err error)) {
			if err := f.Validate(); err != nil {
				done(err)
				return
			}
			record := cfg.Presenter.Record()
			if err := f.SyncValues(record); err != nil {
				done(err)
				return
			}
			done(cfg.Presenter.Save(record))
		}
	}

	if cfg.Presenter.CanDelete() {
		v.OnDelete = func(id string, done func(err error)) {
			done(cfg.Presenter.Delete(id))
		}
	}

	return v, nil
}
```

**Nota sobre `done func(err error)` en `OnSave`/`OnDelete`:** aunque `Presenter.Save`/`Delete` ya son
síncronos (devuelven `error` directo), **no cambies la firma pública de `CrudView.OnSave`/`OnDelete`**
(siguen siendo `func(done func(err error))` / `func(id string, done func(err error))`) — es lo que
`Render()` ya invoca desde los botones (§3.3), y cambiarla forzaría tocar toda esa capa sin necesidad.
`done(cfg.Presenter.Save(record))` es la adaptación correcta: llama `done` con el error síncrono.

### 3.3 `crudview.go` — recorta `Source`/`allItems`/`Reload`/`filter`, delega a `Presenter`

**Borra por completo** el struct `Source` (ya no existe fuente de datos propia) y el campo
`allItems []Item` de `CrudView`. `CrudView` gana un campo `Presenter view.Presenter`:

```go
type CrudView struct {
	Element

	Title     string
	Form      Component
	Presenter view.Presenter

	OnSelect func(it view.Item)
	OnNew    func()
	OnSave   func(done func(err error))
	OnDelete func(id string, done func(err error))
	OnCancel func()

	SearchPlaceholder string

	items     *SignalNodes
	selected  *SignalString
	search    *SignalString
	canSave   *SignalBool
	canDelete *SignalBool
}
```

`Item` (la definición local de 3 campos) se **borra**: usa `view.Item` en todo el paquete (idéntico
campo a campo, sustitución directa — no hace falta mapeo). `Source` se borra del todo (ya no hay
`Source{}` zero-value que decida "full-page vs con lista": esa decisión ahora es
`cfg.Presenter != nil`, que siempre es cierto tras `New` — revisa `Render()` por el uso de
`hasSource := v.Source.Caller != nil` y cámbialo a `hasSource := v.Presenter != nil`).

`OnError` (campo y `handleError`) se borra: `Presenter.Reload()`/`Save()`/`Delete()` YA devuelven el
error directo — no hace falta un callback aparte, el llamante del método ya lo recibe. Si `Render()`
necesita mostrar un error de `Reload()` en el `Init`, hazlo así (reemplaza `Reload`/`handleError`
completos):

```go
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
```

`Init` (86-96) pasa a ignorar o loggear el error de `Reload()` según la convención del paquete para
errores async-en-Init ya existente en otros componentes de `layout` — **grepea** cómo otros
componentes de `layout`/`dom` reportan errores desde `Init` (p.ej. `dom.Log`) y sigue ESE patrón, no
inventes uno nuevo.

`filter()` (129-162) cambia SOLO su fuente: `v.allItems` → `v.Presenter.Items()`, y `it view.Item` en
vez de `it Item`:

```go
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
```

`Select`/`Render()` **no cambian su forma** — siguen usando `v.selected`/`v.canDelete`/los botones
tal cual, solo que ahora `hasSource := v.Presenter != nil` (antes `v.Source.Caller != nil`) decide la
variante full-page vs con lista, y el botón guardar/eliminar sigue condicionado a `v.OnSave != nil` /
`v.OnDelete != nil` — que `New` ya solo asigna si `Presenter.CanSave()`/`CanDelete()` es cierto, así
que el comportamiento observable (qué botones aparecen) **no cambia**.

### 3.4 `README.md` — enlace muerto

`README.md:11` enlaza a `docs/PLAN.md`, que no existe (se borró en un plan anterior). Reemplaza esa
línea por un enlace a este mismo `docs/PLAN.md` (el que estás ejecutando) o bórrala si el README no
tiene una sección de "Implementation Plan" con sentido una vez cerrado este plan también.

## 4. Migrar los tests (obligatorio — los 4 archivos usan la forma vieja)

- **`crudview_test.go`**: `TestCrudView_Render_Basic` construye un `&CrudView{...}` a mano sin
  `Source` — no debería necesitar cambios (no usa `Source`). `TestCrudView_Render_WithSource` usa
  `Source: Source{Caller: &mock.Caller{}}` — cámbialo a `Presenter: <un view.Presenter falso o real
  construido con view.New sobre un router/mock.Caller>`.
- **`consumer_test.go`**: el `fakeCaller` local (`Call(op, args, cb func([]byte, error))`) **se
  borra por completo** — ya no hace falta, ni compila contra el `router.Caller` nuevo. Los 10 casos
  (`TestConsumer_NewNoWidgets`, `TestConsumer_ListOp`, `TestConsumer_ListRendersCards`,
  `TestConsumer_SelectPopulatesForm`, `TestConsumer_SaveWithFormData`,
  `TestConsumer_SaveInvalidForm`, `TestConsumer_DeleteSelected`, `TestConsumer_DeleteNoSelection`,
  `TestConsumer_ListErrorPropagated`, `TestConsumer_SearchFiltering`) se **reescriben** construyendo
  un `view.Presenter` real vía `view.New(caller, record, listOp, newList, project, opts...)` con un
  `router/mock.Caller` (o el `conformance.FakeCaller` de `view/conformance`, codec-free — preferible,
  evita que `layout` dependa de `router/mock` solo para tests) y pasándolo a `crudview.New(Config{
  ParentID: ..., Presenter: p})`. El *comportamiento* de cada caso (guardar manda los valores del
  form, no el registro original; seleccionar llena el form; búsqueda filtra; etc.) **no cambia** —
  solo cómo se arma el `Presenter` que lo prueba. No borres ningún caso: cada uno prueba una
  invariante real del ciclo CRUD.
- **`crudview_wasm_test.go`** / **`consumer_wasm_test.go`**: mismo patrón, con
  `router/mock.Caller{}` (ya usado hoy en el wasm test) o el fake de `view/conformance`. Borra el
  helper duplicado `github_com_tinywasm_fmt_Contains` en `consumer_wasm_test.go` (ya hay
  `fmt.Contains` en `tinywasm/fmt` — usa ese, no una copia local).

## 5. Fuera de alcance

- **No** cambies la paridad visual Pa100T (clases CSS, estructura DOM de `Render()`) — solo la fuente
  de datos que alimenta esas clases.
- **No** dupliques la lógica de `view.Presenter` dentro de `crudview` "por si acaso" — si algo falta
  en `Presenter` para que `crudview` pinte correctamente, el defecto está en `tinywasm/view`: repórtalo,
  no lo reimplementes aquí.
- **No** toques `css.go`/`svg.go` (SSR split, sin relación con este cambio).

## 6. Criterios de aceptación

- `go build ./...` y `GOOS=js GOARCH=wasm go build ./...` verdes con `router@v0.1.13`,
  `view@v0.1.0`(+).
- `grep -rn "\.Call(" crudview/*.go` (no-test) **vacío** — ningún `Caller.Call` directo queda en
  `crudview`.
- `grep -n "type Item struct" crudview/*.go` **vacío** — se usa `view.Item`.
- `grep -n "type Source struct" crudview/*.go` **vacío**.
- Los 4 archivos de test migrados y en verde, sin ningún `fakeCaller`/`mock.Caller` con la firma
  vieja de 3 argumentos.
- `gotest ./...` (o `go test ./...` + `GOOS=js GOARCH=wasm go test ./...` si `gotest` no está
  disponible) verde en ambos targets.
- `README.md` sin enlaces muertos a `docs/PLAN.md`.

## 7. Etapas

| # | Etapa | Archivo(s) | Criterio |
|---|---|---|---|
| 1 | Bump deps | `go.mod`, `go.sum` | `router@v0.1.13`, `view@v0.1.0`(+) |
| 2 | `Config`/`New` | `crud.go` | forma de §3.2 |
| 3 | `CrudView`/`Reload`/`filter` | `crudview.go` | forma de §3.3; `Source`/`Item` locales borrados |
| 4 | README | `README.md` | enlace muerto corregido |
| 5 | Migrar tests | `crudview_test.go`, `consumer_test.go`, `crudview_wasm_test.go`, `consumer_wasm_test.go` | §4, verdes |
| 6 | Verificación | — | `gotest ./...` verde en ambos targets |
