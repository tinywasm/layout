---
PLAN: "Doble menú en móvil: CueWithinHover en widget/style en vez de JS en platformd"
TAG: v0.1.6
STATUS: running
SESSION: 2026-08-05
---

# PLAN — Doble menú en móvil: CueWithinHover

> El tap en móvil dispara `:hover` sintético y `mouseenter` sintético. Por eso la
> regla `CueWithin(Hover, menu, drawer-panel)` flotaba un panel duplicado sobre el
> Drawer, y por eso el parche JS (`drawerHovered`, commits `67f1c18`/`3ef5516`)
> tampoco funcionó. La solución vive en la librería que posee el concepto:
> `widget/style`, con un CueWithin scoped a puntero fino.

## Antes de escribir código: lee CONSTRUCTION_HARNESS.md

**Es vinculante, no orientativo.** Principios que gobiernan este trabajo:

| # | Principio | Cómo se aplica aquí |
|---|---|---|
| 1 | Typed over `any` | `css.Capability` tipado, espejo del `css.Device` ya existente. |
| 2 | Explicit over implicit | `CueWithinHover` declara la intención: solo punteros que pueden hacer hover. |
| 3 | Illegal states unrepresentable | Sin parámetro de capability que acertar: el método es fijo a `(hover: hover)`. |
| 9 | Lego pieces, never forks | El fix NO va en `platformd` (prohibido: `syscall/js` solo en `dom`, y un wrap local sería un fork). Va en `widget/style`; `platformd` solo consume. |

---

## 1. Por qué existe

| Intento | Resultado |
|---|---|
| `CueWithin(Hover, menu, drawer-panel)` (CSS puro) | El tap móvil dispara `:hover` → panel flotante duplicado dentro del Drawer abierto. |
| `67f1c18`: `mouseenter`/`mouseleave` JS + `When(Open, drawer-panel)` | El tap móvil también dispara eventos mouse sintéticos → `drawerHovered=true` → mismo duplicado. |
| `3ef5516`: resets de `drawerHovered` en hamburger/overlay/Activate | Mitiga el síntoma, no la causa; parches sobre parches. |

La causa real: `:hover` y `mouseenter` se disparan en dispositivos táctiles aunque
no exista puntero fino. La señal correcta es el media feature `(hover: hover)`:
solo lo cumplen dispositivos cuyo puntero primario puede hacer hover. El tap no
lo cumple.

## 2. La solución (de abajo a arriba)

### 2.1 `tinywasm/css` — capability no-ancho

`Device` es una partición cerrada de anchos (`TestDeviceClassesPartition`) — no se
toca. Se añade un concepto ortogonal:

```go
type Capability uint8

const Hover Capability = iota

func (c Capability) Query() string // "(hover: hover)"
```

Test: `Hover.Query() == "(hover: hover)"`.

### 2.2 `tinywasm/widget/style` — `CueWithinHover`

- `sheet.go`: key `cueWithinHoverKey{cue, container, part}` + map
  `cueWithinHover`, init en `For()`, método
  `CueWithinHover(c widget.Cue, container, p widget.Part, opts ...Option) *Sheet`
  con `overlay = true` y acumulación idéntica a `CueWithin`.
- `validate.go`: espejar las 3 comprobaciones de `CueWithin` (container
  declarado, part declarado, container ≠ part).
- `emit.go`: tras el `@layer states` plano, emitir
  `@media (hover: hover) { @layer states { .n__menu:hover .n__part { … } } }`.
  Incluir el map en el barrido `hasMotion`.
- `shell_test.go`: el selector está DENTRO de `@media (hover: hover)` y el
  `@layer states` plano NO lo contiene; caso de error de validación.

### 2.3 `layout/platformd` — quitar el JS, consumir el método

- `platformd.go`: eliminar `drawerHovered` (campo), handlers
  `mouseenter`/`mouseleave`, `BindStateFunc(Open, …)` del drawerPanel, y los 3
  resets de `3ef5516`. `menuOpen` y sus resets se quedan.
- `css.go`: `When(Open, drawer-panel, Docked…)` → `CueWithinHover(Hover, menu,
  drawer-panel, Docked…)` (la forma original de la regla, ahora scoped). Las dos
  reglas de labels (`link-text`, `nav-link`) también pasan a `CueWithinHover`.
  Comment nuevo explicando por qué `(hover: hover)` y no JS.
- Test: el float solo existe bajo `@media (hover: hover)`.
- `go.mod`: bump de `tinywasm/css` y `tinywasm/widget`.

## 3. Orden de ejecución

1. Docs (este archivo + `widget/docs/SPECS.md`).
2. `css` → `gotest` → gopush.
3. `widget` → `gotest` → gopush.
4. `layout` → `gotest` → commit → **gopush solo tras revisión del usuario**.
5. Verificar con el MCP en `localhost:8080` (emulación móvil 375x812): tap en
   hamburger sin panel duplicado; hover de ratón en desktop sigue flotando el
   panel. Sin builds manuales (hot reload).

## 4. Calidad

- `gotest`, nunca `go test`.
- Cobertura del caso que falló: tap móvil no produce float (regla fuera de
  alcance por el media query) y hover desktop sí lo produce.
- Los asserts existentes de `nav-link:hover` (regla `Cue()`, no `CueWithin`) no
  cambian.
