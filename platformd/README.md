# platformd

Reusable, fully-typed UI shell for tinywasm "platform"-style apps: persistent navigation, sliding module panels, integrated notification system.

## Usage

```go
package main

import (
    . "github.com/tinywasm/dom"
    . "github.com/tinywasm/fmt"
    "github.com/tinywasm/layout/platformd"
)

func main() {
    p := &platformd.Platform{
        AppName: "My Platform",
        Modules: []platformd.Module{
            {
                ID: "home", Label: "Home", Default: true,
                Icon: myHomeIcon(),
                View: myHomeView(),
            },
            // ... more modules
        },
    }
    Append("body", p)

    p.Notify(Msg.Info, "Welcome back!")
}
```

## Responsive Behavior

- **Desktop**: Vertical rail fixed to the RIGHT edge; collapsed (icon only), expands on hover. Modules slide in from the left.
- **Mobile**: Horizontal bar fixed at the TOP. Modules slide in from the left as full-screen panels.

## Styling

Styled using `tinywasm/css` with design tokens for easy customization. See `tokens.go` for available tokens.
