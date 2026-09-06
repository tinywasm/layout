---
PLAN: "fix(crudview): drop a selection the filtered list no longer shows"
TAG: v0.2.0
EXECUTOR: jules
REVIEWER: none
---

> Este plan se despacha con el flujo CodeJob. Ver skill: agents-workflow.

# Plan — `crudview`: la ficha cargada nunca puede sobrevivir a un cambio de filtro

> **Idioma:** este documento está en español porque lo pidió el autor.
> **El código, los comentarios de código y los nombres de símbolos van SIEMPRE en
> inglés** — `webtyp/*` es librería pública. No traduzcas identificadores ni
> escribas comentarios en español dentro de los `.go`.

## Prerrequisito (ejecutar primero)

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

Se ejecuta `gotest` (nunca `go test`) desde la raíz del repo. No invoques
`gopush` ni `codejob`.

Este plan es **independiente**: no depende de los planes de `widget` ni de
`components`. Puede ejecutarse en paralelo con ellos.

---

## 1. El defecto

`crudview` es el controlador lista-detalle del framework: una columna con la
lista y otra con el formulario del registro elegido. Un control externo
(`Config.Filter`, cualquier `widget.Filterable`) estrecha la lista.

Reproducción real, en el módulo `medicalhistory` de la demo:

1. El selector de pacientes (un `selectsearch`) elige **Paciente A**.
2. La lista muestra las fichas de A. El usuario abre una: el formulario la carga.
3. El usuario vuelve al selector y elige **Paciente B**.
4. La lista se repuebla con las fichas de B — **pero el formulario sigue
   mostrando la ficha de A**.

El usuario está viendo, y puede editar y guardar, un registro que ya no existe
en la vista actual. Es el defecto más grave del lote: no es estético, es
corrupción de datos a un clic de distancia.

## 2. La causa exacta

En `crudview/crudview.go`, `Init` conecta el filtro así (líneas ~133-138):

```go
	if src, ok := v.Filter.(widget.Filterable); ok {
		src.OnFilterChange(func(term string) {
			v.search.Set(term)
			v.filter()
		})
	}
```

y `filter()` (línea ~423) es:

```go
func (v *CrudView) filter() {
	v.list.SetItems(v.Presenter.Filter(v.search.Get()))
}
```

Repuebla la lista y **nada más**. Las señales `v.selected`, `v.composing`,
`v.canDelete` y el formulario (`v.form`) quedan exactamente como estaban.

## 3. La solución: un invariante, no una opción

El autor pidió que esto sea *"una configuración de comportamiento global, ya
que ese caso se volverá a repetir con otro componente/layout"*. La forma
correcta de hacerlo global, según `CONSTRUCTION_HARNESS.md`, **no** es añadir
un flag de configuración:

> *3. **Illegal states unrepresentable.** If something must not happen, it must
> not be writable.*
> *4. **One way to do each thing.** A single construction pattern, with no
> alternatives that force a choice.*
> *8. **Closed by default.** … A resource left reachable because nobody said
> otherwise is a silent failure.*

Un formulario que sostiene un registro fuera de la lista visible **es un estado
ilegal**. No es una preferencia que un consumidor pueda querer al revés. Por eso:

> **Invariante (incondicional, sin flag, sin opt-in):**
> lo que el formulario sostiene tiene que ser algo que la lista sigue mostrando.
> Si el filtro deja fuera la selección, la selección se cae.

Al vivir dentro de `crudview` y no en cada módulo, **todo consumidor lo hereda
sin escribir una línea** — que es lo que "global" significa aquí. No hay nada
que un módulo pueda olvidar activar. Un futuro segundo controlador
lista-detalle heredará el mismo invariante extrayendo `clearSelection` a una
pieza compartida; hoy `crudview` es el único, así que vive aquí.

**Nota de alcance:** el borrador de un registro nuevo (`composing`) también
pertenece al ámbito viejo, así que también se cae. Un borrador de una ficha de
Paciente A no tiene sentido bajo Paciente B.

## 4. Cambios exactos — archivo `crudview/crudview.go`

### 4.1 Extraer el vaciado que ya existe

