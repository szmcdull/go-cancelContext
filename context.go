package cancelContext

import (
	"context"
	"time"
)

type (
	// 用event模拟的Context，实验性质，请勿使用
	// Context struct {
	// 	exitEvent *Event
	// }

	// Wrapping context.WithCancel
	// An empty CancelCtx behaves as it is already canceled
	CancelCtx struct {
		context.Context
		cancelFunc context.CancelFunc
		//isDone     int32
	}
)

var (
	ContextDoneError = context.Canceled
	CanceledCtx      = CancelCtx{}
)

// closedChan is a reusable closed channel.
var closedChan <-chan struct{}

func init() {
	c, cancel := context.WithCancel(context.Background())
	cancel()
	closedChan = c.Done()
}

// func NewContext() Context {
// 	return Context{
// 		exitEvent: NewEvent(),
// 	}
// }

// func (c Context) Deadline() (time.Time, bool) {
// 	return time.Time{}, false
// }

// func (c Context) Done() <-chan struct{} {
// 	return c.exitEvent.Done()
// }

// func (c Context) Err() error {
// 	if c.exitEvent.IsSet() {
// 		return ContextDoneError
// 	} else {
// 		return nil
// 	}
// }

// func (c Context) Value(key interface{}) interface{} {
// 	return nil
// }

// func (c Context) Close() {
// 	c.exitEvent.Set()
// }

// func (me *CancelCtx) Err() error {
// 	return me.Context.Err()
// }

// func (me *CancelCtx) Done() <-chan struct{} {
// 	if me.isDone != 0 {
// 		return closedChan
// 	}
// 	return me.Context.Done()
// }

// returns context.closedchan
func ClosedChan() <-chan struct{} {
	return closedChan
}

func NewCancelCtx(parent context.Context) CancelCtx {
	c, f := context.WithCancel(parent)
	return CancelCtx{
		Context:    c,
		cancelFunc: f,
	}
}

func NewTimeoutCtx(parent context.Context, timeout time.Duration) CancelCtx {
	c, f := context.WithTimeout(parent, timeout)
	return CancelCtx{
		Context:    c,
		cancelFunc: f,
	}
}

func (me CancelCtx) Deadline() (deadline time.Time, ok bool) {
	if me.Context == nil {
		return
	}
	return me.Context.Deadline()
}

func (me CancelCtx) Done() <-chan struct{} {
	if me.Context == nil {
		return closedChan
	}
	return me.Context.Done()
}

func (me CancelCtx) Err() error {
	if me.Context == nil {
		return ContextDoneError
	}
	return me.Context.Err()
}

func (me CancelCtx) Value(key interface{}) interface{} {
	if me.Context == nil {
		return nil
	}
	return me.Context.Value(key)
}

func (me CancelCtx) Cancel() {
	// if me.Context.Err() != nil {
	// 	return false
	// }
	me.cancelFunc()
	// return true
}

// Canceled provides another way to check if the context is canceled.
//   - no lock if canceled before first wait/check
//   - 1 lock if canceled afterwards
//   - 2 locks if not canceled
//
// While Err() always requires 1 lock
func (me CancelCtx) Canceled() bool {
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

func (me CancelCtx) IsEmpty() bool {
	return me.Context == nil
}
