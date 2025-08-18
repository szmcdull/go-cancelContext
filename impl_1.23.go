//go:build go1.23
// +build go1.23

package cancelContext

import (
	"context"
)

// A canceler is a context type that can be canceled directly. The
// implementations are *cancelCtx and *timerCtx.
type canceler interface {
	cancel(removeFromParent bool, err, cause error)
	Done() <-chan struct{}
}

var (
	propagateCancel func(cancelCtx uintptr, parent context.Context, child canceler)
)

// func init() {
// 	f := runtime.Func{}
// 	_ = &f

// 	if err := forceexport.GetFunc(&propagateCancel, `context.(*cancelCtx).propagateCancel`); err != nil {
// 		panic(err)
// 	}
// }

// NewLinkedCancelCtx creates a new context that links with all parents.
// When any parent is done, the new context is canceled automatically (in a separate goroutine).
//
//   - Performance WARNING 1: each parent context that is not a CancelCtx or created with WithCancel without
//     being wrapped in a custom implementation providing a different done channel, will be monitored in a new goroutine.
//
//   - Performance WARNING 2: when any parent context is canceled, the cancel func is call in a new goroutine.
//     This is due to go 1.23 locked down the use of go:linkname (see https://github.com/golang/go/issues/67401),
//     and we use context.AfterFunc as a workaround.
//
// A new version of go-forceexport supports go 1.23+ but requires setting additional build flags, or spending
// additional 3-4 second boot time searching for FirstModuleData, both are not acceptable in some situations.
func (parent *CancelCtx) NewLinkedCancelCtx(otherParents ...context.Context) CancelCtx {
	result := NewCancelCtx(parent)

	cancel := func() {
		result.Cancel()
	}

	for _, c := range otherParents {
		context.AfterFunc(c, cancel)
	}
	context.AfterFunc(parent, cancel)

	return result
}
