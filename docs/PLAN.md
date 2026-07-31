---
PLAN: "platformd chrome pass — brand slot, honest selection, edge geometry, control rhythm"
TAG: v0.2.0
---

# Plan — `layout`: seven corrections to the platform chassis

Seven items from a review of the running demo. Each one below states what was
**measured**, not what was assumed — every number here came from the live page at
982×722, and the selectors quoted are from the served stylesheet.

Three of the seven are **not owned by this repository**. They are marked as such and
carry the repo that owns them. Per `app-releases/docs/CONSTRUCTION_HARNESS.md`, a
consumer never patches around a gap upstream:

> A missing contract at a boundary is a defect in the library, not in the consumer.

Order below is dependency order, not priority: item 3 has to land before items 1 and 2
can be verified in a single pass.

---

## 1. The header is two blocks and needs three

### What is there now

`platformd.go` `Render()` builds:

```go
header := Header().Set(clsHeader.AsAttr())
header.Child(msgSlot)   // Fill() — grows to eat the free space
header.Child(right)     // KeepSize() — the user menu
```

Two children. `msg-slot` carries `style.Fill()` and that is the only thing pushing
the user menu to the trailing edge.

### What it should be

```
[ brand ]        [ messages ]        [ user menu ]
 KeepSize          Fill, centred        KeepSize
```

The leading slot is where a platform's logo goes. Every real deployment has one, and
today there is nowhere to put it.

### The contract

Model it on `Identity`, which already solves the same problem for the avatar — the
shell asks for facts, the consumer supplies them, and `platformd` owns the rendering
and the fallback:

```go
// Brand is what the platform calls itself in its own chrome. The shell asks for a
// mark and a name; how they are drawn, sized and spaced is platformd's business.
type Brand interface {
	BrandName() string  // shown beside the mark, and as the mark's alt text
	BrandMark() string  // URL or inline SVG data URI; empty is normal
}
```

`BrandMark()` empty must be a **normal, expected outcome**, exactly as
`UserAvatar()` empty is: `platformd` falls back to its own glyph. Registering a
default mark in `svg.go` alongside `IconUser` is part of this item, not a follow-up.

⚠️ **Do not ask the contract for an `svg.Icon`.** That was already settled once for
the avatar — *"el icono no lo deberia pedir en el contrato de interfaz, para eso
existe platformd que ya tiene estilo y puede tener el icono por defecto"*. A
consumer supplies **facts**; sprite selection is a rendering decision. The same
reasoning applies here and the mistake is easy to repeat.

`Brand` must be optional (`nil` = no brand slot rendered), like `User` already is.
A platform without a logo is not a broken platform.

### Rendering

A new `brand` part holding the mark and the name, mirroring `usermenu`'s trigger so
the two ends of the header read as the same kind of object:

- mark: `IconBox(IconLg)` + `Round(RadiusFull)` + `HideOverflow()` — **the avatar's
  exact treatment**, which is what the request asked for
- name: `FontSize(TextBase)` + `FontWeight(WeightBold)`
- the slot itself: `Row(Space2)` + `KeepSize()`

`AppName` already exists on `Platform` and is currently **unrendered**. Do not let
it stay dead a fourth time — either `BrandName()` supersedes it and the field is
removed, or `AppName` is the fallback when `Brand` is nil. Decide, and write the
decision in the doc comment.

### The demo

`web/client.go` gains a `demoBrand` stub beside `demoIdentity`, returning a simple
inline SVG data URI. It exists to prove the empty-and-non-empty paths both work, so
it must return a **non-empty** mark — `demoIdentity` already covers the empty case
with its avatar.

---

## 2. The notification is a green box on the left; it should be centred and plain

### What is there now

`css.go`:

```go
Part(widget.Part("msg-slot"), style.Row(style.Space1), style.Fill()).
Part(widget.Part("msg"), style.Pad(style.Space1), style.Round(style.RadiusSm)).
Part(widget.Part("msg-info"),    style.As(style.Subtle)).
Part(widget.Part("msg-success"), style.As(style.Success)).   // ← the green box
Part(widget.Part("msg-warning"), style.As(style.Highlight)).
Part(widget.Part("msg-error"),   style.As(style.Danger)).
```

`As(Success)` resolves to `triplet{bg: --color-success, text: --color-on-success}` —
a filled `#1e7a30` block. And `Row(Space1)` with no content alignment leaves it hard
against the leading edge.

### What to change

- **`msg-slot` centres its content**: `CenterContent()` alongside the existing
  `Fill()`. It keeps `Fill()` — that is what holds the brand and the user menu apart
  at the two edges.
