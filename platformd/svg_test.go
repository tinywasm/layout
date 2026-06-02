//go:build !wasm

package platformd_test

import (
    "strings"
    "testing"
    "github.com/tinywasm/layout/platformd"
)

func TestPlatform_IconSvg_HasRequiredIcons(t *testing.T) {
    p := &platformd.Platform{}
    sprite := p.IconSvg()
    if sprite == nil { t.Fatal("IconSvg() returned nil") }

    s := sprite.String()
    required := []string{"icon-home", "icon-products", "icon-info"}
    for _, id := range required {
        if !strings.Contains(s, id) {
            t.Errorf("missing icon: %s", id)
        }
    }
}

func TestPlatform_IconSvg_HasCurrentColor(t *testing.T) {
    p := &platformd.Platform{}
    sprite := p.IconSvg()
    s := sprite.String()
    if !strings.Contains(s, "currentColor") {
        t.Error("icons must use fill=currentColor or stroke=currentColor")
    }
}
