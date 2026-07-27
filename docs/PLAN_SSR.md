---
PLAN: "`tinywasm/ssr`: sustituir la detección por regex de nombre por una interfaz tipada"
---

> Depende de: [`PLAN_WIDGET`](PLAN_WIDGET.md) (§6, `style.Styler`).
> Es el plan más pequeño de los cuatro y el que más fallos silenciosos elimina.

---

## El problema

De `AGENTS.md`, sección *"SSR asset provider names are matched by regex — the name must be
EXACT"*:

> *"`tinywasm/ssr` collects a package's SSR output by scanning `css.go`/`js.go`/`svg.go`/
> `html.go` for functions whose names match **exactly**: `RenderCSS`, `RootCSS`, `RenderHTML`,
> `RenderJS`, `IconSvg`. A CSS builder named anything else (e.g. `GenerateCSS`) is **silently
> never emitted** — the component renders with **zero styling** and nothing fails at build
> time."*

Y una segunda regla, tampoco expresada en ningún tipo:

> *"`ssr` requires all providers in a package to share ONE receiver (or all be free
> functions)… never mix a method with a free function, or receiver detection produces code
> that calls a method that doesn't exist."*

Esto es un contrato **por convención de nombres**, verificado por una expresión regular sobre
el texto fuente. Viola directamente el principio 6 (*fallar en compilación, nunca en
silencio*) y el 7 (*firmas auto-descriptivas*): la firma correcta no es descubrible por
autocompletado, y equivocarse produce un componente sin estilo que compila, arranca y se ve
mal en el navegador sin un solo error.

El síntoma está documentado hasta con su aspecto visual: *"a component renders unstyled while
its icons appear giant/black"*. Que exista una guía para reconocer el síntoma es la prueba de
que el defecto es estructural.

Peor aún: es la regla que **fuerza** la forma retorcida que hoy tiene el código. `AGENTS.md`
explica que `RenderCSS` debe declararse como **método** solo para no colisionar con la función
libre `RenderCSS` del paquete `css` importado con `.` — una restricción de diseño impuesta por
un detector de texto, no por una necesidad real.

---

## La solución: una interfaz

```go
// widget/style — ya definida en PLAN_WIDGET §6
type Styler interface {
	widget.Widget
	Style() *Sheet
}
```

`ssr` deja de escanear texto y **asevera la capacidad**, que es el patrón que la casa ya usa
en todas las demás costuras (`router.APIModule`, `view.Saver`, `view.Deleter`):

```go
// ssr — recolección
func Collect(parts ...widget.Widget) *Bundle {
	b := newBundle()
	for _, p := range parts {
		if s, ok := p.(style.Styler); ok {
			b.addSheet(s.Style())
		}
		if i, ok := p.(svg.IconProvider); ok {
			b.addIcons(i.Icons())
		}
		if h, ok := p.(HTMLProvider); ok {
			b.addHTML(h.HTML())
		}
	}
	return b
}
```

Lo que cambia, en concreto:

| Antes | Después |
|---|---|
| El nombre debe ser exactamente `RenderCSS` | El nombre lo fija la interfaz: `Style()` |
| Un nombre equivocado → CSS nunca emitido, sin error | Un nombre equivocado → **no satisface `Styler`** → error de compilación en el sitio de registro |
| Todos los proveedores del paquete comparten receptor | Irrelevante: son métodos de un tipo, no funciones sueltas |
| Método obligatorio solo por colisión con el `css` dot-importado | Ya no existe dot-import de `css` en los widgets |
| El detector lee el fuente | El detector no existe |

---

## Etapas

### Etapa 1 — Publicar las interfaces (aditivo)

Declarar `Styler`, `HTMLProvider`, `JSProvider` e `IconProvider`. `Collect` prueba **primero**
la aserción de interfaz y **cae** al escaneo por regex solo si el tipo no la satisface. Nada
se rompe; los paquetes migrados dejan de depender del escáner.

### Etapa 2 — Diagnóstico ruidoso para el camino viejo

Mientras conviven ambos, el escáner debe **gritar**. Al detectar un paquete con un
`RenderCSS` pero sin `Styler`, emitir un aviso en build:

```
ssr: paquete "crudview" usa la detección por nombre (obsoleta).
     Implementa style.Styler para obtener verificación en tiempo de compilación.
```

Y si detecta una función que *casi* coincide (`GenerateCSS`, `Styles`, `RenderCss`), **fallar
el build**. Ese es el caso exacto que hoy pasa en silencio y no debe sobrevivir ni durante la
transición: escala de preferencia del arnés, *error de compilación → diagnóstico ruidoso →
(nunca) fallo silencioso*.

### Etapa 3 — Retirar el escáner

Cuando `components` y `layout` estén migrados
([`PLAN_COMPONENTS`](PLAN_COMPONENTS.md), [`PLAN.md §8`](PLAN.md)): borrar el escáner de
fuente, las reglas de nombre exacto y la restricción de un receptor por paquete. Borrar
también de `AGENTS.md` la sección *"SSR asset provider names are matched by regex"* completa
—unas 25 líneas de reglas que el lector debía recordar— porque el tipo pasa a decirlas.

Es la reducción de documentación que el arnés promete: *"Because the API is the harness,
documentation shrinks to minimal 'how' instructions"*.

---

## Beneficio colateral: el orden de emisión deja de ser un accidente

Con `Collect` recibiendo `[]widget.Widget` en lugar de descubrir paquetes por escaneo, el
bundle puede ordenar la salida de forma determinista por capa
([`PLAN_WIDGET` §7](PLAN_WIDGET.md)):

```css
@layer tokens, primitives, widgets, states;
```

Hoy el orden de las hojas depende del orden de descubrimiento del escáner, que depende del
orden del sistema de archivos. Eso significa que **la cascada puede cambiar entre máquinas**.
Con capas explícitas, deja de importar: la capa manda sobre la especificidad y sobre el orden
de aparición.

---

## Criterios de aceptación

1. Un widget cuyo método de estilo se llame mal **no compila** en el punto de registro.
2. No queda escaneo de fuente en `ssr`; `grep -rn "RenderCSS\|IconSvg" ssr/` vacío salvo en el
   CHANGELOG.
3. La sección correspondiente de `AGENTS.md` está borrada, no reescrita.
4. El bundle emitido es idéntico byte a byte en dos ejecuciones y en dos máquinas.
5. Test en `ssr`: un tipo que satisface `Styler` emite; uno que solo tiene un `RenderCSS`
   suelto **no compila** al pasarse a `Collect` (test de compilación negativa).
