//go:build go1.21
// +build go1.21

package cancelContext

import (
	"context"
	"runtime"

	"github.com/szmcdull/go-forceexport"
)

// A canceler is a context type that can be canceled directly. The
// implementations are *cancelCtx and *timerCtx.
type canceler interface {
	cancel(removeFromParent bool, err, cause error)
	Done() <-chan struct{}
}

// innerCanceler is a hashable canceler for map[canceler]struct{} registration.
// Value CancelCtx is unhashable (contains cancelFunc); this wraps only the
// heap *cancelCtx address so NewLinkedCancelCtx2 need not allocate *CancelCtx.
type innerCanceler uintptr

var (
	propagateCancel func(cancelCtx uintptr, parent context.Context, child canceler)
	cancelInner     func(cancelCtx uintptr, removeFromParent bool, err, cause error)
	doneInner       func(cancelCtx uintptr) <-chan struct{}
)

func init() {
	f := runtime.Func{}
	_ = &f

	if err := forceexport.GetFunc(&propagateCancel, `context.(*cancelCtx).propagateCancel`); err != nil {
		panic(err)
	}
	if err := forceexport.GetFunc(&cancelInner, `context.(*cancelCtx).cancel`); err != nil {
		panic(err)
	}
	if err := forceexport.GetFunc(&doneInner, `context.(*cancelCtx).Done`); err != nil {
		panic(err)
	}
}

func (p innerCanceler) cancel(removeFromParent bool, err, cause error) {
	cancelInner(uintptr(p), removeFromParent, err, cause)
}

func (p innerCanceler) Done() <-chan struct{} {
	return doneInner(uintptr(p))
}

func (me CancelCtx) cancel(removeFromParent bool, err, cause error) {
	go me.cancelFunc()
}

// NewLinkedCancelCtx creates a new context that links with all parents.
// When any parent is done, the new context is canceled automatically.
// Only the cancel signal is propagated; linked.Err() is always context.Canceled.
// To inspect why a linked parent canceled, check the original parent contexts directly.
func (parent *CancelCtx) NewLinkedCancelCtx(otherParents ...context.Context) *CancelCtx {
	withCancel := NewCancelCtx(parent)
	p := innerCancelCtx(*withCancel)
	for _, c := range otherParents {
		propagateCancel(p, c, withCancel)
	}

	return withCancel
}

// func (parent CancelCtx) NewLinkedCancelCtx2(otherParents ...context.Context) CancelCtx {
// 	n, f := context.WithCancel(parent)
// 	w := CancelCtx{
// 		Context:    n,
// 		cancelFunc: f,
// 	}

// 	//w := NewCancelCtx2(bg)
// 	p := innerCancelCtx(w)
// 	child := innerCanceler(p)
// 	for _, c := range otherParents {
// 		propagateCancel(p, c, child)
// 	}

// 	return w
// }
