---
PLAN: "feat: crudview.New wires form+model+caller once, with a consumer-shaped test"
TAG: v0.0.11
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Fase **C** (gate de las apps) de la ola CRUD Harness:
> https://github.com/tinywasm/app/blob/main/docs/CRUD_HARNESS_MASTER_PLAN.md
> Gates: requiere `tinywasm/model v0.0.14` y `tinywasm/form v0.2.14` publicados.

# Plan — `crudview.New`: el pegamento se escribe **una vez**, aquí

## El problema

`crudview.CrudView` expone el layout Pa100T (form izquierda, lista derecha, barra CRUD,
búsqueda) pero deja `Form Component` como **slot opaco** y los callbacks
(`OnSelect`/`OnNew`/`OnSave`/`OnDelete`/`OnCancel`) sin cablear. Resultado: **cada app tiene
que reescribir el mismo pegamento** — generar el form del modelo, rellenarlo al seleccionar
una card, validar y sincronizar al guardar, mandar por el `router.Caller`, recargar la lista.

Eso son ~40 líneas idénticas por app, y ya han producido deuda real aguas abajo
(`veltylabs/mjosefa-cms` declaró una interfaz local `type Model interface{ model.Fielder;
model.Encodable }` porque no encontró un tipo que nombrar en esa frontera).

Los tests actuales (`crudview_test.go`, `crudview_wasm_test.go`) prueban `crudview`
**aislado**, con un `Form` falso: **nunca cruzan `form` + un modelo real + un `router.Caller`**.
Por ese agujero se escaparon tres defectos de API que solo aparecieron cuando una app intentó
usarlo de verdad. Este plan lo cierra.

## La regla que este plan instaura

> **Una API no está publicada hasta que un test con forma de consumidor, dentro de la propia
> librería, la demuestra.** El test de la etapa C4 es esa demostración: construye la vista CRUD
> completa exactamente como la escribirá una app.

## Paso 1 — bumps de dependencia

`go.mod`:
- `github.com/tinywasm/model` → **v0.0.14** (en esta versión `model.Model` es el contrato
  completo: `Fielder` + `ModuleNaming` + `Encodable` + `Decodable`).
- `github.com/tinywasm/form` → **v0.2.14** (aporta `Form.LoadValues`, el inverso de
  `SyncValues`).
- `github.com/tinywasm/form` es una **dependencia nueva** de este repo. No hay ciclo:
  `form` depende de `css`, `dom`, `fmt`, `model` — nunca de `layout`.

## Paso 2 — `crudview/crud.go` (archivo nuevo): el constructor

```go
package crudview

// Config is everything a module decides about its CRUD view: its record and its ops.
// Everything else — generating the form from the record, filling it on select, validating
// and shipping it on save, reloading the list — is wired HERE, once, for every app.
type Config struct {
	// ParentID is the DOM id the form mounts under (a module passes its own ID).
	ParentID string

	Caller router.Caller
	Title  string // h1, top-left

	// Record is BOTH the form's source of truth and the payload sent over the wire.
	// model.Model (v0.0.14) is exactly that contract — never declare a local intersection.
	Record model.Model

	ListOp   string // required
	SaveOp   string // "" → no save button
	DeleteOp string // "" → no delete button

	Args   func() model.Encodable           // list args (nil → none)
	Decode func(raw []byte) ([]Item, error) // list response → cards

	// Fill returns the full record for id, from what the module ALREADY decoded in Decode.
	// A nil return (including a typed-nil pointer) resets the form — that is not an error.
	Fill func(id string) model.Model

	OnError func(err error)

	// SearchPlaceholder overrides the search box text ("" → "Search…").
	SearchPlaceholder string
}

// New builds the standard CRUD view and wires the whole form↔list↔transport cycle.
// A module calls this and passes its record and its ops. Nothing else.
func New(cfg Config) (*CrudView, error)
```

### Cableado exacto de `New` (no lo improvises)

