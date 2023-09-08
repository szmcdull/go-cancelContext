//go:build go1.18 && !go1.21
// +build go1.18,!go1.21

package cancelContext

import (
	"context"
	"runtime"
	"sync/atomic"

	"github.com/szmcdull/go-forceexport"
)

type (
	canceler interface {
		cancel(removeFromParent bool, err error)
		Done() <-chan struct{}
	}
)

var (
	//newCancelCtx    func(parent context.Context) cancelCtx
	propagateCancel func(parent context.Context, child canceler)
)

func init() {
	// context.WithCancel(context.Background())
	// if err := forceexport.GetFunc(&newCancelCtx, `context.newCancelCtx`); err != nil {
	// 	panic(err)
	// }
	f := runtime.Func{}
	_ = &f

	if err := forceexport.GetFunc(&propagateCancel, `context.propagateCancel`); err != nil {
		panic(err)
	}
}

func (me *CancelCtx) cancel(removeFromParent bool, err error) {
	if atomic.CompareAndSwapInt32(&me.isDone, 0, 1) {
		me.cancelFunc()
	}
}

// 将多个Context聚合在一起，任意一个parent Done，聚合Context都会Done
func (parent *CancelCtx) NewLinkedCancelCtx(otherParents ...context.Context) *CancelCtx {
	count := len(otherParents)
	// if count == 0 {
	// 	panic(`at least 1 ctx expected`)
	// }

	withCancel := NewCancelCtx(parent)
	for i := 0; i < count; i++ {
		propagateCancel(otherParents[i], withCancel)
	}

	return withCancel
}
