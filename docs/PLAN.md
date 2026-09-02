---
PLAN: "feat!: crudview bulk actions — one options button, selection mode, batch delete and field patch"
TAG: v0.2.0
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 3635758060311448738
---

> Este plan se despacha con el flujo CodeJob. Ver skill: agents-workflow.
>
> Forma parte de una ola: `docs/BULK_ACTIONS_MASTER_PLAN.md` en la raíz del
> monorepo.
>
> **PUERTA: no arranques hasta que estén publicados** `form` (con
> `DirtyFields`), `view` (contratos plurales) y `components` (modo selección).
> Este plan los consume a los tres.
>
> **Es un cambio breaking.** Ver §8.

# Plan — `crudview`: acciones masivas

## 0. Prerrequisito

```bash
go install github.com/tinywasm/devflow/cmd/gotest@latest
```

Los tests se ejecutan con `gotest`. **Nunca `go test`.**

## 1. La forma nueva

```
MODO NORMAL          pie:  [      +      ][  🗑  ][  ✏  ]
                     fila: [ Pc Taller          192.168.122.30 ]
                     tap en la fila → carga la ficha en el formulario (como hoy)

  ↓ pulsar 🗑  (o ✏)

MODO SELECCIÓN       pie:  [   ↺   ][ 🗑 3 ]
                     fila: [ ☑ Pc Taller        192.168.122.30 ]
                     tap en cualquier punto de la fila → marca/desmarca

  ↓ pulsar 🗑 3 → modal de confirmación, UNA vez para los 3
```

Con `✏` la selección funciona igual, pero al confirmar se abre el formulario
vacío: el usuario toca **sólo** el campo que quiere corregir y se aplica ese
campo a los N marcados.

## 2. La máquina de estados

Un único `SignalString` con un conjunto cerrado de valores. **No** tres
booleanos: tres booleanos permiten estados imposibles (menú y selección a la
vez), y esta es exactamente la clase de estado ilegal que el chasis prohíbe
representar.

Fichero: **`layout/crudview/crudview.go`**.

```go
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
```

Transiciones, exhaustivas:

| Desde | Gesto | Hacia | Efecto |
|---|---|---|---|
| `modeNormal` | pulsar `🗑` | `modeDeleting` | `list.SetSelectMode(true)` |
| `modeNormal` | pulsar `✏` | `modeEditing` | `list.SetSelectMode(true)` |
| `modeDeleting`/`modeEditing` | pulsar `↺` | `modeNormal` | `list.SetSelectMode(false)` (que ya limpia las marcas) |
| `modeDeleting` | pulsar `🗑 N`, N>0 | — | abre el modal; al confirmar → borra y vuelve a `modeNormal` |
| `modeEditing` | pulsar `✏ N`, N>0 | — | abre el formulario en masivo; al aplicar → parchea y vuelve a `modeNormal` |

Reglas duras:

- `🗑` y `✏` **sólo** se pintan en `modeNormal`, y además **sólo cuando
  `!v.active()`**: con una ficha abierta o un borrador en curso, entrar en modo
  masivo dejaría al usuario con dos contextos de edición a la vez. Se
  **deshabilitan, no se esconden** — un botón que desaparece sin explicación se
  lee como un fallo.
- El `✏` **no se pinta en absoluto** si el presenter no implementa
  `view.Updater`. Una acción visible que no puede funcionar es peor que una
  ausente, y el harness prefiere que lo imposible no sea representable a que
  falle al pulsarlo.
- El botón de commit (`🗑 N` / `✏ N`) está **deshabilitado con N == 0**.
- Salir de cualquier modo masivo **siempre** llama a
  `list.SetSelectMode(false)`, que ya limpia las marcas.

## 3. Etapa 1 — La interfaz `ListView`

Fichero: **`layout/crudview/crudview.go`**.

```go
type ListView interface {
	Component
	SetItems(items []view.Item)
	Items() []view.Item
	Count() int

	// Selection mode — implemented by targetlist and targetdate alike, both
	// by assembling the components/listselect lego piece.
	SetSelectMode(on bool)
	CheckedIDs() []string
	OnCheckedChange(fn func(n int))
}
```

