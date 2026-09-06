# IE_LAYOUT — Ingeniería Inversa del Layout Principal de Pa100T

> Capturado de `http://192.168.122.10:1100` · Pa100T ver.3.0.1 · 2026-06-30

---

## 1. Estructura HTML del layout

Pa100T usa `<body>` directamente como raíz del grid (no hay wrapper `<div>`).

```html
<body>                                        <!-- CSS Grid root -->
  <div class="menu-container">               <!-- grid-area: menu-container -->
    <a class="btn-url-up" name="chat"  href="#chat">  <svg>…</svg> </a>
    <a                    name="service" href="#service"><svg>…</svg> </a>
    <a                    name="staff"   href="#staff">  <svg>…</svg> </a>
    <a                    name="medicalhistory" href="#medicalhistory"><svg>…</svg></a>
    <a                    name="reservation"   href="#reservation">  <svg>…</svg></a>
    <a                    name="patient"       href="#patient">      <svg>…</svg></a>
    <a                    name="device"        href="#device">       <svg>…</svg></a>
    <a class="btn-url-down" name="info" href="#info"> <svg>…</svg> </a>
  </div>

  <header>                                   <!-- grid-area: header -->
    <div id="user">…</div>                   <!-- col 1: usuario/logout -->
    <div id="mjs">…</div>                    <!-- col 2: mensajes de estado -->
    <h2  id="area">…</h2>                    <!-- col 3: nombre módulo activo -->
  </header>

  <picture name="logo-index">…</picture>     <!-- logo de fondo (pantalla inicio) -->

  <!-- Módulos: todos presentes en DOM, posicionados absolute -->
  <div id="chat"           class="module-content slider_panel">…</div>
  <div id="service"        class="module-content slider_panel">…</div>
  <div id="staff"          class="module-content slider_panel">…</div>
  <div id="medicalhistory" class="module-content slider_panel">…</div>
  <div id="reservation"    class="module-content slider_panel">…</div>
  <div id="patient"        class="module-content slider_panel">…</div>
  <div id="device"         class="module-content slider_panel">…</div>
  <div id="info"           class="module-content slider_panel">…</div>
  <div id="login"          class="module-content slider_panel">…</div>
</body>
```

**Clases CSS usadas en el layout:**
| Selector | Rol |
|---|---|
| `body` | Grid root (2×2: header + module-content + menu-container) |
| `.menu-container` | Columna de navegación vertical (derecha) |
| `.btn-url-up` | Primer ítem del menú (posición superior) |
| `.btn-url-down` | Último ítem del menú (posición inferior) |
| `.btn-url` | Ítem inactivo del menú |
| `.btn-selected` | Ítem activo del menú (aplicado al `<svg>` interno) |
| `header` | Barra horizontal superior, 3 columnas |
| `#user` | Celda izquierda del header: nombre/logout |
| `#mjs` | Celda central del header: mensajes de estado |
| `#area` | Celda derecha del header: módulo activo en mayúsculas |
| `.module-content` | Área de contenido — `grid-area: module-content` |
| `.slider_panel` | Panel de módulo, `position: absolute` |
| `picture[name="logo-index"]` | Logo decorativo de fondo (inicio) |
| `picture[name="logo-module"]` | Logo decorativo de fondo (módulos) |

---

## 2. Variables CSS / Tokens detectados

```css
:root {
  --header-height:  3vh;       /* ⚠ platformd usa 5vh */
  --content-height: 97vh;      /* ⚠ platformd usa 94vh/95vh */
}

/* Segunda declaración :root (interior del CSS) */
:root {
  --title-height:    8vh;
  --mag-pri:         .5rem;
  --mag-sec:         .2rem;
  --mag-cua:         .2rem;
  --controls-height: 9vh;
}
```

**Colores (hardcodeados en Pa100T, tokens en platformd):**
| Valor | Uso | Token platformd equivalente |
|---|---|---|
| `#e9e9e9` | fondo body, header, menu | `--pd-color-gray` ✓ |
| `#000000` | texto links | `--pd-color-quaternary` ✓ |
| `#ffffff` | texto `#area` | `--pd-color-primary` ✓ |
| `#c2c1c1` | borde `.module-content` | `--pd-color-tertiary` ✓ |

