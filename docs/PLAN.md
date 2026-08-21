---
PLAN: "feat(landing): srcset en las imagenes de Split y Cards"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 980339425852960845
---

> Este plan se despacha con el flujo CodeJob. Ver skill: agents-workflow.
>
> **PUERTA: no despachar hasta que `github.com/tinywasm/image` publique
> `image.Responsive`.** Este plan la usa; sin ella no compila.

# Plan — `layout/landing`: `srcset` en Split y Cards

Hacer que las dos imágenes de contenido de `landing` declaren las tres variantes
que el pipeline ya genera, en vez de servir la de escritorio a todo el mundo.

## El problema, medido

Sobre `veltylabs/mjosefa-website` en producción,
`grep -o 'srcset="[^"]*"' web/public/index.html` devuelve **vacío**. Un teléfono
de 400 px descarga los mismos 110 KB que un escritorio, cuando le bastaban
~25 KB. Las variantes ya se generan; falta **declararlas**.

## Lo que NO se toca

- **Los `.Lazy()` actuales se quedan.** Están en el media de `Split` y en las
  tarjetas de `Cards`, que van **bajo el pliegue**: ahí `lazy` es lo correcto. La
  única imagen que debe ser `eager` es la del hero, y esa la maneja `herobanner`,
  no este repo.
- **`platformd`, `crudview`, `rightpanel`, `login`**: no emiten imágenes de
  contenido. **No los toques.**
- El pánico de `RenderPages()` ante títulos o descripciones duplicados: es una
  red de seguridad deliberada.

---

## 0. Reglas de desarrollo

`landing.go` **compila para wasm** (no lleva build tag, y debe seguir sin
llevarlo). Por lo tanto:

- **Sin `fmt`, `errors`, `strconv`, `strings`, `log`** de la stdlib. Usa
  `github.com/tinywasm/fmt`.
- **Sin `map[K]V`**, sin `reflect`, sin `encoding/json`.
- **Anti-footgun:** `css.go` **sí** es `//go:build !wasm` y usa el DSL de estilo
  legítimamente. No le apliques las reglas de arriba.
- Código en inglés; documentación y comentarios de prosa en español.
- Sin strings mágicos: todo string repetido es una constante nombrada.

---

## 1. Los dos cambios

Ambos en `landing/landing.go`, y son los **únicos** dos puntos del repo que
emiten una imagen de contenido.

### 1.1 `Split` — el media lateral

Hoy:

```go
media.Child(image.Img(imageSrc, title).Lazy().AsElement())
```

Pasa a:

```go
media.Child(image.Responsive(imageSrc, title).Sizes(SizesSplit).Lazy().AsElement())
```

### 1.2 `Cards` — la imagen de cada tarjeta

Hoy:

```go
head = image.Img(c.Image, c.Title).Lazy().AsElement()
```

Pasa a:

```go
head = image.Responsive(c.Image, c.Title).Sizes(SizesCard).Lazy().AsElement()
```

### 1.3 Las constantes de `sizes` — donde está el valor real de este plan

```go
const (
	// SizesSplit: el media de Split ocupa el ancho completo en telefono y
	// la mitad cuando la banda se parte en dos columnas.
	SizesSplit = "(max-width: 768px) 100vw, 50vw"

	// SizesCard: las tarjetas van en grilla — una columna en telefono,
	// dos en tablet, tres en escritorio.
	SizesCard = "(max-width: 600px) 100vw, (max-width: 1024px) 50vw, 33vw"
)
```

**`sizes` es lo que hace útil al `srcset`, y omitirlo lo desperdicia.** Sin él el
navegador asume `100vw` y baja la variante como si la imagen ocupara la pantalla
entera — que para una tarjeta en una grilla de tres columnas significa bajar
tres veces más de lo necesario. El `srcset` sin `sizes` corrige el caso del
teléfono y no corrige el del escritorio.

**Los valores deben coincidir con los puntos de quiebre reales de `css.go`.**
Léelos antes de escribirlos: si el CSS parte la grilla en 900 px y aquí dice
1024 px, el navegador elige mal justo en esa franja. Si no coinciden con lo que
encuentres, usa los del CSS y **anótalo en el commit**.

---

## 2. Tests

| # | Caso | Espera |
|---|---|---|
| 1 | `Split` con imagen | el `<img>` trae `srcset` con las tres variantes |
| 2 | `Split` | `sizes` igual a `SizesSplit` |
| 3 | `Cards` con imagen | `srcset` con las tres variantes |
| 4 | `Cards` | `sizes` igual a `SizesCard` |
| 5 | `Split` y `Cards` | siguen con `loading="lazy"` |
| 6 | `Split` sin imagen | no emite el media, sin pánico |
| 7 | `Card` sin imagen | no emite la cabecera, sin pánico |
| 8 | `RenderPages()` con dos páginas de igual descripción | sigue entrando en pánico |

El caso 5 es de **regresión**: verifica que este cambio no tocó el `loading`, que
ya estaba bien. El 8 verifica que la red de seguridad sigue puesta.

---

## 3. Documentación

- `README.md` de `layout` — que `Split` y `Cards` reciben **rutas base sin
  sufijo de variante** (`/img/foto.jpg`, no `/img/foto.M.jpg`).
- `landing`: dejar anotado que `sizes` debe seguir a los puntos de quiebre de
  `css.go`, y que cambiar uno sin el otro degrada la elección del navegador sin
  romper nada visible — el tipo de error que nadie encuentra mirando.
- Si escribes diagramas: **nunca uses `subgraph`** (rompe el TUI).

---

## 4. Criterios de aceptación

- [ ] `go vet ./...` limpio; tests en verde con los 8 casos.
- [ ] `GOOS=js GOARCH=wasm go build ./...` sin errores.
- [ ] `grep -n "image.Img(" landing/*.go` → vacío.
- [ ] `grep -c "Lazy()" landing/landing.go` → **el mismo número que antes**.
- [ ] `git diff --stat` toca **sólo** `landing/`: ni `platformd/`, ni
      `crudview/`, ni `rightpanel/`, ni `login/`.
- [ ] `head -1 landing/landing.go` → **no** empieza con `//go:build`.
- [ ] `grep -n "\"fmt\"\|\"strings\"\|\"errors\"\|\"strconv\"" landing/landing.go` → vacío.
- [ ] Los valores de `SizesSplit` y `SizesCard` coinciden con los puntos de
      quiebre de `landing/css.go`.

## 5. Fuera de alcance

`herobanner` (vive en `tinywasm/components`, su propio plan), las apps
consumidoras, y el resto de los layouts. Tampoco AVIF ni art direction.