Hoy `undoAction()` (línea ~324) hace el vaciado completo. Ese cuerpo se
convierte en un método reutilizable, y `undoAction` pasa a ser
"vaciado + callback del usuario". Es DRY: un único sitio sabe qué significa
"aquí no hay nada seleccionado".

Sustituir el cuerpo de `undoAction` por:

```go
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
//
// CloseMenus alongside selected.Set(""): a row's ⋮ tap sets Selected directly
// (see targetlist's buildRow) so the mobile-docked Eliminar icon reads as
// belonging to that row, but native <details open> is separate state Selected
// does not touch. Without this, clearing drops the amber highlight while the
// floating icon for whichever row's menu was last opened stays on screen with
// nothing left for it to act on.
func (v *CrudView) clearSelection() {
	v.selected.Set("")
	v.canDelete.Set(false)
	v.composing.Set(false)
	if v.list != nil {
		v.list.CloseMenus()
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
```

> Conserva el comentario largo que hoy encabeza `undoAction` explicando por qué
> NO llama a `Form.Focus()`; muévelo a `undoAction` o intégralo arriba. No lo
> borres: documenta una cláusula del conformance (`cancel_clears_focus`).

### 4.2 Hacer cumplir el invariante en `filter()`

```go
// filter repopulates the list for the current term and then enforces the one
// invariant a list-detail controller cannot let slip: what the form holds must
// be something the list still shows.
func (v *CrudView) filter() {
	v.list.SetItems(v.Presenter.Filter(v.search.Get()))
	v.dropSelectionOutOfScope()
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
```

`ListView` ya expone `Items() []view.Item` y `view.Item.ID` es la clave de
selección — no hace falta ningún contrato nuevo. Verificable:

```bash
grep -n "Items() \[\]view.Item" crudview/crudview.go
```

## 5. Reglas de calidad obligatorias

- **Sin flag de configuración.** No añadas `Config.ClearOnFilter`, ni
  `KeepSelectionOnFilter`, ni nada que ofrezca el comportamiento roto como
  alternativa (principios 3, 4 y 8).
- **Sin strings sueltos en la lógica.** Este cambio no introduce literales; si
  necesitas uno, va a constante de paquete.
- **Sin librería estándar en código compartido con WASM.** `crudview.go` es
  neutro (sin build tag) y compila a WASM: usa `webtyp.com/fmt`,
  nunca `strings`/`errors`/`strconv` del stdlib. *Anti-footgun:* los `css.go`
  de este repo llevan `//go:build !wasm` y ahí el stdlib sí es legítimo — no
  "arregles" esos imports.
- **Superficie mínima.** `clearSelection` y `dropSelectionOutOfScope` quedan
  **sin exportar**. No son API pública.
- **Embebido por valor.** `CrudView` embebe `Element` por valor, nunca
  `*dom.Element`. No lo cambies.
- **No dupliques el vaciado.** Tras el cambio debe existir **un solo** sitio
  que ponga `selected`, `composing` y `canDelete` a cero.

## 6. Tests

Los tests de este repo son *consumer-shaped* y usan los dobles que ya viven en
`crudview/*_test.go` (`conformance.FakeCaller`, `fakeCtx`, `Device`,
`DeviceList`). Reutilízalos; **no** crees dobles nuevos ni uses testify.

Añadir a `crudview/crudview_test.go` (o un archivo nuevo
`crudview_scope_test.go`):

