# IE_MODULE_CONTENT — Ingeniería Inversa del contenido de módulo de Pa100T (form + lista)

> Capturado de `http://192.168.122.10:1100` · Pa100T ver.3.0.1 · 2026-07-08
> Complementa [IE_LAYOUT.md](IE_LAYOUT.md) (shell). Este documento cubre el
> **interior** de cada `.module-content`: la forma que casi todos los módulos
> repiten y que `crudview` debe replicar.

---

## 1. Estructura (dos columnas: formulario izquierda, lista derecha)

```html
<div id="device" class="module-content slider_panel">   <!-- panel del módulo -->

  <div class="article-contend">            <!-- COLUMNA IZQUIERDA (66vw) -->
    <div class="title-container">
      <div class="title"><h1>Computadores:</h1></div>   <!-- título sup-izq, blanco -->
    </div>
    <article>                              <!-- fondo blanco -->
      <div class="box-content">            <!-- marco gris #c2c1c1 -->
        <form class="form-distributed-fields">  <!-- FORMULARIO: flex-wrap -->
          <fieldset>                       <!-- por campo: min-width 45%, flex:auto -->
            <legend><label>Ip Dispositivo</label></legend>  <!-- chip azul -->
            <input type="text" name="device_ip">
          </fieldset>
          <fieldset>…Nombre…</fieldset>
        </form>
      </div>
    </article>
    <div class="controls">                 <!-- BARRA CRUD inferior -->
      <div class="contebuton">
        <button name="btn_cruddel">…</button>    <!-- eliminar (−) -->
        <button name="btn_crudcancel">…</button> <!-- deshacer (↺) -->
        <button name="btn_crudnew">…</button>    <!-- nuevo (+) -->
        <button name="btn_crudsave">…</button>   <!-- guardar (💾) -->
      </div>
    </div>
  </div>

  <div class="aside-contend">              <!-- COLUMNA DERECHA (29vw) -->
    <div class="aside-list">
      <div class="listaBox">               <!-- marco gris -->
        <ul class="lista">                 <!-- scroll propio -->
          <li class="target-li">           <!-- ítem 60px, card blanca -->
            Pc Administracion
            <span class="description-target">192.168.122.10</span> <!-- chip abajo-der -->
          </li>
        </ul>
      </div>
    </div>
    <div class="aside-search">             <!-- BÚSQUEDA inferior -->
      <label><svg><!-- lupa --></svg></label>
      <input type="search">
    </div>
  </div>
</div>
```

Variante sin lista: `.article-contend-full-page` (96vw, solo title+article, sin
controls ni aside).

## 2. CSS estructural capturado (fuente de verdad para `crudview`)

