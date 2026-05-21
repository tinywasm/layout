# PLAN — Eliminar `SSRInstance()` de los layouts

## Objetivo

Quitar el boilerplate `func SSRInstance() *T { return &T{} }` de los layouts.
El contrato real de un layout SSR es la firma de sus métodos `Render*`; la
función accesoria sólo existe para que el extractor externo construya el
receiver, lo cual puede deducirse de la firma del método.

## Justificación

- `SSRInstance` no aporta lógica — cuerpo siempre `return &T{}`.
- Reduce un símbolo público por layout y mantiene la API mínima.

## Precondición técnica

Borrar `SSRInstance` rompe la compilación del extractor SSR upstream mientras
éste invoque `pkg.SSRInstance()` en el `main.go` que genera. Aplicar este plan
sólo cuando el extractor ya descubra el tipo receiver automáticamente desde la
firma del método (o acepte su ausencia como fallback). Verificación previa:

```bash
# Un layout sin SSRInstance debe extraerse sin error en el flujo upstream.
go test ./... -run TestExtract_NoSSRInstanceFunction
```

## Layouts afectados

| Layout | Archivo | Tipo |
|---|---|---|
| platformd | `platformd/ssr.go` | `*Platform` |
| rightpanel | `rightpanel/ssr.go` | `*RightPanel` |

En cada uno: borrar la función `SSRInstance` (con su comentario) y los imports
que queden huérfanos.

## Tests

- `grep -rn SSRInstance layout/` debe quedar vacío tras los cambios.
- `go test ./...` en `tinywasm/layout` verde.
- Verificación visual vía MCP browser de una página que use ambos layouts —
  sin regresiones.

## Documentación

- `layout/docs/` no tiene actualmente archivo de arquitectura. Si se agrega
  uno, no incluir referencia a `SSRInstance`.
- Verificar que ningún README documente `SSRInstance` como contrato público
  de los layouts.

## Migración adicional: `RenderJS()` → `[]*js.Script` (breaking change)

Los layouts que implementen `RenderJS()` deben migrar de `string` a
`[]*js.Script` (ver tipo `Script` en `github.com/tinywasm/js`). Reglas:

- Bundle global: `&js.Script{Content: content}` (Name vacío).
- Standalone crudo (escape hatch): `&js.Script{Name: "raw.js", Content: content}`.
- **Recomendado para SW/Worker:** constructores tipados de `tinywasm/js`
  (`js.ServiceWorker(name, handler)`, `js.WebWorker(name, handler)`) — el
  layout implementa la lógica como interfaz Go y el framework genera el
  JS-shim.

Estado actual: ningún layout implementa `RenderJS()` todavía. La migración
aquí se limita a usar la firma nueva si en el futuro se añade un layout con
JS (p.ej. `platformd` registrando un service worker para PWA).

Precondición técnica: `tinywasm/js`, `tinywasm/dom` y `tinywasm/assetmin`
publicados con el contrato `[]*js.Script`. Verificación:

```bash
go list -m github.com/tinywasm/js github.com/tinywasm/dom github.com/tinywasm/assetmin
```

## Migración adicional: `ssr.go` → split por extensión (breaking change)

El motor de `assetmin` deja de reconocer `ssr.go` y descubre assets por
archivos con nombre de extensión (`css.go`, `js.go`, `html.go`, `svg.go`),
todos `//go:build !wasm`. Ver el stage homónimo en `assetmin/docs/PLAN.md`.

Ambos layouts (`platformd`, `rightpanel`) solo implementan `RenderCSS()`, así
que cada `ssr.go` se **renombra directo a `css.go`** (contenido literal, mismo
receiver y `//go:embed`). No hay split en dos.

Precondición: `assetmin` publicado con la whitelist `ssrSourceFiles`. Aplicar
en el PR coordinado del cambio de motor.

## Stages

| # | Tarea | Done |
|---|---|---|
| 1 | Confirmar precondición técnica (test de extractor sin SSRInstance) | [ ] |
| 2 | Borrar `SSRInstance` en `platformd/ssr.go` | [ ] |
| 3 | Borrar `SSRInstance` en `rightpanel/ssr.go` | [ ] |
| 4 | Confirmar precondición `[]*js.Script` publicada en js/dom/assetmin | [ ] |
| 5 | Aplicar firma nueva si se añade `RenderJS()` a algún layout | [ ] |
| 6 | `go test ./...` verde en `tinywasm/layout` | [ ] |
| 7 | Verificación visual vía MCP browser | [ ] |
| 8 | Confirmar precondición: `assetmin` con whitelist `ssrSourceFiles` publicado | [ ] |
| 9 | Renombrar `platformd/ssr.go` → `platformd/css.go` | [ ] |
| 10 | Renombrar `rightpanel/ssr.go` → `rightpanel/css.go` | [ ] |
| 11 | `go test ./...` verde tras el renombrado | [ ] |