- **The variants carry no background.** Replace `As(<surface>)` with `Glyph(<family>)`
  on all four. `Glyph` emits `color` + `fill: currentColor` and no background, which
  is exactly "coloured text, no box". The severity stays legible; the slab does not.
- **Drop `Round(RadiusSm)` and `Pad(Space1)` from `msg`.** Both exist to shape a box
  that will no longer be there.

`Glyph(Subtle)` for info leaves the message the same muted grey as the rest of the
chrome — check it is still legible against the header panel before settling; if not,
info should use the on-surface colour rather than the muted one.

---

## 3. Selecting a row changes nothing — and the cause is a silent bug, not a colour

> **Owner: `github.com/tinywasm/components`** (`targetlist`), with a seam defect in
> `dom`/`widget`. `layout` cannot fix this.

### Measured

Clicking a row:

| | before click | after click |
|---|---|---|
| `data-selected` | *(absent)* | `""` |
| computed background | `rgb(242, 242, 247)` | `rgb(242, 242, 247)` |

The state **is** being written. The style never applies. The served rule is:

```css
.targetlist__row[data-selected="true"] { background: var(--color-selection, …) }
```

and the attribute in the DOM is `data-selected=""`. `""` is not `"true"`, so the
selector never matches. Nothing errors; the row simply never highlights.

### Root cause

Two ways to write a widget state, and one of them silently does nothing:

```go
// targetlist.go:185 — writes data-selected=""  → the rule never matches
BindAttrBool("data-selected", isSelSig)

// platformd.go — writes data-current="true"    → works
BindAttrFunc(cur.Key, func() string { if …  { return cur.Value }; return "" })
```

`widget.State.Attr()` declares the contract — `Selected → {data-selected, "true"}` —
and the style DSL emits a selector matching that value. `BindAttrBool` writes the
HTML boolean-attribute form (`=""`), which is correct for `disabled`/`checked` and
wrong for every `data-*` state this system defines.

This is the harness failure mode exactly: *more than one way to do the same thing*,
where the wrong one compiles and fails silently.

**The immediate fix** is `targetlist` switching to the `BindAttrFunc` + `attr.Value`
form platformd already uses.

**The real fix** is upstream and should be raised there: a state should be bindable
in one way that cannot be got wrong — e.g. a `BindState(widget.State, *SignalBool)`
helper in `widget`, so `BindAttrBool` is never reached for a data-state at all. File
it against `widget`; do not implement a local variant here.

### The colour, once the binding works

`Highlight` resolves to `--color-selection` =
`color-mix(in oklab, #1b5d8c, transparent 85%)` — a **15 % blue tint** over a light
panel. Even with the binding fixed this is close to invisible, so the request stands
on its own merits:

- **Selected row → `As(Accent)`** (`#e8a33d`, amber). Verified present in the token
  catalogue.
- **Current nav item → `As(Accent)`** too, replacing today's `As(Primary)`
  (`#1b5d8c`, measured live as `rgb(27,93,140)`). "Where I am" then means one colour
  across the whole chassis instead of two.
- **Hover must stop being amber.** `Cue(Hover, nav-link, As(Accent))` is the rule
  today; if selection becomes amber, hover and selected become indistinguishable and
  the problem is worse than before. Move hover to `As(Secondary)` or `As(Inset)` —
  a tonal shift, not a colour statement. This is the request's *"el hover deberia
  ser mas tenue o usar otro color para no confundir"*, and it is **not optional**:
  changing selection to amber without changing hover is a regression.

⚠️ `Accent`'s pair is `ColorAccent`/`ColorOnAccent` (`#1C1C1E` — dark text on amber).
Check the contrast where the nav icon rides the filled surface through
`currentColor`; the icon is currently drawn on a dark blue fill and will now sit on
a light amber one.

---

## 4. Fieldset labels are too tall and centred

> **Owner: `github.com/tinywasm/components`** (`fieldset`).

### Measured

| element | height | alignment |
|---|---|---|
| `.tw-field__label` | **26 px** | `justify-content: center` |
| `.targetlist__badge` | **20 px** | — |

Both are `ChipBox()` + `FontSize(TextXs)`, both 112 px wide (`--chip-width: 7rem`).
The 6 px difference is one declaration: the label adds `style.Pad(style.Space1)` and
the badge does not.

### Fix

In `fieldset/css.go`, `Part(widget.PartLabel, …)`:

- **remove `style.Pad(style.Space1)`** → the label matches the badge at 20 px
- **`style.CenterContent()` → `style.StartContent()`** → text starts at the leading
  edge, which is what the request asks for and what the rail already does once
  expanded

⚠️ The label is positioned with
`OnEdge(EdgeTop, SideStart, Space2, Space4)` — it **rides the input's top border**
and hangs half its height inside the box. Reducing its height from 26 px to 20 px
moves where that half falls. `Part(widget.PartInput, …)` carries
`Pad(Space4)` with the comment *"the legend rides the input's top border and hangs
half its height inside the box, so the value needs room to clear it"*. Re-check that
the value still clears the legend after the change; the padding may now be larger
than it needs to be.

**Follow-up (landed):** `StartContent()` left the chip text flush against the
leading edge — "no Pad" had over-corrected into "no air". `widget/style` gained
**`PadInline(Space)`** (`padding-inline`, zero height cost) and the legend now
carries `PadInline(Space2)`: the text sits 8 px off the chip's edges while the
height stays at the badge's 18–20 px. The fieldset test's blanket `padding` ban
was narrowed to height-affecting padding (`padding:` shorthand, `padding-block`).

---

## 5. Rounded corners on elements that touch the application frame

### Measured — every offender, at 982×722

| element | radius | touches |
|---|---|---|
| `pd__header` | 8 px | left + right + top |
| `pd__menu` (rail) | 8 px | right + bottom |
| `crudview` root | 4 px | left + bottom |
| `rp` (rightpanel root) | 8 px | right + bottom |
| `rp__header` | 8 px | right |
| `rp__title` | 4 px | right |

None of these set a radius explicitly. They inherit it from `As(Panel)` / `As(Primary)`,
which apply a `defaultRadius` per surface.

### The principle

A radius says "this box is a separate object floating on something". A box welded to
the window frame is not floating on anything — the curve just leaves four slivers of
background in the corners. Interior boxes keep their radius; boxes on the frame
square off.

### Fix

`style.EdgeToEdge()` already exists and emits exactly `margin: 0; border-radius: 0;`.
Apply it to the parts above. `crudview`'s root **already has it** — and still measures
4 px, so **check why before assuming the option is enough**. Likely a later rule with
equal specificity re-adds the radius; if so, that ordering bug is part of this item
and not a separate one.

⚠️ Do not square off the interior: `crudview__article`, `crudview__list`,
`crudview__fields`, `crudview__aside` and the targetlist rows keep their radius. The
request is explicit — *"los elementos internos se ven bien con esquinas
redondeadas"*.

**Consider raising upstream**: if "square where I meet the frame" recurs across
layouts, `widget/style` wants a first-class option for it rather than every layout
remembering `EdgeToEdge()`. Note it; do not build it as part of this work.

---

## 6. The search bar and the action button do not agree on a size

### Measured

| element | height |
|---|---|
| `.crudview__search` | 66 px |
| `.crudview__action` | 47 px |
| `.crudview__search-icon` | 40 px |
| `--control-height` | 72 px |

Three heights and none of them is the token. They sit at opposite ends of the same
column, so the disagreement is on screen at all times.

### Fix

Both are controls; both should be sized by the same token rather than by whatever
their padding adds up to.

- `Part("search", …)` and `Part("action", …)` both gain **`style.ControlBox()`**
  (`min-height: var(--control-height)`).
- `search-icon` fills its parent's height instead of being padded to 40 px, so the
  magnifier reads as the card's button edge to edge — the shape in the reference.
- Once both derive from `--control-height`, they agree **by construction**. Do not
  fix this by hand-tuning `Pad` steps until the numbers happen to match; that is what
  produced 66 / 47 / 40.

The mobile rules for `action` (`On(css.Mobile, widget.Part("action"), …)`) turn it
into a floating square and must keep doing so — verify `ControlBox` does not fight
`IconBox(IconLg)` there.

**Follow-up (landed):** sizing by the token was necessary but not sufficient — the
card's own `Pad(Space3)` still inflated it to 97.6 px (72 icon + 24 padding) and
the input, with no height of its own, floated at 25.6 px in the middle. The bar is
now **one fused control**, the shape in the reference: no card surface or padding,
`Row(SpaceNone)`, the wrapper's `Round(RadiusMd)` + `HideOverflow()` clips the pair;
the icon is a square via `MediaBox(AspectSquare)` (aspect-ratio, not padding, so the
width derives from the token too) and the input gained its own `ControlBox()`.
Measured: search = icon = input = action = **72 px** — criterion 9 holds exactly.
The input also inherits the app font now: the `css` reset gained
`input, textarea, select { font: inherit }` (UA default was Arial 13px inside a
field measured for system-ui 16px). Buttons were deliberately excluded — the theme
toggle's emoji is sized by that font.