`CloseMenus()` **desaparece**: ya no hay menús por fila que cerrar.

La fábrica de `Config` también cambia, porque `OnDelete` ya no existe en los
componentes:

```go
List func(selected *dom.SignalString, onSelect func(view.Item)) ListView
```

**No** metas `onCheckedChange` en la fábrica. `ListView` ya lo expone como
método, y pasarlo además por el constructor daría dos caminos para lo mismo —
principio 4 del harness, *"one way to do each thing"*. `crudview` lo engancha
en `Init`, después de construir la lista:
`v.list.OnCheckedChange(func(n int) { … })`.

Actualiza el valor por defecto en **los dos** sitios donde está escrito hoy
(`crud.go` en `New`, y `crudview.go` en `Init`) — están duplicados a
propósito y deben seguir coincidiendo exactamente.

## 4. Etapa 2 — El pie

Fichero: **`layout/crudview/crudview.go`**, en `Render`, donde hoy se
construye `toggle` y se asigna `v.panel.AsideFooter`.

`rightpanel` ya declara su pie como `Row(Space1)` (`layout/rightpanel/css.go`,
parte `aside-footer`), así que **acepta varios hijos sin tocar `rightpanel`**.
Sustituye el botón suelto por un contenedor con los botones dentro.

Partes nuevas en **`layout/crudview/css.go`**:

| Parte | Skin |
|---|---|
| `action-delete` | `As(Danger)`, mismo box que `action` |
| `action-edit` | `As(Primary)`, mismo box que `action` |
| `action-count` | el número junto al glifo: `FontSize(TextXs)`, `FontWeight(WeightBold)` |

**Ojo con `action`:** hoy lleva `style.Width(style.Full)`
(`layout/crudview/css.go`, parte `action`). Con hermanos al lado, `Full` hace
que se coma la fila. Cámbialo por `Grow()`, que toma el espacio libre pero
cede el suyo a los hermanos — es justo la distinción que documenta `Grow` en
`widget/style`: *"Grow takes the free space along the inline axis and nothing
else… Use Grow() for the item in a Row that should push its siblings to the
trailing edge."*

`🗑` y `✏` llevan sólo su glifo (sin texto), así que `MediaBox(AspectSquare)`
sobre el mismo `ControlBox()` los deja cuadrados y a la altura del `+`. No
inventes un token de ancho.

Visibilidad por modo: **estado, no clases a mano**. Cada botón lleva
`BindStateFunc(widget.Open, …)` leyendo `v.mode`, y `css.go` los revela con
`RevealedBy(widget.Open)`. Es el mismo mecanismo que ya usa `action-cancel`.

## 5. Etapa 3 — Borrado masivo

Fichero: **`layout/crudview/crudview.go`**.

```go
// bulkDeleteAction commits the marked rows. One call, not a loop: view.Deleter
// is variadic precisely so the whole batch is one statement, and a loop would
// bring back the half-applied failure the plural contract exists to prevent.
func (v *CrudView) bulkDeleteAction() {
```

- Lee `ids := v.list.CheckedIDs()`. Si está vacío, no hace nada (el botón ya
  debería estar deshabilitado; esto es el cinturón).
- Abre el modal `v.confirmDelete` que **ya existe**, con el mensaje adaptado al
  plural. Hoy el texto se arma con
  `lang.Translate("Delete", "%s?", "This", "action", "cannot", "be", "undone.")`
  y un `%s` con la etiqueta del registro. Para N registros el `%s` debe llevar
  el recuento, no una lista interminable de etiquetas: con 1 marcado, la
  etiqueta de ese registro; con más de 1, el recuento. Sigue usando `lang` —
  **no** metas literales en español ni en inglés a pelo.
- Al confirmar: una sola llamada `deleter.Delete(ids...)`, luego
  `v.setMode(modeNormal)` — que ya llama a `SetSelectMode(false)` y con ello
  limpia las marcas — y `v.Reload()`.
- Si devuelve error, deja el modo como está para que el usuario reintente, y
  reporta por el mismo camino que ya usa `deleteAction` hoy.

## 6. Etapa 4 — Edición masiva

Fichero: **`layout/crudview/crudview.go`**.

