# PLAN — Typed CSS migration for tinywasm/layout

## Goal

Migrate `layout/rightpanel` (and any future layout package) to the typed CSS DSL. After this plan:
- `layout/rightpanel/rightpanel.css` is deleted.
- `layout/rightpanel/ssr.go` returns `*css.Stylesheet`.
- Class names used by rightpanel are declared as `css.Class` constants and shared with HTML emission.

## Scope

- `layout/rightpanel/` — sole layout module with a `.css` file today.
- Future layout modules added under `layout/` adopt the same pattern from day one.

## Why a separate plan from components

`layout/rightpanel` is structurally a component (implements the same SSR interfaces) but lives in `tinywasm/layout` because its role is page-level composition, not a reusable widget. It shares the migration mechanics with `tinywasm/components` but has its own consumers and its own docs surface (`layout/docs/`), so the plan is mirrored here for visibility to layout maintainers.

This plan **does not duplicate logic** with `components/docs/PLAN_typed_css.md` — it is intentionally short and points to the component template for canonical form.

## Migration steps

1. Wait for `tinywasm/css`, `tinywasm/dom`, and `tinywasm/components` plans to land. The DSL surface and the component migration pattern must exist first.

2. Audit `rightpanel/rightpanel.css`:
   - Enumerate selectors, declarations, token references.
   - Identify any layout-specific patterns (grid templates, named areas, container queries) that may require DSL extensions not yet covered by `tinywasm/css`. If found, file extension requests against the `tinywasm/css` plan **before** beginning the rewrite.

3. Declare `css.Class` constants in `rightpanel/ssr.go`:
   ```go
   var (
       ClsRightPanel    Class = "right-panel"
       ClsHead          Class = "right-panel__head"
       ClsHeadControls  Class = "right-panel__head-controls"
       ClsArticle       Class = "right-panel__article"
       ClsAside         Class = "right-panel__aside"
       ClsAsideControls Class = "right-panel__aside-controls"
   )
   ```
   Names follow the BEM-ish convention already implied by the existing CSS; align with whatever names the current `.css` uses to keep the diff focused.

4. Rewrite `RenderCSS()` to return `*Stylesheet` using the same template as components (see `components/docs/PLAN_typed_css.md`, "Component template").

5. Update `layout.go` and `rightpanel.go` to consume the new class constants through `dom.Class()` / `dom.Classes()`. Remove any string-literal class names.

6. Add `SSRInstance()`:
   ```go
   //go:build !wasm
   func SSRInstance() *RightPanel { return &RightPanel{} }
   ```

7. Delete `rightpanel/rightpanel.css` and the corresponding `//go:embed` directive.

8. Run tests; verify visual equivalence against the pre-migration output via golden snapshot.

## Files removed

- `layout/rightpanel/rightpanel.css`
- `//go:embed rightpanel.css` directive in `layout/rightpanel/ssr.go`.

## Files modified

- `layout/rightpanel/ssr.go` — rewritten as DSL.
- `layout/rightpanel/layout.go` (and any sibling files emitting class strings) — switch to `css.Cls<...>` constants.
- `layout/rightpanel/README.md` — refresh any code samples.

## Files added

- None.

## Interaction with existing layout `docs/PLAN.md`

The existing `layout/docs/PLAN.md` covers an unrelated topic (RightPanel slot embedding-pointer rule). It is **not affected** by this plan and stays as-is. This typed-CSS plan lives at `layout/docs/PLAN_typed_css.md` to avoid filename collision.

## Acceptance

- No `.css` file under `layout/`.
- No `//go:embed` directive in `layout/rightpanel/ssr.go`.
- `RightPanel.RenderCSS()` returns `*css.Stylesheet`.
- `SSRInstance()` exists and is callable by assetmin's invoke pipeline.
- Rendered CSS is visually equivalent to the legacy `rightpanel.css` (golden diff is whitespace-only or empty).
- `layout/docs/PLAN.md` is untouched.

## Out of scope

- New layout modules.
- Container queries or CSS grid named-area refactors beyond what `rightpanel.css` already uses.
- Changes to the slot embedding-pointer rule (covered by `layout/docs/PLAN.md`).
