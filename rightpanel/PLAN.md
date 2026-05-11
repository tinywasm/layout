# PLAN: Migración de tokens CSS v2 — `tinywasm/layout/rightpanel`

## Contexto

Migración consecuente al breaking change de `tinywasm/dom` que adopta la
convención Material Design 3 / Bootstrap 5. `rightpanel.css` usa los tokens
CSS viejos y debe actualizarse.

**Prerequisito bloqueante:** completar `tinywasm/dom` → `dom/docs/PLAN.md`
antes de ejecutar este plan.

---

## Tabla de sustitución

| Token viejo | Token nuevo |
|---|---|
| `--color-secondary` | `--color-primary` |
| `--color-tertiary` | `--color-muted` |
| `--color-quaternary` | `--color-surface` |
| `--color-gray` | `--color-background` |

---

## Archivo afectado: `rightpanel/rightpanel.css`

Uso actual de los tokens (según análisis):

| Selector | Propiedad | Viejo | Nuevo |
|---|---|---|---|
| `.rp-main` | `border-right` | `--color-tertiary` | `--color-muted` |
| `.rp-article` | `background` | `--color-gray` | `--color-background` |
| `.rp-article scrollbar-thumb` | `background` | `--color-tertiary` | `--color-muted` |
| `.rp-aside-header` | `background` | `--color-quaternary` | `--color-surface` |
| `.rp-aside-content` | `background` | `--color-quaternary` | `--color-surface` |
| `.rp-aside-content scrollbar-thumb` | `background` | `--color-tertiary` | `--color-muted` |
| `--rp-title-color` (custom var) | `color` fuente | `--color-secondary` | `--color-primary` |

---

## Checklist de implementación

**Prerequisito:** `tinywasm/dom` → `dom/docs/PLAN.md` completado y publicado.

- [ ] Leer `rightpanel/rightpanel.css` completo para confirmar la tabla arriba
- [ ] Aplicar sustituciones token por token según la tabla
- [ ] Verificar visualmente con el demo de `layout` si existe, o integrar en cualquier app de prueba
- [ ] `gopush 'feat(layout)!: migrate CSS tokens to v2 (Material/Bootstrap convention)'`