Es la parte con más filo. Tres cosas tienen que pasar juntas:

**a) Suspender el auto-guardado.** Hoy `New` conecta
`f.OnFieldChange(func() { v.autoSaveAction() })`. En modo masivo eso
persistiría campo a campo sobre un registro que no es de nadie. `autoSaveAction`
debe salir sin hacer nada cuando `v.mode` es `modeEditing`. Ponlo como primera
guarda del método, con un comentario que diga por qué.

**b) Formulario en blanco.** Al entrar en `modeEditing` con N marcados:
`v.form.Reset()` y `v.form.MarkPristine()`, para que la baseline quede vacía y
`DirtyFields()` devuelva exactamente lo que el usuario escriba a partir de ahí.

**c) Aplicar sólo lo tocado.**

```go
// bulkEditAction patches the marked rows with ONLY the fields the user
// touched. Sending whole records instead would silently revert every column
// someone else changed since this client last reloaded — see the master plan's
// "por qué ids + delta".
func (v *CrudView) bulkEditAction() {
```

- `ids := v.list.CheckedIDs()`; vacío → no hacer nada.
- `fields := v.form.DirtyFields()`. **Si está vacío, no llames a `Update`**:
  muestra el aviso de "no hay cambios" por el camino que el paquete ya use y
  quédate en el modo. Una llamada con cero campos es un error en `view` y
  aquí sería un bug del consumidor.
- `v.form.SyncValues(record)` sobre `v.Presenter.Record()`, igual que hace
  `saveAction`.
- Una sola llamada: `updater.Update(ids, record, fields)`.
- `updater` se descubre por type assertion sobre `v.Presenter`, exactamente
  como se hace hoy con `view.Saver` y `view.Deleter`. Si el presenter **no**
  implementa `view.Updater`, el botón `✏` no debe pintarse siquiera — una
  acción visible que no puede funcionar es peor que una ausente.
- Al terminar bien: `v.setMode(modeNormal)` (limpia las marcas por sí solo),
  `v.form.Reset()`, `v.Reload()`.

## 7. Etapa 5 — Tests

Ficheros: **`layout/crudview/bulk_test.go`** y
**`layout/crudview/bulk_stylesheet_test.go`**.

Comportamiento:

| Test | Comprueba |
|---|---|
| `TestBulkEntriesOnlyInNormalMode` | `🗑` y `✏` sólo están activos en `modeNormal` |
| `TestBulkEntriesDisabledWhileEditingARecord` | Con `active()` verdadero, ambos están deshabilitados |
| `TestDeleteEntryGoesStraightToSelection` | Pulsar `🗑` entra en `modeDeleting` sin paso intermedio |
| `TestDeleteModeTurnsOnListSelection` | Pulsar 🗑 llama a `SetSelectMode(true)` en el doble de `ListView` |
| `TestCommitDisabledWithNothingChecked` | Con 0 marcados el commit está deshabilitado |
| `TestBulkDeleteShipsOneCall` | 3 marcados → el presenter doble recibe **una** llamada `Delete` con los 3 ids |
| `TestCancelClearsSelection` | `↺` en modo selección llama a `SetSelectMode(false)` y vuelve a normal |
| `TestBulkEditSuspendsAutoSave` | En `modeEditing`, disparar `OnFieldChange` **no** llama a `Save` |
| `TestBulkEditShipsOnlyDirtyFields` | Tocar un campo de cinco → `Update` recibe ese único nombre |
| `TestBulkEditRefusesWithNoDirtyFields` | Sin cambios → **cero** llamadas a `Update` |
| `TestEditButtonAbsentWithoutUpdater` | Presenter sin `view.Updater` → el `✏` no se pinta |
| `TestRowTapStillLoadsRecordInNormalMode` | El camino de hoy sigue vivo |

Hoja de estilos:

| Test | Comprueba |
|---|---|
| `TestActionGrowsInsteadOfFillingTheRow` | `.crudview__action` ya **no** emite `width: 100%` |
| `TestFooterButtonsShareTheControlHeight` | Todos los botones del pie emiten `min-height: var(--control-height…)` |