**Tamaños de grid:**
| Valor | Pa100T | platformd (`--pd-*`) |
|---|---|---|
| header height | `3vh` | `5vh` |
| content height | `97vh` | `94vh` (mobile) / `95vh` (desktop) |
| menu width | `4vw` | `5vw` |
| content width | `96vw` | `95vw` |

---

## 3. Flujos JS críticos

### 3.1 Navegación hash (módulo activo)

```js
// Click en <a> del menú → ONOFFURL + CheckAndRequestModule
document.querySelector(".menu-container").addEventListener("click", function(e) {
  if (e.target && e.target.tagName === 'A') {
    URLNOW = e.target;
    ONOFFURL(URLNOW);
    CheckAndRequestModule(URLNOW.name, true);
  }
});

// ONOFFURL: cambia clases CSS en el SVG
function ONOFFURL(urlNow) {
  if (URLOLD) {
    MODULO[URLOLD.name]?.ListenerModuleOFF();
    changeClass(URLOLD, "svg", "btn-url");    // inactivo
  }
  changeClass(urlNow, "svg", "btn-selected"); // activo
  URLOLD = urlNow;
}

// CheckAndRequestModule: carga lazy vía WebSocket si el módulo está vacío
function CheckAndRequestModule(moduleName, activateListener) {
  let div = document.getElementById(moduleName);
  if (div.children.length == 0) {
    SOCKSEND("module", moduleName, ...); // solicita HTML del módulo al servidor
  } else if (activateListener) {
    MODULO[moduleName]?.ListenerModuleON(); // activa listeners del módulo
  }
}
```

**No hay `hashchange` listener explícito.** La navegación se maneja por clicks en `<a>` con `href="#modulo"`. El hash en URL es efecto del `href`, no la causa.

### 3.2 Login → carga inicial de módulos

Tras login exitoso el servidor responde con lista de módulos autorizados (`msg.modules` JSON). Por cada módulo autorizado se llama `loadModuleInDom(div, html)` que inyecta el HTML y evalúa los `<script>` internos con `eval()`.

### 3.3 Módulo activo (indicador visual)

- Estado activo = clase `.btn-selected` en el `<svg>` dentro del `<a>` del menú.
- Estado inactivo = clase `.btn-url` en el mismo `<svg>`.
- No hay clase en el `<div#módulo>` — el "panel activo" no se controla por CSS sino por quién ocupa el `z-index`/`position:absolute`.

---

## 4. Diferencias con `platformd` y ajustes aplicados

| # | Aspecto | Pa100T | platformd actual | Acción |
|---|---|---|---|---|
| D1 | `--header-height` | `3vh` | `5vh` | ⬜ Ajustar token fallback a `3vh` |
| D2 | `--content-height` | `97vh` | `94vh`/`95vh` | ⬜ Ajustar a `97vh` desktop |
| D3 | `--menu-size` (ancho) | `4vw` | `5vw` | ⬜ Ajustar a `4vw` |
| D4 | `--content-width` | `96vw` | `95vw` | ⬜ Ajustar a `96vw` |
| D5 | `border-radius` panel | `0em .4em .4em 0em` | `0 .4em 0 0` | ⬜ Corregir: ambos lados derechos |
| D6 | `.btn-url-up`/`.btn-url-down` | primer/último item | no existe | ⬜ Agregar al CSS como modificadores opcionales |
| D7 | Grid root | `<body>` | `<div class="pd-root">` | ✅ OK — wrapper es mejor práctica |
| D8 | Carga de módulos | lazy por WebSocket | todos en DOM desde inicio | ✅ OK — webtyp renderiza todo en WASM |
| D9 | Activación panel | `z-index` implícito | clase `.pd-panel-active` | ✅ OK — más explícito y correcto |
| D10 | JS activo menú | `.btn-selected` en `<svg>` | `pd-nav-active` en `<a>` | ✅ OK — misma semántica, mejor target |
| D11 | `font-family: Arvo` | declarado en `body` | declarado en reset universal | ✅ OK — ambos aplican Arvo |
| D12 | `font-size` body | `16px` | desktop: `16px` | ✅ Coincide |
| D13 | Header col-2 | `#mjs` (status) | `clsMsgDesktop` | ✅ Equivalentes |
| D14 | Header col-3 | `#area` (texto módulo) | `clsArea` (BindText activo) | ✅ Equivalentes |
