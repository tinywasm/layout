# platformd

Reusable, fully-typed UI shell for tinywasm "platform"-style apps: persistent navigation, sliding module panels, integrated notification system.

## Usage

```go
package main

import (
    . "github.com/tinywasm/dom"
    . "github.com/tinywasm/fmt"
    "github.com/tinywasm/layout/platformd"
    "github.com/tinywasm/layout/rightpanel"
)

func main() {
    p := &platformd.Platform{
        AppName: "My Platform",
        UserBlock: Span().Text("User Name"),
        Modules: []platformd.Module{
            {
                ID: "home", Label: "Home", Default: true,
                Icon: myHomeIcon(),
                View: rightpanel.New(
                    "Home",
                    Div().Text("Main Content"),
                    Div().Text("Aside Content"),
                ),
            },
            {
                ID: "settings", Label: "Settings",
                Icon: mySettingsIcon(),
                View: mySettingsView(),
            },
        },
    }
    Append("body", p)

    // Notify with 3s auto-dismiss
    p.Notify(Msg.Info, "Welcome back!", 3000)

    // Persistent error
    p.Notify(Msg.Error, "Connection lost", 0)
}
```

## Responsive Behavior

- **Desktop**: Vertical rail fixed to the RIGHT edge; collapsed (icon only), expands on hover. Modules slide in from the left.
- **Mobile**: Horizontal bar fixed at the TOP. Modules slide in from the left as full-screen panels.

## Styling

Styled using `tinywasm/css` with design tokens for easy customization.

### CSS Tokens

| Token | Default | Description |
| --- | --- | --- |
| `--pd-font-size-normal` | `1.1rem` | Primary font size |
| `--pd-font-size-small` | `.6rem` | Secondary/small font size |
| `--pd-color-primary` | `#ffffff` | Primary text/foreground color |
| `--pd-color-secondary` | `#3f88bf` | Brand/secondary color |
| `--pd-color-tertiary` | `#c2c1c1` | Tertiary/accent color |
| `--pd-color-quaternary` | `#000000` | Background/contrast color |
| `--pd-color-gray` | `#e9e9e9` | Neutral gray color |
| `--pd-color-selection` | `#ff9300` | Active/selection color |
| `--pd-color-hover` | `#ff95008e` | Hover state color |
| `--pd-color-success` | `#aadaff7c` | Success message color |
| `--pd-color-error` | `#f20707` | Error message color |
| `--pd-menu-size` | `6vh` | Menu bar width (desktop) or height (mobile) |
| `--pd-header-height` | `5vh` | Header bar height |
| `--pd-content-height` | `94vh` | Main content area height |
| `--pd-content-width` | `100vw` | Main content area width |
| `--pd-slide-duration` | `0.6s` | Panel transition duration |
| `--pd-transition-wait` | `0s` | Delay before transitions |