```css
:root {                      /* tokens del contenido */
  --title-height:    8vh;
  --controls-height: 9vh;
  --mag-pri: .5rem;  --mag-sec: .2rem;  --mag-cua: .2rem;
}

.module-content     { display: flex; border: .1vw solid #c2c1c1;
                      border-radius: 0 .4em .4em 0; }
.article-contend    { flex: none; display: grid;
                      grid-template: "title" var(--title-height)
                                     "article" 80vh
                                     "controls" var(--controls-height) / 66vw; }
.article-contend-full-page { display: grid;
                      grid-template: "title" var(--title-height)
                                     "article" 89vh / 96vw; }
.aside-contend      { flex: none; margin-left: var(--mag-pri); display: grid;
                      grid-template: "aside-list" var(--title-height)
                                     "aside-list" 80vh
                                     "aside-search" var(--controls-height) / 29vw; }

.title-container    { grid-area: title; display: flex; }
.title h1           { margin-left: .7em; color: #ffffff; }   /* sobre franja azul */

article             { grid-area: article; background: #fff;
                      border-radius: .4em .4em 0 0; display: flex;
                      flex-direction: column; margin-left: var(--mag-cua); }
.box-content        { flex-grow: 1; display: flex; min-height: 0;
                      background: #c2c1c1; border-radius: .4em;
                      padding: var(--mag-pri) 0 var(--mag-pri) var(--mag-pri); }
.form-distributed-fields { flex-grow: 1; overflow: auto; min-height: 0;
                      display: inline-flex; flex-wrap: wrap;
                      align-content: flex-start; flex-direction: row; }
fieldset            { background: #fff; border-color: #e9e9e9;
                      border-radius: .4em; min-width: 45%; min-height: 5em;
                      flex: auto; margin: .2em .2em .5em .2em; }
legend              { font-size: .8rem; background: #3f88bf;  /* chip azul */
                      border-radius: .4em; padding: 0 .5em .3em .5em; }

.controls           { grid-area: controls; margin-bottom: var(--mag-pri); }
.contebuton         { display: flex; height: 100%; background: #fff;
                      justify-content: space-between; padding: .1em;
                      border-radius: 0 0 .4em .4em; }
button[name*="btn"] { flex: auto; margin: .25em; padding: .4rem;
                      background: #3f88bf center no-repeat; background-size: 12%;
                      color: #fff; border-radius: .4em; }
/* iconos svg 16x16 embebidos: btn_crudnew=+, btn_cruddel=−,
   btn_crudcancel=flecha-undo, btn_crudsave=disquete (fill #fff) */

.aside-list         { grid-area: aside-list; margin-top: var(--mag-pri);
                      background: #fff; border-radius: .4em .4em 0 0;
                      display: flex; flex-direction: column;
                      padding-top: var(--mag-sec); }
.listaBox           { display: flex; height: 100%; flex-direction: column;
                      background: #c2c1c1; border-radius: .4em; padding: .2rem;
                      overflow: hidden; }
.lista              { flex-grow: 1; overflow: auto; min-height: 0; padding: .2em; }
.lista .target-li   { position: relative; display: flex; align-items: center;
                      min-height: 60px; max-height: 60px; width: 100%;
                      padding: .3em; margin-bottom: 1.2em; cursor: pointer;
                      border: 2px solid #e9e9e9; border-radius: .3em;
                      transition: .3s all ease; }        /* card blanca */
.description-target { position: absolute; right: 5px; margin-bottom: -55px;
                      background: #e9e9e9; color: #000; font-size: 80%;
                      border-radius: .4em; padding: .3em; } /* chip abajo-der */
.left-description   { border-right: 2px solid #e9e9e9; width: 2em; min-width: 2em;
                      font-weight: bold; text-align: center; }  /* inicial opcional */

.aside-search       { grid-area: aside-search; display: flex; background: #fff;
                      padding: .1em; border-radius: 0 0 .4em .4em;
                      margin-bottom: var(--mag-pri); }
/* .aside-search label: caja azul con lupa svg; input[type=search] al lado */
```

Estados relevantes: `button:disabled` (gris), `.target-li-on/off` (ítem
seleccionado), `.search-error`, `.head-aside-info` (franja azul 9vh sobre la
lista, usada por algunos módulos).

## 3. Semántica de interacción (observada)

1. La **lista** muestra los registros; click en un ítem carga sus datos en el
   **formulario** de la izquierda (y marca `.target-li-on`).
2. `btn_crudnew` (+) limpia el formulario para crear; `btn_crudsave` guarda
   (create o update según haya selección); `btn_cruddel` elimina el
   seleccionado; `btn_crudcancel` deshace la edición en curso.
3. Botones deshabilitados según estado: sin selección → `del` gris; sin cambios
   → `save` gris (ver captura de referencia: del/save grises, undo/new azules).
4. La **búsqueda** de abajo-derecha filtra la lista.
5. El título del módulo es claro y fijo en la **esquina superior izquierda**
   del panel (h1 blanco sobre franja azul), además del `#area` del header.

## 4. Mapeo a `crudview` (tinywasm/layout)

| Pa100T | crudview |
|---|---|
| `.article-contend` (66vw) | zona `Form` + título + barra CRUD |
| `.form-distributed-fields` + fieldsets | slot `Form` (típicamente `*form.Form`; los fieldset/legend los pinta `tinywasm/form`) |
| `.contebuton` (4 botones) | barra CRUD propia de crudview (callbacks New/Save/Delete/Cancel) |
| `.aside-contend` (29vw) | zona lista + búsqueda |
| `.lista .target-li` + `.description-target` | ítems tipados `Item{ID, Label, Description}` |
| `.aside-search` | búsqueda integrada (filtro local) |
| `.article-contend-full-page` | variante sin lista (lista/búsqueda ocultas cuando no hay Source) |
| colores hardcodeados (#3f88bf, #e9e9e9, #c2c1c1) | tokens del tema (el branding lo pone la app vía `RootCSS`) |
