package cancelContext

import (
	"context"
	"reflect"
	"sync/atomic"
	"time"
)

type (
	// Wrapping context.WithCancel.
	CancelCtx struct {
		context.Context
		cancelFunc context.CancelFunc
		isDone     int32
	}
)

var (
	ContextDoneError = context.Canceled
	CanceledCtx      *CancelCtx
)

// closedChan is a reusable closed channel.
var closedChan <-chan struct{}

func init() {
	CanceledCtx = NewCancelCtx(context.Background())
	CanceledCtx.Cancel()
	closedChan = CanceledCtx.Context.Done()
}

// returns context.closedchan
func ClosedChan() <-chan struct{} {
	return closedChan
}

func NewCancelCtx(parent context.Context) *CancelCtx {
	c, f := context.WithCancel(parent)
	return &CancelCtx{
		Context:    c,
		cancelFunc: f,
	}
}

// func NewCancelCtx2(parent context.Context) CancelCtx {
// 	c, f := context.WithCancel(parent)
// 	return CancelCtx{
// 		Context:    c,
// 		cancelFunc: f,
// 	}
// }

// func NewCancelCtx2V(parent CancelCtx) CancelCtx {
// 	c, f := context.WithCancel(parent)
// 	return CancelCtx{
// 		Context:    c,
// 		cancelFunc: f,
// 	}
// }

func NewTimeoutCtx(parent context.Context, timeout time.Duration) *CancelCtx {
	c, f := context.WithTimeout(parent, timeout)
	return &CancelCtx{
		Context:    c,
		cancelFunc: f,
	}
}

// func NewTimeoutCtx2(parent context.Context, timeout time.Duration) CancelCtx {
// 	c, f := context.WithTimeout(parent, timeout)
// 	return CancelCtx{
// 		Context:    c,
// 		cancelFunc: f,
// 	}
// }

func (me *CancelCtx) Deadline() (deadline time.Time, ok bool) {
	if me.Context == nil {
		return
	}
	return me.Context.Deadline()
}

func (me *CancelCtx) Done() <-chan struct{} {
	if atomic.LoadInt32(&me.isDone) != 0 {
		return closedChan
	}
	return me.Context.Done()
}

func (me *CancelCtx) Err() error {
	if atomic.LoadInt32(&me.isDone) != 0 {
		return ContextDoneError
	} else {
		return me.Context.Err()
	}
}

func (me *CancelCtx) Value(key interface{}) interface{} {
	if me.Context == nil {
		return nil
	}
	return me.Context.Value(key)
}

// Cancel the context itself.
// returns:
// - true: the context ITSELF was not canceled, even if the parent is canceled
// - false: the context ITSELF was already canceled
func (me *CancelCtx) Cancel() bool {
	if atomic.CompareAndSwapInt32(&me.isDone, 0, 1) {
		me.cancelFunc()
		return true
	}
	return false
}

// Canceled provides another way to check if the context is canceled.
//   - no lock if canceled by itself, or parent canceled before first wait/check
//   - 1 lock if parent was canceled afterwards
//   - 2 locks if not canceled
//
// While Err() always requires 1 lock
func (me *CancelCtx) Canceled() bool {
	if atomic.LoadInt32(&me.isDone) != 0 {
		return true
	}
	ch := me.Done()
	if ch == closedChan {
		return true
	}
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

// func (me CancelCtx) IsEmpty() bool {
// 	return me.Context == nil
// }

// innerCancelCtx returns the heap *cancelCtx address embedded in c.
func innerCancelCtx(c *CancelCtx) uintptr {
	if c.Context == nil {
		panic("cancelContext: internal context is not cancelable")
	}
	return reflect.ValueOf(c.Context).Pointer()
}