```go
func TestFilterDropsASelectionItNoLongerShows(t *testing.T) {
	// The reported bug: pick patient A, open one of A's records, switch to
	// patient B. The list repopulates; the form must not keep A's record
	// loaded, editable and saveable against a list that no longer shows it.
	v := newTestCrudView(t) // the package's existing helper

	// Select something the current list contains.
	items := v.list.Items()
	if len(items) == 0 {
		t.Fatal("fixture must provide at least one item")
	}
	v.selectAction(items[0])
	if v.selected.Get() == "" {
		t.Fatal("precondition: a record must be selected")
	}

	// Move the scope so the filtered list can no longer contain it.
	v.search.Set("no-such-scope-xyz")
	v.filter()

	if got := v.selected.Get(); got != "" {
		t.Errorf("selection must be dropped when it leaves the filtered list, still holds %q", got)
	}
	if v.composing.Get() {
		t.Error("a composing draft must be dropped with the scope it belonged to")
	}
	if v.canDelete.Get() {
		t.Error("delete must be disarmed once nothing is selected")
	}
}

func TestFilterKeepsASelectionStillInScope(t *testing.T) {
	// The invariant must not overreach: narrowing a search that still matches
	// the open record has to leave it alone, or typing in a search box would
	// wipe the form mid-edit.
	v := newTestCrudView(t)
	items := v.list.Items()
	v.selectAction(items[0])
	want := v.selected.Get()

	v.search.Set("") // widest possible scope: everything matches
	v.filter()

	if got := v.selected.Get(); got != want {
		t.Errorf("a selection still in the list must survive the filter: want %q, got %q", want, got)
	}
}

func TestFilterOnAnEmptyControllerIsANoop(t *testing.T) {
	// Nothing selected, nothing composing: filtering must not touch the form
	// or bounce a phone back to the list for no reason.
	v := newTestCrudView(t)
	v.search.Set("anything")
	v.filter() // must not panic and must leave the resting state alone
	if v.selected.Get() != "" || v.composing.Get() {
		t.Error("filtering an empty controller must change nothing")
	}
}
```

> Si `newTestCrudView` no existe con ese nombre, usa el patrón de construcción
> que ya emplean los tests del paquete (`&CrudView{Title: …, Presenter: …,
> Form: html.Div()}` seguido de `v.Init(&fakeCtx{})`) y extrae el helper.

## 7. Criterios de aceptación (verificables)

```bash
gotest                                                    # vet ✅ race ✅ tests ✅ wasm ✅
grep -c "v.selected.Set(\"\")" crudview/crudview.go       # 1 — un solo sitio vacía
grep -n "func (v \*CrudView) clearSelection" crudview/crudview.go        # 1 resultado
grep -n "func (v \*CrudView) dropSelectionOutOfScope" crudview/crudview.go  # 1 resultado
grep -n "dropSelectionOutOfScope()" crudview/crudview.go  # llamado desde filter()
grep -rn "ClearOnFilter\|KeepSelection" crudview/         # → vacío (no hay flag)
```

El conformance de vista (`view/conformance`) debe seguir verde sin tocarlo: las
cláusulas `cancel_clears_focus`, `edit_focuses_first_field` y
`new_focuses_first_field` dependen de `undoAction`/`newAction`, y este cambio
sólo reorganiza el primero.

## 8. Verificación manual (la hace el desarrollador)

En la demo, módulo **Ficha Paciente**:

1. Elegir Paciente A → abrir una de sus fichas → el formulario la carga.
2. Cambiar a Paciente B en el selector.
3. **El formulario queda vacío** y, en móvil, la vista vuelve a la lista.
4. Escribir en un buscador normal (módulo *Computadores*) con una ficha
   abierta y un término que **sigue** encontrándola: la ficha **no** se vacía.

## 9. Fuera de alcance (NO hacer)

- No toques `components/selectsearch`: sus defectos visuales van en
  <https://github.com/webtyp/components/blob/main/docs/PLAN.md>.
- No toques `widget/capability.go`: **no** hace falta un contrato nuevo
  (`ScopeSource` o similar). `widget.Filterable` + `ListView.Items()` ya
  alcanzan, y añadir una capacidad que un control puede olvidar implementar
  reintroduce el fallo silencioso que este plan elimina.
- No cambies `app-demo`: el arreglo se hereda sin tocar el módulo.
- No añadas `internal/`.

## 10. Etapas

| # | Etapa | Archivos | Cierra cuando |
|---|---|---|---|
| 1 | Extraer `clearSelection` de `undoAction` | `crudview/crudview.go` | `gotest` verde; `grep -c 'v.selected.Set("")'` = 1 |
| 2 | `dropSelectionOutOfScope` + llamada en `filter()` | `crudview/crudview.go` | `TestFilterDropsASelectionItNoLongerShows` pasa |
| 3 | Tests de no-regresión | `crudview/crudview_scope_test.go` | los otros dos tests pasan |
| 4 | Suite completa | — | `gotest` verde, conformance intacto |

La etapa 1 es **gate** de la 2.