---

## 7. Form inputs have no placeholders

> **Owner: `github.com/tinywasm/form`** (`input`), for the type-level part.

### Measured

`.tw-field__input` → `placeholder` attribute: **absent**.

### The plumbing already works

`form/render_input.go:294` emits the attribute whenever `GetPlaceholder()` is
non-empty. Nothing is broken; nothing sets one.

Two input types already do:

```go
// input/rut.go:20
r.SetPlaceholder("12345678-9")
// input/address.go:21
a.SetPlaceholder("Enter Address")
```

`input.IP()` (`ip.go`) does not, though it is the type in this demo with the least
guessable format — it accepts IPv4 **and** IPv6, so a user cannot infer the shape
from the field name.

### Fix

- **`input.IP()` sets its own placeholder**, e.g. `192.168.1.1`. A format hint
  belongs to the type that defines the format, which is the precedent `rut.go`
  already set. This is upstream work in `form`.
- The demo's `deviceDef` may add field-specific placeholders where the *meaning*, not
  the format, needs a hint. Those are demo copy and belong in `web/client.go`.

⚠️ `input/base.go:38` carries a deliberate note: *"Placeholder is NOT defaulted to
name: a host UI (e.g. components/fieldset) already shows the field name as a label
chip, so a placeholder mirroring it is noise."* Honour it. A placeholder that repeats
the label is worse than none — only add one where it teaches the **format** or the
**meaning**.

---

## Files

| repo | files |
|---|---|
| `tinywasm/layout` | `platformd/platformd.go`, `platformd/css.go`, `platformd/svg.go`, `platformd/web/client.go`, `crudview/css.go`, `rightpanel/css.go` |
| `tinywasm/components` | `targetlist/targetlist.go`, `targetlist/css.go`, `fieldset/css.go` |
| `tinywasm/form` | `input/ip.go` |
| `tinywasm/widget` | *(raise only)* a bind-safe state helper — item 3 |

Everything through local `replace`. **Do not publish.**

---

## Acceptance criteria

Each is measurable, and several exist because the eye cannot judge them.

1. `gotest` green in `layout`, `components` and `form`.
2. **Header**: three children; brand at `x ≈ 0`, user menu right-aligned, message
   block centred — assert `|centre(msg) − centre(header)| < 4 px`, not "looks
   centred".
3. **Brand fallback**: a `Platform` whose `Brand` returns an empty mark renders the
   default glyph at the same box as the avatar. A `Platform` with `Brand == nil`
   renders no brand slot and does not shift the other two.
4. **Notification**: `getComputedStyle(msg).backgroundColor` is
   `rgba(0, 0, 0, 0)` for all four variants, and the four differ in `color`.
5. **Selection**: after a row click, `data-selected === "true"` (**not `""`**) and
   the computed background differs from the unselected row. The current code passes a
   "the attribute is present" test — assert the *value* and the *colour*, or this
   regresses invisibly.
6. **Selection vs hover**: selected background ≠ hover background, as computed
   colours. This is the whole point of the hover change.
7. **Fieldset label**: height equals `.targetlist__badge` height exactly, and
   `justify-content` is `flex-start`. Re-verify the input value still clears the
   legend.
8. **Edges**: re-run the edge sweep that produced the table in item 5 — every element
   touching a viewport edge reports `border-radius: 0px`, and at least one interior
   element still reports a non-zero radius (so the fix did not square everything).
9. **Controls**: `.crudview__search` and `.crudview__action` report the **same**
   height, and it equals `--control-height`.
10. **Placeholder**: the IP field carries a `placeholder` attribute; no field's
    placeholder equals its label text.
11. Mobile at 375×812 unregressed: title and FAB absent off-route, one visible panel
    inside the drawer, theme toggle reachable. These were fixed recently and every
    item here touches shared chrome.
12. Console clean.

## Out of scope

- **Reworking the notification stack into a toast system** (positioning, queueing,
  animation). This item is about colour and placement of what already exists.
- **A brand slot on mobile.** There is no header on a phone; where the brand goes in
  the drawer is a separate design question, and guessing at it now would repeat the
  `AppName` mistake.
- **Retuning the colour palette.** Item 3 re-points existing surfaces at existing
  tokens; it does not change what `--color-accent` is.
- **The duplicate-id and post-mount work** tracked in `dom/docs/PLAN.md`.
