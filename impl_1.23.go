//go:build go1.23
// +build go1.23

package cancelContext

import (
	"context"
)

// NewLinkedCancelCtx creates a new context that links with all parents.
// When any parent is done, the new context is canceled automatically (in a separate goroutine).
func (parent *CancelCtx) NewLinkedCancelCtx(otherParents ...context.Context) *CancelCtx {
	result := NewCancelCtx(parent)

	afterFunc := func() {
		result.Cancel()
	}

	for _, c := range otherParents {
		context.AfterFunc(c, afterFunc)
	}
	context.AfterFunc(parent, afterFunc)

	return result
}
