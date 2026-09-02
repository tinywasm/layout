# Translation Dictionary — Consumer Guide

`layout` renders its own UI chrome (dialog titles, button labels, confirmation
messages) through `lang.Translate` from `github.com/tinywasm/fmt/lang`. The
dictionary itself is **consumer-owned**: the library never registers words, so
you decide the language your users see. Without any dictionary, everything
renders in English (the canonical keys pass through unchanged).

## What is translatable

| Component | Text | Dictionary keys |
|---|---|---|
| `crudview` delete dialog — title | "Confirm" | `"Confirm"` |
| `crudview` delete dialog — message | "Delete `%s`? This action cannot be undone." | `"Delete"`, `"This"`, `"action"`, `"cannot"`, `"be"`, `"undone."` (`"%s?"` passes through) |
| `crudview` delete dialog — buttons | "Cancel" / "Delete" | `"Cancel"`, `"Delete"` (reused from the message) |
| `platformd` hamburger — `aria-label` | "Menu" | not translated — the word is the same in both languages, so it is written as-is |

Every word is an independent key joined with spaces at read time, so keys are
reusable: `"Delete"` is shared between the message and the confirm button, and
the same entry can serve your own UI.

## How to add your dictionary

1. Import the language package: `import "github.com/tinywasm/fmt/lang"`.
2. Register your translations once (in `init()` or at app startup).
3. Activate the language with `lang.OutLang(lang.ES)` — or call
   `lang.OutLang()` with no arguments to auto-detect the system/browser
   language. English is the default and needs no activation.

```go
import "github.com/tinywasm/fmt/lang"

func init() {
	lang.RegisterWords([]lang.DictEntry{
		{EN: "Confirm", ES: "Confirmar"},
		{EN: "Cancel", ES: "Cancelar"},
		{EN: "Delete", ES: "Eliminar"},
		{EN: "This", ES: "Esta"},
		{EN: "action", ES: "acción"},
		{EN: "cannot", ES: "no"},
		{EN: "be", ES: "se"},
		{EN: "undone.", ES: "puede deshacer."},
		{EN: "records", ES: "registros"},
	})
}

func main() {
	lang.OutLang(lang.ES) // "Confirmar" / "Eliminar" / "Eliminar %s? Esta acción no se puede deshacer."
}
```

## Rules

- **EN is the lookup key** — always register the English word; other languages
  fall back to it when empty.
- **Words, not sentences** — pass each word to `Translate` separately (the
  library does this for you); dictionary entries are per word so they can be
  reused across the app.
- **Write keys in an order that reads naturally in your target language** —
  the words are joined in the given order. See the fmt translation guide
  (`github.com/tinywasm/fmt/docs/TRANSLATE.md`) for the word-order rules.
- **RegisterWords merges** — registering only some languages for a key never
  wipes the others; repeated registration adds translations to existing keys.
- **9 languages** — `EN, ES, ZH, HI, AR, PT, FR, DE, RU` are supported in
  `DictEntry`.
- Translation is resolved at read time: text built after a language switch
  uses the new language without rebuilding the component.
