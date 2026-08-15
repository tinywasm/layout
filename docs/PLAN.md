---
PLAN: "feat: layout landing — sitio público multipágina desde datos tipados"
EXECUTOR: none
REVIEWER: none
BLOCKED_BY: "tinywasm/components (infobar, sitenav, herobanner, statgrid) y tinywasm/sitec (RenderPages) publicados"
---

> Este plan se despacha con el flujo CodeJob **cuando sus bloqueantes estén
> publicados**. Ver skill: agents-workflow.

# Plan — falta el layout de sitio público

## Contexto

Los cuatro layouts de este repo (`platformd`, `crudview`, `login`,
`rightpanel`) son shells de **aplicación**: una pantalla, sesión iniciada,
navegación interna. Un sitio público es otra cosa — varias URLs, cada una con
su metadata, contenido que viene de datos y no de estado de sesión.

`landing` es ese layout. El nombre es el término de industria para este
artefacto exacto y no se confunde con los otros cuatro.

## Bloqueantes

No empieces hasta que estén publicados:

- **`tinywasm/components`**: `infobar`, `sitenav`, `herobanner`, `statgrid`.
- **`tinywasm/sitec`**: el descubrimiento del productor `RenderPages()`.

Los tipos `html.Page` / `html.PagesProvider` y la metadata SEO
(`DocumentOptions.Description/Canonical/Image/JSONLD`) **ya están publicados**
en `github.com/tinywasm/html v0.0.17`. No los redefinas.

## La API

Una función por intención, al estilo `tinywasm/json` que
`CONSTRUCTION_HARNESS.md` señala como patrón de la casa. Una página se lee así:

```go
landing.New(marca,
	landing.InfoBar(contacto),
	landing.Header(menu...),
	landing.Hero(titular, bajada, cta, galeria...).At("inicio"),
	landing.Split("Nuestra Historia", foto, parrafos...).At("nosotros"),
	landing.Cards("Especialidades", tarjetas...).At("especialidades"),
	landing.Stats(cifras...).At("compromiso"),
	landing.Form("Voluntariado", intro, formulario).At("voluntariado"),
	landing.Hours("Contáctanos", contacto, horarios...).At("contacto"),
	landing.MapEmbed("Ubicación", mapaURL).At("ubicacion"),
	landing.Footer(menu...),
)
```

`.At(id)` fija el ancla de navegación de forma **explícita**. No derives el id
del título: un id derivado por magia se rompe en silencio el día que alguien
corrige una tilde del título y el menú deja de saltar.

### Tipos que el layout expone

`Brand`, `Link`, `Contact`, `Schedule`, `Card`, `Stat`, `Slide`. Nada más —
superficie mínima (principio 5). Son los tipos que un repo de sitio rellena;
el layout no sabe de dónde salen los valores (fichero, base de datos,
generador), y **eso es lo que lo hace reutilizable**.

Ningún tipo lleva HTML crudo dentro. Si una sección necesita una forma nueva,
es un tipo nuevo aquí, no un `string` con `<div>`.

### Secciones que son del layout, no componentes

`Split`, `MapEmbed`, `Footer` y `Hours` no tienen estado ni comportamiento: solo
tienen sentido como banda horizontal de una página. Viven aquí. Las que sí
tienen comportamiento o se usarían fuera de un landing ya son componentes y se
**consumen**, no se reimplementan: `infobar`, `sitenav`, `herobanner`,
`statgrid`, más `contentcard` (tarjetas) y `fieldset`+`form` (formularios).

## Multipágina

Es la mitad del valor de este layout. `landing` implementa
`html.PagesProvider`:

```go
func (p *Page) RenderPages() []html.Page
```

- La portada es `Path: "/"`.
- Cada página de detalle lleva su propia ruta, `Title`, `Description` y
  `Canonical`. **Dos páginas no pueden salir con el mismo `Title` ni la misma
  `Description`** — es la forma más común de perder el posicionamiento por
  página, y debe ser un test, no una recomendación.
- El `JSONLD` de la portada describe el negocio local (`schema.org`); las
  páginas de detalle describen su propio contenido. El layout **no** inventa el
  esquema: recibe el string ya armado, porque modelar schema.org no es su
  trabajo.

El caso motivador: una página por especialidad, listando sus servicios. Diseña
la API para "una colección de ítems produce N páginas", no cableado a la
palabra "especialidad" — el layout no sabe de clínicas.

## Restricciones

- **Nada de valores literales de estilo**: todo por tokens de `tinywasm/css`
  vía `widget/style`. Igual que en `components`.
- Sin carpetas `internal/`.
- Si a un componente le falta algo para servir aquí, **se arregla en
  `components` y se publica** — no lo envuelvas ni lo copies (regla lego).
- El layout no hace peticiones ni lee ficheros: recibe valores ya construidos.

## Verificación

La regla que mantiene honesto el arnés: **una API no está publicada hasta que
un test con forma de consumidor, dentro de esta librería, la prueba.**

- Test que arma una página completa con datos realistas y afirma:
  - el orden de las secciones es el de la llamada;
  - cada `.At(id)` aparece como ancla y el menú enlaza a ella;
  - `RenderPages()` devuelve la portada más una página por ítem de la
    colección;
  - **ninguna de las páginas comparte `Title` ni `Description` con otra**;
  - un `Path` terminado en `/` produce el `Path` esperado (no lo normaliza a
    algo distinto de lo que el sitio va a servir).
- `RenderCSS()` sin literales de estilo.
- `go build ./... && go vet ./... && go test ./...` verde.

## Etapas

| # | Alcance | Aceptación |
|---|---|---|
| 1 | Tipos (`Brand`, `Link`, `Contact`, `Schedule`, `Card`, `Stat`, `Slide`) + `New` + `.At` | compila; el test de composición arma una página |
| 2 | Secciones propias: `Split`, `MapEmbed`, `Footer`, `Hours` | orden y anclas probados |
| 3 | Secciones que envuelven componentes: `InfoBar`, `Header`, `Hero`, `Cards`, `Stats`, `Form` | ningún componente reimplementado |
| 4 | `RenderPages()` + metadata por página | portada + N páginas; títulos y descriptions únicos, probado |
| 5 | `RenderCSS()` del layout | sin literales; capas correctas |
