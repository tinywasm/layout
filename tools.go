//go:build tools

package layout

// SSR-scanner support: the daemon loads
// github.com/tinywasm/components/calendarslider in this module's context
// although no layout package imports it. Without a real import, go mod tidy
// prunes date's sums and registration fails with "missing go.sum entry for
// module providing package github.com/tinywasm/date". This blank import is
// the standard tools.go mechanism to keep them; it never compiles into any
// binary (tools tag).
import _ "github.com/tinywasm/date"