```go
func New(cfg Config) (*CrudView, error) {
	if cfg.Caller == nil { return nil, fmt.Errorf("crudview.New: Caller is required") }
	if cfg.ListOp == ""  { return nil, fmt.Errorf("crudview.New: ListOp is required") }
	if model.IsNil(cfg.Record) { return nil, fmt.Errorf("crudview.New: Record is required") }

	f, err := form.New(cfg.ParentID, cfg.Record)
	if err != nil { return nil, err } // a record with no widgets fails HERE, loudly
	f.HideSubmit() // the CRUD bar owns save — the form must not paint its own submit

	v := &CrudView{
		Title:   cfg.Title,
		Form:    f,
		Source:  Source{Caller: cfg.Caller, ListOp: cfg.ListOp, Args: cfg.Args, Decode: cfg.Decode},
		OnError: cfg.OnError,
		SearchPlaceholder: cfg.SearchPlaceholder,
	}

	v.OnSelect = func(it Item) {
		if cfg.Fill == nil { return }
		_ = f.LoadValues(cfg.Fill(it.ID)) // nil record → LoadValues resets. Not an error.
	}
	v.OnNew    = func() { f.Reset() }
	v.OnCancel = func() { f.Reset() }

	if cfg.SaveOp != "" {
		v.OnSave = func(done func(err error)) {
			if err := f.Validate(); err != nil { done(err); return }
			if err := f.SyncValues(cfg.Record); err != nil { done(err); return }
			cfg.Caller.Call(cfg.SaveOp, cfg.Record, func(_ []byte, err error) { done(err) })
		}
	}

	if cfg.DeleteOp != "" {
		v.OnDelete = func(id string, done func(err error)) {
			rec := cfg.Fill(id)
			if model.IsNil(rec) {
				done(fmt.Errorf("crudview: no record for id %s", id))
				return
			}
			cfg.Caller.Call(cfg.DeleteOp, rec, func(_ []byte, err error) { done(err) })
		}
	}

	return v, nil
}
```

**Decisiones tomadas — no las re-evalúes:**

- **`Record` es `model.Model`, no una interfaz local.** Ese tipo existe justo para esta
  frontera: el form necesita `Fielder`, el `Caller` necesita `Encodable`, `Decode` necesita
  `Decodable`. Si te falta un contrato, el defecto está en `model`, no aquí: **detente y
  repórtalo**.
- **Delete manda el registro completo, no un ID suelto.** Es lo que evita inventar un tipo de
  payload nuevo (`IDArg`) solo para borrar: el registro seleccionado ya está en la caché del
  módulo y `Fill(id)` lo devuelve. El tool de backend lee el ID del registro. Un tipo nuevo
  sería una frontera más que nombrar y mantener.
- **`f.HideSubmit()`.** El botón 💾 de la barra CRUD es el que guarda. Un submit dentro del
  form pintaría dos botones de guardar y dispararía dos caminos distintos.
- **`Fill` la implementa el módulo con su caché privada** de lo que ya decodificó en `Decode`.
  **No** añadas `Raw []byte` a `Item` (metería el payload de dominio en un tipo de
  presentación) y **no** dispares una op `get_x` por click (un viaje de red extra para traer
  lo que ya tienes).

## Paso 3 — `SearchPlaceholder` (quitar el string hardcodeado)

`crudview.go` tiene hoy `Input("search").Attr("placeholder", "Buscar...")` — un literal, y
además en español, dentro de una librería del ecosistema. Viola la regla "cero strings
hardcodeados en lógica".

- Añade el campo `SearchPlaceholder string` al struct `CrudView`.
- En `Render()`: usa `v.SearchPlaceholder`; si está vacío, usa la constante
  `const defaultSearchPlaceholder = "Search…"`.
- El idioma es decisión de la app, no de la librería.

## Paso 4 — el test con forma de consumidor (la razón de ser de este plan)

Crea `crudview/consumer_test.go` (sin tag: lógica pura) y `crudview/consumer_wasm_test.go`
(`//go:build wasm`, DOM real). **No** reutilices los dobles opacos de `crudview_test.go`:
este test debe atravesar `form` + un modelo con widgets + un `router.Caller`, que es
exactamente la pila que usa una app.

Piezas del test:

```go
// Un modelo con la forma EXACTA que emite ormc — y con widgets de verdad
// (input.Text() es un model.Kind: el Kind ES el widget).
var deviceModel = model.Definition{
	Name: "device",
	Fields: model.Fields{
		{Name: "id",   Type: input.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "name", Type: input.Text(), NotNull: true},
		{Name: "ip",   Type: input.Text()},
	},
}

type Device struct{ Id, Name, Ip string }

func (d *Device) ModelName() string       { return "device" }
func (d *Device) Schema() []model.Field   { return deviceModel.Fields }
func (d *Device) Pointers() []any         { return []any{&d.Id, &d.Name, &d.Ip} }
func (d *Device) IsNil() bool             { return d == nil }
func (d *Device) EncodeFields(w model.FieldWriter) { w.String("id", d.Id); w.String("name", d.Name); w.String("ip", d.Ip) }
func (d *Device) DecodeFields(r model.FieldReader) { d.Id, _ = r.String("id"); d.Name, _ = r.String("name"); d.Ip, _ = r.String("ip") }

var _ model.Model = (*Device)(nil) // ← si esto no compila, el contrato se rompió

// Caller falso: registra la op y los args, devuelve bytes enlatados.
type fakeCaller struct {
	ops  []string
	args []model.Encodable
	resp []byte
	err  error
}

func (c *fakeCaller) Call(op string, args model.Encodable, cb func([]byte, error)) {
	c.ops = append(c.ops, op)
	c.args = append(c.args, args)
	cb(c.resp, c.err)
}
func (c *fakeCaller) Dispatch(op string, args model.Encodable) { c.Call(op, args, func([]byte, error) {}) }
```

