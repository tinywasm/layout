# Fix: hardcoded Spanish UI chrome in tinywasm/layout (no i18n hook)

## Context

Following the `tinywasm/form/input` fix (raw translatable words stored, resolved
lazily via `lang.Translate` at read time, dictionary registered by the consumer —
never by the library), the user flagged that `tinywasm/layout` has the same class
of bug: several pieces of the library's own UI chrome (dialog title, button
labels, a confirmation message, an aria-label) are hardcoded Spanish string
literals with **no** `lang.Translate` call at all — `layout` doesn't even import
`github.com/tinywasm/fmt/lang` anywhere today.

This is framework-owned chrome text (not app-supplied content like a record's
`Title`/`Placeholder`, which is parameterized input from the consumer and stays
out of scope — translating an app-author-chosen literal isn't this library's
call to make). The scope here is exactly the boilerplate labels `layout` itself
authors: the delete-confirmation dialog in `crudview` and the hamburger
aria-label in `platformd`.

**Confirmed with the user**: the fix follows the input package's convention
exactly — dictionary entries use **English as the canonical key** (`"Cancel"`,
`"Delete"`, `"Confirm"`, `"Menu"`, matching the `"example:"` precedent), not
Spanish. This means the **default rendered text changes from Spanish to English**
unless the consuming app registers ES translations and activates them via
`lang.OutLang(lang.ES)` (or system-locale auto-detection) — there is currently no
such call anywhere in `tinywasm/app` or `tinywasm/layout`, confirmed by grep. That
activation is the consuming app's responsibility, same as dictionary
registration was ruled to be for `input`. This is an accepted, intentional
behavior change, not an oversight.

Out of scope (explicitly, to avoid unnecessary blast radius): `CrudView.Title` /
`RightPanel.Title` / `view.Presenter.Title()` / `view.WithTitle(...)`. That chain
carries an app-author-supplied literal (set once via `view.WithTitle("...")` at
presenter construction, a public API used by many call sites across the
ecosystem and by consumer repos outside this workspace, e.g.
`veltylabs/modules/clinical_encounter`, `veltylabs/mjosefa-cms` — not accessible
in this session). It is not "hardcoded framework chrome" the way the dialog/button
literals are, so it does not need this fix and reworking it would be a much
larger, unrelated, cross-repo breaking change. Same reasoning applies to the
`platformd/web/client.go` and other `web/client.go` demo files (`package main`,
illustrative sample code, not the reusable library surface).

## Fix

### `layout/crudview/crudview.go`
Add `"github.com/tinywasm/fmt"` (for `Sprintf`) and `"github.com/tinywasm/fmt/lang"`
imports (plain, not dot — the file already dot-imports `dom`/`html`).

- `Title: "Confirmar"` (line ~122) → `Title: lang.Translate("Confirm").String()`
- Delete-confirm message (line ~141-143): currently interpolates the record label
  directly into a hardcoded Spanish sentence inside `BindTextFunc`. Punctuation-heavy
  full sentences don't decompose cleanly into `lang.Translate`'s space-joined
  word list, so treat the whole sentence as one dictionary key with a `%s`
  placeholder, translate the template, then interpolate separately:
  ```go
  msg := P().BindTextFunc(func() string {
  	tmpl := lang.Translate("Delete %s? This action cannot be undone.").String()
  	return fmt.Sprintf(tmpl, "«"+v.deleteLabel.Get()+"»")
  })
  ```
- `cancel := Button()...Text("Cancelar")` → `.Text(lang.Translate("Cancel").String())`
- `confirm := Button()...Text("Eliminar")` → `.Text(lang.Translate("Delete").String())`

### `layout/platformd/platformd.go`
Add `lang "github.com/tinywasm/fmt/lang"` import (the file already dot-imports
`github.com/tinywasm/fmt`, so `lang` needs a normal import to stay unambiguous).

- `Attr("aria-label", "Menú")` (line ~292) → `Attr("aria-label", lang.Translate("Menu").String())`

### Tests (consumer-owned dictionary, same as `input`)
No `lang.RegisterWords` call in production code — mirror `form/input/inputs_test.go`:
register the ES translations and assert both languages inside the test itself.

- `layout/crudview/crudview_test.go` (or a new test in the package): build a
  `CrudView`, call `Init`, render, assert English defaults ("Confirm", "Cancel",
  "Delete", "This action cannot be undone.") appear; then
  `lang.RegisterWords([]lang.DictEntry{...})` + `lang.OutLang(lang.ES)` (deferred
  reset to `lang.EN`), render again, assert the Spanish text appears.
- `layout/platformd/platformd_test.go`: same EN default / ES-after-registration
  check for the `aria-label`.

## Verification

Run `gotest` in `/home/cesar/Dev/Project/tinywasm/layout` (per
[[feedback_use_gotest]]) and confirm all existing tests still pass — in
particular `crudview`'s behavioral tests (`consumer_test.go` Case 12,
`TestRowClick_WritesDataSelected`, etc.) which drive `confirmDelete`/
`confirmDeleteAction` programmatically and don't assert literal button text, so
they should be unaffected by the wording change. Confirm the new EN/ES tests
pass.
