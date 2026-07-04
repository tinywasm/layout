// Package layout provides pre-built, consistent, reusable UI layout skeletons for
// tinywasm modules. A layout defines WHERE elements are placed (structure, grid, spacing)
// but NOT what those elements contain — that is the consumer's responsibility.
package layout

// Module is the identity contract for the ecosystem.
// Any component that can be identified by a slug (for routing or keys)
// implements this interface.
type Module interface {
	ModelName() string
}
