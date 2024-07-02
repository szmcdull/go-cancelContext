package cancelContext

import (
	"context"
	"sync/atomic"
	"time"
)

type (
	// 用event模拟的Context，实验性质，请勿使用
	// Context struct {
	// 	exitEvent *Event
	// }

	// 封装标准库context.WithCancel
	CancelCtx struct {
		context.Context
		cancelFunc context.CancelFunc
		isDone     int32
	}
)

var (
	ContextDoneError = context.Canceled
)

// closedChan is a reusable closed channel.
var closedChan = make(chan struct{})

func init() {
	close(closedChan)
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

func (me *CancelCtx) Cancel() bool {
	if atomic.CompareAndSwapInt32(&me.isDone, 0, 1) {
		me.cancelFunc()
		return true
	}
	return false
}

func (me *CancelCtx) Err() error {
	if me.isDone != 0 {
		return ContextDoneError
	} else {
		return me.Context.Err()
	}
}

func (me *CancelCtx) Done() <-chan struct{} {
	if me.isDone != 0 {
		return closedChan
	}
	return me.Context.Done()
}

func ClosedChan() chan struct{} {
	return closedChan
}

func NewCancelCtx(parent context.Context) *CancelCtx {
	c, f := context.WithCancel(parent)
	return &CancelCtx{
		Context:    c,
		cancelFunc: f,
	}
}

func NewTimeoutCtx(parent context.Context, timeout time.Duration) *CancelCtx {
	c, f := context.WithTimeout(parent, timeout)
	return &CancelCtx{
		Context:    c,
		cancelFunc: f,
	}
}
