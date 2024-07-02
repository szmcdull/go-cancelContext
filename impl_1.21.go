//go:build go1.21 && !go1.23
// +build go1.21,!go1.23

package cancelContext

import (
	"context"
	"reflect"
	"runtime"
	"sync/atomic"

	"github.com/szmcdull/go-forceexport"
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

func init() {
	f := runtime.Func{}
	_ = &f

	if err := forceexport.GetFunc(&propagateCancel, `context.(*cancelCtx).propagateCancel`); err != nil {
		panic(err)
	}
}

func (me *CancelCtx) cancel(removeFromParent bool, err, cause error) {
	if atomic.CompareAndSwapInt32(&me.isDone, 0, 1) {
		me.cancelFunc()
	}
}

// NewLinkedCancelCtx creates a new context that links with all parents.
// When any parent is done, the new context is canceled automatically.
func (parent *CancelCtx) NewLinkedCancelCtx(otherParents ...context.Context) *CancelCtx {
	withCancel := NewCancelCtx(parent)
	p := reflect.ValueOf(withCancel.Context).Pointer()
	for _, c := range otherParents {
		propagateCancel(p, c, withCancel)
	}

	return withCancel
}
