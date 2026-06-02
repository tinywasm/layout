# tinywasm/layout Architecture

## Package Layout

    tinywasm/layout/
    ├── platformd/      # Shell: header, nav rail, hash routing, notifications
    │   ├── platformd.go    # Main struct and Render()
    │   ├── css.go          # !wasm: RenderCSS() *css.Stylesheet
    │   ├── svg.go          # !wasm: IconSvg() *svg.Sprite (built-in nav icons)
    │   ├── tokens.go       # !wasm: CSS token definitions
    │   └── web/            # Demo app (wasm binary entry point)
    └── rightpanel/     # Content panel with header/body/aside slots

## Dependencies

    tinywasm/layout → tinywasm/html (element builders)
    tinywasm/layout → tinywasm/svg  (Icon helper, *Sprite)
    tinywasm/layout → tinywasm/dom  (Component, Event, lifecycle)
    tinywasm/layout → tinywasm/css  (Stylesheet, Token)
