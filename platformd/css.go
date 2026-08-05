//go:build !wasm

package platformd

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

// RenderSheet returns the style Sheet containing the rules for platformd.
func (p *Platform) RenderSheet() *style.Sheet {
	return style.For(p).
		// The outermost frame of the application: locks to the viewport, stacks
		// header over body. HideOverflow() keeps the frame itself from ever
		// scrolling — a tall module scrolls inside its own panel instead.
		Root(
			style.Cover(),
			style.HideOverflow(),
			style.As(style.Page),
		).
		// A slim bar, not a banner: Space1 padding and the IconMd avatar keep
		// it near 40px — the header frames the stage, it does not compete
		// with it.
		Part(widget.Part("header"),
			style.Row(style.Space2),
			style.KeepSize(),
			style.As(style.Panel),
			style.Pad(style.Space1),
			style.EdgeToEdge(),
		).
		// The brand slot mirrors the user menu at the other end of the header:
		// the mark is the avatar's exact box (IconMd, full round, clipped), the
		// name the trigger's text treatment.
		Part(widget.Part("brand"),
			style.Row(style.Space2),
			style.KeepSize(),
		).
		Part(widget.Part("brand-mark"),
			style.IconBox(style.IconMd),
			style.Round(style.RadiusFull),
			style.HideOverflow(),
			style.KeepSize(),
		).
		Part(widget.Part("brand-name"),
			style.FontSize(style.TextBase),
			style.FontWeight(style.WeightBold),
			style.KeepSize(),
		).
		// Fill() here is what pushes header-right to the far edge: it grows to
		// take the free space between the two blocks. CenterContent() makes the
		// message itself read in the middle of that space instead of hard
		// against the brand — the same block, held apart at the two edges.
		Part(widget.Part("msg-slot"),
			style.Row(style.Space1),
			style.Fill(),
			style.CenterContent(),
		).
		// No box: a notification is coloured text on the header's own surface.
		// Pad/Round existed to shape a slab that is no longer there. KeepSize
		// keeps a toast from being squeezed by its sibling; the part itself is
		// declared so the variant rules below have a shared home.
		Part(widget.Part("msg"), style.KeepSize()).
		// Glyph, not As: the variants tint the text and leave the background
		// alone, so severity stays legible and no green/red slab breaks the
		// header. Warning uses the accent family, not Highlight: Glyph(Highlight)
		// resolves to the surface colour and would be invisible on the panel.
		Part(widget.Part("msg-info"), style.Glyph(style.Subtle)).
		Part(widget.Part("msg-success"), style.Glyph(style.Success)).
		Part(widget.Part("msg-warning"), style.Glyph(style.Accent)).
		Part(widget.Part("msg-error"), style.Glyph(style.Danger)).
		Part(widget.Part("header-right"),
			style.Row(style.Space2),
			style.KeepSize(),
		).
		// The rail sits at the inline-end edge; the stage takes everything else.
		// Below the stage's minimum width the two reflow into one column with no
		// media query — that is Sidebar's own behaviour, not something to add.
		Part(widget.Part("body"),
			style.Sidebar(style.SideEnd, style.RailNarrow, style.SpaceNone),
			style.Fill(),
		).
		// Capas, no tira: los paneles se apilan y el activo entra deslizándose desde
		// el borde inicial. RevealedBy conmuta `display`, que es discreta y no puede
		// transicionar, de ahí que el movimiento venga de un transform.
		//
		// Esto NO es un scroller, y esa es la razón de ser del cambio: cuando lo era,
		// el scroll-snap horizontal de cada módulo (rightpanel MasterDetail, en móvil)
		// encadenaba con este al llegar a su extremo y un gesto dentro del contenido
		// arrastraba la aplicación al módulo siguiente. Un solo eje horizontal
		// desplazable por página.
		//
		// Todos los módulos siguen montados; el estado Current decide cuál se ve.
		Part(widget.Part("stage"),
			style.SlideDeck(style.MotionBase),
			style.Fill(),
		).
		// Scroll(), no Fill(): un módulo más alto que el viewport tiene que
		// desplazarse dentro de su propia capa. Scroll() es Fill() más overflow-y.
		//
		// Sin Anchor(): SlideDeck ya posiciona cada capa en absoluto, lo que la
		// convierte en el bloque contenedor de su contenido — que es lo que el cromo
		// flotante de un módulo necesita para resolverse contra SU panel. Añadir
		// Anchor() aquí lo ROMPE: emite position:relative en @layer widgets, que gana
		// sobre el position:absolute que el flujo emite en @layer primitives, y las
		// capas volverían al flujo apiladas una debajo de otra.
		Part(widget.Part("panel"),
			style.Stack(style.SpaceNone),
			style.Scroll(),
		).
		// Grow() for its min-width:0 — without it the rail is sized by whatever
		// it contains, so revealing the labels on hover widened it and pushed
		// the stage. Its width is the Sidebar's rail token and nothing else.
		// EdgeToEdge: the rail is welded to the window frame on its right and
		// bottom, where a radius leaves two background slivers.
		Part(widget.Part("menu"),
			style.Anchor(),
			style.Stack(style.Space1),
			style.As(style.Panel),
			style.Fill(),
			style.Grow(),
			style.EdgeToEdge(),
		).
		// The drawer's copy of the identity exists only where there is no header
		// to hold one. Two visible at once is the redundancy this whole change
		// set out to remove.
		OnlyOn(css.Mobile, widget.Part("drawer-identity"),
			style.Row(style.SpaceNone),
			style.KeepSize(),
		).
		// Text with no icon beside it: it can only exist where the panel opens
		// wholesale, which is the phone drawer. No hover rule reveals it on a
		// wide screen, on purpose.
		OnlyOn(css.Mobile, widget.Part("app-name"),
			style.Row(style.SpaceNone),
			style.Pad(style.Space2),
			style.FontSize(style.TextBase),
			style.FontWeight(style.WeightBold),
		).
		Part(widget.Part("drawer-panel"),
			style.Stack(style.Space1),
			style.Fill(),
		).
		Part(widget.Part("navbar"),
			style.Stack(style.SpaceNone),
			style.Fill(),
		).
		// No Fill: spreading the items down the rail makes every row resize the
		// moment the panel floats out and stops being full height. A rail packed
		// from the top measures the same collapsed and expanded.
		Part(widget.Part("nav-item"),
			style.Row(style.SpaceNone),
			style.KeepSize(),
		).
		// Glyph, not As: an item that is merely available is a coloured icon on
		// the rail's own surface. Only the current one is filled.
		Part(widget.Part("nav-link"),
			style.Row(style.Space1),
			style.Pad(style.Space2),
			style.Width(style.Full),
			style.CenterContent(),
			style.Glyph(style.Primary),
			style.Animate(style.MotionFast),
		).
		// A bare <svg> with no box falls back to 300x150 and wrecks the rail.
		Part(widget.Part("nav-icon"),
			style.IconBox(style.IconLg),
		).
		// The rail is icon-only: at RailNarrow a label does not fit, and forcing
		// it in is what widened the rail past its token. The mobile drawer is
		// two thirds of the viewport, so there the label rides along.
		// Row() is not decoration here: OnlyOn hides the part outside the device
		// and only a flow puts a display back on it inside, so a rule carrying
		// nothing but FontSize would stay hidden everywhere.
		// Left-aligned on a phone too: the drawer shows icon and label together,
		// which is the same row the expanded rail draws, so it aligns the same
		// way. Centring is only right for a lone icon.
		On(css.Mobile, widget.Part("nav-link"),
			style.Row(style.Space2),
			style.Pad(style.Space2),
			style.StartContent(),
		).
		OnlyOn(css.Mobile, widget.Part("link-text"),
			style.Row(style.SpaceNone),
			style.FontSize(style.TextBase),
		).
		// The active route reads as "current", the same vocabulary the rail and
		// crudview's list rows share. It is a STATE, never a class.
		// The route you are on is the one filled block in the rail; its icon
		// rides the filled surface through currentColor. Primary keeps the icon
		// white against a dark fill — the same scheme as the mobile hamburger
		// button, for a familiar look on both viewports.
		When(widget.Current, widget.Part("nav-link"),
			style.As(style.Primary),
		).
// El control entero se ilumina, no solo el glifo: el hover habla del blanco
		// al que apuntas, y el blanco es el botón. Inset, no Accent: la selección ya
		// es ámbar y un hover del mismo color sería indistinguible de ella.
		//
		// El borde de Inset NO ensancha la caja: una regla de estado lo emite como
		// outline (ver widget/style), así que el ítem mide lo mismo con y sin
		// puntero encima. Esa igualdad es obligatoria aquí — el rail mide exactamente
		// --rail-narrow y el panel flotante se dimensiona con width: max-content: dos
		// píxeles de más y el ítem se sale del puntero que lo activó, lo pierde,
		// encoge, y el menú entero entra en un bucle de parpadeo.
		Cue(widget.Hover, widget.Part("nav-link"),
			style.As(style.Inset),
		).
		// Hovering the rail floats the whole panel out over the content at label
		// width. The panel leaves the flow, so the rail's box — already pinned
		// to the Sidebar's rail token by Grow's min-width:0 — cannot change, and
		// nothing beside it moves.
		CueWithin(widget.Hover, widget.Part("menu"), widget.Part("drawer-panel"),
			style.Docked(style.Parent, style.EdgeTop, style.SideEnd, style.SpaceNone),
			style.Width(style.Content),
			style.As(style.Panel),
			style.Raise(style.Floating),
			style.Pad(style.Space1),
			style.Round(style.RadiusMd),
			style.Stack(style.Space1),
		).
		// The labels only exist while the rail is expanded — or on a phone,
		// where the drawer is two thirds of the viewport and has room for them.
		CueWithin(widget.Hover, widget.Part("menu"), widget.Part("link-text"),
			style.Row(style.SpaceNone),
			style.FontSize(style.TextBase),
		).
		// Left-aligned once expanded: centred is right for a lone icon in a
		// narrow rail, wrong for an icon-and-label row where the labels have to
		// start on the same line as each other.
		CueWithin(widget.Hover, widget.Part("menu"), widget.Part("nav-link"),
			style.Row(style.Space2),
			style.Pad(style.Space2),
			style.StartContent(),
		).
		// ── mobile-only chrome ────────────────────────────────────────────────
		// No header on a phone: the module brings its own title and the chrome
		// floats over the content. The button pins to the screen so it is
		// reachable from either page of a swipe strip.
		On(css.Mobile, widget.Part("header"),
			style.Hide(),
		).
		OnlyOn(css.Mobile, widget.Part("hamburger"),
			style.Row(style.Space1),
			style.As(style.Primary),
			style.Pad(style.Space2),
			style.Round(style.RadiusSm),
			style.Raise(style.Floating),
			style.CenterContent(),
			style.Docked(style.Viewport, style.EdgeTop, style.SideEnd, style.Space4),
			// Se guarda mientras el usuario baja. Es un estado, no una clase: lo
			// escribe Go y lo lee la hoja, y el atributo sale del propio State para
			// que marcado y selector no puedan discrepar.
			style.RevealedBy(widget.Open),
		).
		OnlyOn(css.Mobile, widget.Part("nav-overlay"),
			style.Backdrop(style.Viewport),
			style.Veil(),
			style.RevealedBy(widget.Open),
		).
		// On a phone the rail stops being a column and becomes a panel that
		// slides in from the edge, gated by the same Open state as the overlay.
		On(css.Mobile, widget.Part("menu"),
			style.Drawer(style.SideEnd, style.TwoThirds),
			style.Stack(style.Space1),
			style.As(style.Panel),
			style.RevealedBy(widget.Open),
		)
}

// RenderCSS implements the visual contract for platformd using the style DSL.
func (p *Platform) RenderCSS() *css.Stylesheet {
	return p.RenderSheet().Stylesheet()
}