`TestBulkEditShipsOnlyDirtyFields` y `TestBulkDeleteShipsOneCall` son los dos
que protegen las decisiones de la ola: delta en vez de registro completo, y una
llamada en vez de un bucle.

### La regla que decide CÓMO se escriben estos tests

El harness es explícito, y `crudview` es literalmente el ejemplo que usa:

> **An API is not published until a consumer-shaped test, inside the library
> itself, proves it.** […] if a CRUD layout is meant to take a model, generate
> a form from it, and ship it through a caller, then the library must contain a
> test that does exactly that — **with a real model, the real form package, and
> a fake caller**. If that test is awkward to write, the API is awkward to use,
> and you have found the defect before shipping it.

Por lo tanto, en estos tests:

- **`ListView` es real**: un `targetlist.TargetList` de verdad. **No** escribas
  un doble a mano. Un doble de `ListView` sólo demuestra que `crudview` llama a
  los métodos que tú mismo escribiste en el doble; no demuestra que el modo
  selección funcione de punta a punta, que es lo único que importa aquí.
- **El formulario es real**: el paquete `form` de verdad, generado desde un
  `model` de verdad. `DirtyFields()` es la pieza nueva de la que depende la
  edición masiva; un doble la dejaría sin probar justo donde puede fallar.
- **Lo único falso es el borde de E/S**: un `router.Caller` de mentira que
  registra las llamadas. Es el patrón que el harness nombra, y el que ya usa
  `view/conformance`.

Los tests de `TestBulkDeleteShipsOneCall` y `TestBulkEditShipsOnlyDirtyFields`
inspeccionan lo que llegó a ese caller falso — no un contador dentro de un
doble de presenter.

**Anti-footgun:** sólo librería estándar de testing. Nada de `testify` ni
`gomega`.

## 8. Lo que rompe

- `crudview.ListView` pierde `CloseMenus()` y gana tres métodos.
- `crudview.Config.List` cambia de firma (`onDelete` → `onCheckedChange`).

Consumidor conocido fuera de este repositorio, fase C de la ola:

```
app-demo/modules/medicalhistory/medicalhistory.go:130
```

## 9. Criterios de aceptación

- [ ] `gotest` en verde (vet, race, cover, wasm).
- [ ] `grep -rn "CloseMenus" layout/` → **sin resultados**.
- [ ] `grep -n "style.Width(style.Full)" layout/crudview/css.go` →
      **sin resultados** en la parte `action`.
- [ ] Ni el borrado ni la edición iteran — **una** llamada cada uno. Contar
      las llamadas es más preciso que buscar la forma del bucle (un
      `for i := 0; i < len(ids); i++` no lo pillaría):
      `grep -c "\.Delete(ids" layout/crudview/crudview.go` → **1**, y
      `grep -c "\.Update(ids" layout/crudview/crudview.go` → **1**.
- [ ] `grep -n "type crudMode" layout/crudview/crudview.go` → una línea, y no
      existen booleanos paralelos de modo:
      `grep -n "selectMode \*SignalBool\|menuOpen \*SignalBool" layout/crudview/crudview.go`
      → sin resultados.
- [ ] El valor por defecto de `List` coincide **literalmente** entre
      `crud.go` (en `New`) y `crudview.go` (en `Init`).
- [ ] Los tests usan colaboradores reales:
      `grep -n "targetlist.TargetList\|form.New" layout/crudview/bulk_test.go`
      → al menos una línea de cada. Si aparece un `type fakeList struct` o un
      `type stubForm struct`, el test no es de forma de consumidor y no vale.

## 10. Etapas

| # | Etapa | Ficheros | Depende de |
|---|---|---|---|
| 1 | `ListView` + firma de `Config.List` | `crudview.go`, `crud.go` | — |
| 2 | `crudMode` y transiciones | `crudview.go` | 1 |
| 3 | Pie con grupo de botones + skin | `crudview.go`, `css.go`, `svg.go` | 2 |
| 4 | Borrado masivo | `crudview.go` | 3 |
| 5 | Edición masiva | `crudview.go` | 3 |
| 6 | Tests | `bulk_test.go`, `bulk_stylesheet_test.go` | 4, 5 |