Casos que **deben** existir (cada uno cierra un defecto real de esta ola):

1. **`New` con un modelo sin widgets falla.** `Record` con `model.Text()` en vez de
   `input.Text()` → `New` devuelve error. *(Sin esto, la app pinta un form vacío en silencio.)*
2. **La op de listado que sale por el cable es `ListOp`.** `Init(ctx)` → `caller.ops[0] ==
   "list_devices"`. *(Este es el test que habría cazado el bug METHOD_NOT_FOUND.)*
3. **La lista pinta las cards** que devuelve `Decode` (Label + Description).
4. **Seleccionar una card rellena el form.** `OnSelect(Item{ID:"1"})` → los inputs del form
   contienen los valores del `Device` con id 1. *(Este es el test que exige `form.LoadValues`.)*
5. **Guardar manda `SaveOp` con los datos del FORM, no con el `Record` original.** Cambia el
   valor de un input, dispara `OnSave` → el `model.Encodable` que recibió el caller lleva el
   valor nuevo. *(Prueba que `SyncValues` corre antes del `Call`.)*
6. **Guardar con el form inválido NO llama al caller** y propaga el error por `done(err)`.
7. **Borrar manda `DeleteOp` con el registro seleccionado.**
8. **Borrar sin selección** (`Fill` devuelve nil) → `done(err)` y el caller **no** se llama.
9. **Un error del caller en el listado llega a `OnError`** — nunca se traga.
10. **La búsqueda filtra** las cards por Label y por Description.

## Paso 5 — documentación

- `docs/ARCHITECTURE.md`: sección nueva **"El pegamento vive aquí"**, con la regla:
  *una app configura (`crudview.Config`), nunca cablea*. Y la de la etapa C4: una API de este
  repo no se publica sin un test con forma de consumidor.
- `README.md`: el ejemplo de uso pasa a ser una llamada a `crudview.New` con `Config`. Borra
  cualquier ejemplo que construya `CrudView{...}` cableando callbacks a mano.
- **No borres el struct `CrudView` público ni sus callbacks.** Siguen siendo la costura de bajo
  nivel para un layout que no sea el estándar. Lo que se añade es el camino de alto nivel.

## Anti-footguns

- **Cero stdlib**: este repo compila a WASM. `tinywasm/fmt` para errores y conversiones —
  nunca `errors`, `strconv`, `strings`.
- **Embedding por valor**: `Element`, jamás `*Element`. Ya se cumple; no lo rompas.
- **SSR split**: CSS en `css.go` y SVG en `svg.go`, ambos con `//go:build !wasm`. Si
  `SearchPlaceholder` necesita estilos nuevos, van en `css.go`. Nunca CSS en `crud.go`.
- **Cero `map`, cero `any`, cero genéricos** en las firmas nuevas.
- Si al bumpear `model` a v0.0.14 algo de este repo deja de compilar, es un implementador a
  mano de `model.Model`: **detente y repórtalo**, no lo parchees con métodos vacíos.

## Criterios de aceptación

1. `crudview.New(Config) (*CrudView, error)` existe y cablea los cinco callbacks.
2. `Config.Record` es `model.Model`. `grep -rn "interface {" crudview/` no declara ninguna
   intersección local de tipos de `model`.
3. Los diez casos del test de consumidor pasan, en ambos targets (`gotest ./...`).
4. `grep -rn '"Buscar' crudview/` → vacío (el placeholder es configurable, con default en inglés).
5. El struct `CrudView` y sus callbacks siguen exportados (costura de bajo nivel intacta).
6. Una app puede escribir su vista CRUD **sin importar `tinywasm/form`**: `crudview.New` es
   suficiente.

## Tabla de etapas

| Etapa | Archivo(s) | Acción | Gate |
|---|---|---|---|
| C1 | `go.mod` | `model` → v0.0.14, `form` → v0.2.14 (dep nueva) | A, B publicadas |
| C2 | `crudview/crud.go` | `New(Config)` — el pegamento, una vez | C1 |
| C3 | `crudview/crudview.go`, `css.go` | `SearchPlaceholder` (fuera el literal "Buscar...") | C1 |
| C4 | `crudview/consumer_test.go`, `consumer_wasm_test.go` | test con forma de consumidor: 10 casos | C2, C3 |
| C5 | `docs/ARCHITECTURE.md`, `README.md` | "el pegamento vive aquí" + ejemplo con `Config` | C4 |
