---
PLAN: "login/html.go — que el login se sirva ya renderizado"
---
Depende de: el cambio de `#app` en `tinywasm/assetmin` (ver su `docs/PLAN.md`).
Sin eso, el HTML de acá cae fuera de `#app` y se ve duplicado.

## Qué se agrega

Un archivo, un método:

```go
// login/html.go
package login

func (l *Login) RenderHTML() string {
    if l.Title == "" {
        return ""
    }
    return l.Render().String()
}
```

`ssr` descubre `RenderHTML()` por nombre y mete lo que devuelva en `index.html`.

## Por qué `Render().String()` y no una plantilla aparte

El swap del WASM es invisible **sólo si** el markup del servidor es idéntico al
que produce `Render()` (ver el PLAN de assetmin: `dom.Render` hace
`innerHTML =`, borra y reemplaza). Dos implementaciones separadas divergen a la
primera edición y el salto vuelve. Una sola fuente lo garantiza por
construcción.

## Por qué el `if l.Title == ""`

Esto no es una guarda defensiva de adorno; es el punto central del diseño.

`ssr` genera el extractor instanciando **el valor cero**:

```go
inst := &m_login.Login{}      // Title:"", Form:nil, LogoMark:""
s.HTML += inst.RenderHTML()
```

Nadie le pasa parámetros. Así que un `RenderHTML()` incondicional emitiría una
tarjeta de login vacía en el `index.html` de **toda** app que importe
`layout/login`, aunque esa app no tenga login.

Devolver `""` en el valor cero significa: *el paquete solo aporta HTML cuando
alguien lo compone de verdad*. Y quien lo compone es la app, desde su propio
`RenderHTML()` de raíz:

```go
// en la app, config/html.go (!wasm)
func RenderHTML() string {
    return (&login.Login{
        Title:    AppName,
        Subtitle: "Ingrese sus credenciales para continuar",
        Form:     buildLoginForm().SetSSR(true).Render(),
    }).RenderHTML()
}
```

Es la misma asimetría que ya existe en `Render()`: `login` no sabe qué
formulario centra, la app se lo pasa. `RenderHTML()` la respeta en vez de
inventarse un login por defecto.

## El formulario en el servidor

`Form` es un `Component` que arma quien llama. En build `!wasm` no hay
`syscall/js`, así que el formulario tiene que renderizar sin tocar el DOM:
`tinywasm/form` ya trae `SetSSR(true)` para eso — emite `method`/`action` en vez
de depender del handler de submit en JS.

Consecuencia buena y deliberada: **el login funciona con el WASM apagado.** Es
un `<form method=post action=/login>` real. Si el WASM falla, tarda o el usuario
tiene JS bloqueado, se puede iniciar sesión igual. El WASM mejora la
experiencia, no la habilita.

Esto obliga a que `/login` acepte tanto POST de formulario (`application/x-www-form-urlencoded`)
como el POST JSON que manda hoy el cliente WASM. Eso es trabajo del lado app —
está en el PLAN de mjosefa-cms.

## Lo que NO cambia

- `Render()`, `RenderCSS()`, los campos, las clases: intactos.
- El markup es el mismo; solo se emite además desde el servidor.

## Pasos

1. `login/html.go` con el método de arriba (sin build tag: se usa en ambos
   lados).
2. Test: `RenderHTML()` en valor cero devuelve `""`.
3. Test: con `Title`+`Form` poblados, la salida de `RenderHTML()` es **igual**
   a `Render().String()` — es el test que impide que alguien "optimice" esto
   escribiendo una segunda plantilla y reintroduzca el salto visual.
4. Test: la salida contiene `login__card` y el markup del `Form` que se le pasó.
5. README: una línea en la sección `login` diciendo que se sirve desde el
   servidor y que el `Form` debe ir en modo SSR.

## Verificación

`gotest` en layout.
