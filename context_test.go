package cancelContext

import (
	"context"
	"testing"
	"time"
)

func TestWithCancel(t *testing.T) {
	ctx := NewCancelCtx(context.Background())
	ctx.Cancel()

	done := false
	select {
	case <-ctx.Done():
		done = true
	default:
	}

	if !done {
		t.Log(`Should be done`)
		t.Fail()
	}
}

func TestWithCancelParent(t *testing.T) {
	ctx := NewCancelCtx(context.Background())
	ctx2 := NewCancelCtx(ctx)
	ctx.Cancel()

	done := false
	select {
	case <-ctx2.Done():
		done = true
	default:
	}

	if !done {
		t.Log(`Should be done`)
		t.Fail()
	}
}

func TestNotCanceled(t *testing.T) {
	ctx := NewCancelCtx(context.Background())

	done := false
	select {
	case <-ctx.Done():
		done = true
	default:
	}

	if done {
		t.Log(`Should not be done`)
		t.Fail()
	}
}

func TestLinkedCancelCtx1(t *testing.T) {
	ctx1 := NewCancelCtx(context.WithValue(context.Background(), `name`, `ctx1`))
	ctx2 := NewCancelCtx(context.WithValue(context.Background(), `name`, `ctx2`))

	ctx := ctx1.NewLinkedCancelCtx(ctx2)
	ctx1.Cancel()

	done := false
	select {
	case <-ctx.Done():
		done = true
	default:
	}

	if !done {
		t.Log(`Should be done`)
		t.Fail()
	}
}

func TestLinkedCancelCtx2(t *testing.T) {
	ctx1 := NewCancelCtx(context.Background())
	ctx2 := NewCancelCtx(context.Background())

	ctx := ctx1.NewLinkedCancelCtx(ctx2)
	ctx2.Cancel()

	done := false
	time.Sleep(time.Millisecond) // go 1.23 will cancel linked contexts in a separate goroutine
	select {
	case <-ctx.Done():
		done = true
	default:
	}

	if !done {
		t.Log(`Should be done`)
		t.Fail()
	}
}

func TestErr(t *testing.T) {
	c, cancel := context.WithCancel(context.Background())
	c2 := NewCancelCtx(c)
	cancel()
	if c2.Err() == nil {
		t.Errorf(`err is nil`)
	}
}

func waitDone(c *CancelCtx, result *bool) {
	<-c.Done()
	*result = true
}

func TestDoneAfterCancel(t *testing.T) {
	c := NewCancelCtx(context.Background())
	c.Cancel()
	done := false
	go waitDone(c, &done)
	time.Sleep(time.Millisecond) // go will cancel linked contexts in a separate goroutine
	if !done {
		t.Fail()
	}
}

func TestPassOn(t *testing.T) {
	c := NewCancelCtx(context.Background())

	cancel := func(ctx *CancelCtx) {
		ctx.Cancel()
	}

	cancel(c)
	if c.Err() == nil {
		t.Log(`Should be canceled`)
		t.Fail()
	}

	c = NewCancelCtx(context.Background())
	c2 := NewCancelCtx(context.Background())
	c3 := c.NewLinkedCancelCtx(c2)
	cancel(c2)
	time.Sleep(time.Millisecond) // seems the child context is canceled in a goroutine (see context.afterFuncCtx.cancel()), so wait a second for it
	if c3.Err() == nil {
		t.Log(`Should be canceled`)
		t.Fail()
	}
}

func TestCanceledCtx(t *testing.T) {
	c := NewCancelCtx(CanceledCtx)
	if c.Err() != context.Canceled {
		t.Errorf(`expected ContextDoneError, got %v`, c.Err())
	}
	c = NewCancelCtx(context.Background())
	c2 := c.NewLinkedCancelCtx(CanceledCtx)
	time.Sleep(time.Millisecond)
	if c2.Err() != context.Canceled {
		t.Errorf(`expected ContextDoneError, got %v`, c2.Err())
	}
	// if !CanceledCtx.IsEmpty() {
	// 	t.Errorf(`expected CancelCtx to be empty`)
	// }
}

func TestCloseChan(t *testing.T) {
	c := NewCancelCtx(context.Background())
	c.Cancel()
	if c.Done() != ClosedChan() {
		t.Fail()
	}
}

func TestCancelIdempotent(t *testing.T) {
	ctx := NewCancelCtx(context.Background())
	if !ctx.Cancel() {
		t.Error(`first Cancel() should return true`)
	}
	if ctx.Cancel() {
		t.Error(`second Cancel() should return false`)
	}
	if ctx.Cancel() {
		t.Error(`third Cancel() should return false`)
	}
}

func TestCancelAfterParentCanceled(t *testing.T) {
	parent := NewCancelCtx(context.Background())
	child := NewCancelCtx(parent)
	parent.Cancel()

	if !child.Cancel() {
		t.Error(`first Cancel() on child should return true when only parent was canceled`)
	}
	if child.Cancel() {
		t.Error(`second Cancel() on child should return false`)
	}
}

func TestDoneFastPathAfterCancel(t *testing.T) {
	ctx := NewCancelCtx(context.Background())
	if ctx.Done() == ClosedChan() {
		t.Error(`Done() must not be ClosedChan() before explicit Cancel()`)
	}
	ctx.Cancel()
	if ctx.Done() != ClosedChan() {
		t.Error(`Done() must be ClosedChan() after explicit Cancel() (isDone fast path)`)
	}
}

func TestIsDoneUnsetWhenOnlyParentCanceled(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	child := NewCancelCtx(parent)
	cancelParent()

	// Any canceled context's Done() may equal ClosedChan() (stdlib shares one
	// closed channel). isDone is verified via Cancel() returning true once.
	if !child.Cancel() {
		t.Error(`Cancel() should return true when only parent was canceled (isDone not set)`)
	}
	if child.Cancel() {
		t.Error(`second Cancel() should return false after isDone was set`)
	}
	select {
	case <-child.Done():
	default:
		t.Error(`child should be done when parent was canceled`)
	}
}

func TestErrAfterExplicitCancel(t *testing.T) {
	timeoutParent, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond)
	if timeoutParent.Err() != context.DeadlineExceeded {
		t.Fatalf(`timeout parent err = %v, want DeadlineExceeded`, timeoutParent.Err())
	}

	child := NewCancelCtx(timeoutParent)
	if child.Err() != context.DeadlineExceeded {
		t.Errorf(`before self Cancel(), err = %v, want DeadlineExceeded`, child.Err())
	}
	child.Cancel()
	if child.Err() != context.Canceled {
		t.Errorf(`after self Cancel(), err = %v, want Canceled`, child.Err())
	}
	if child.Err() != ContextDoneError {
		t.Errorf(`after self Cancel(), err should be ContextDoneError`)
	}
}

func TestErrExplicitCancelAfterParentCanceled(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	child := NewCancelCtx(parent)
	cancelParent()
	if child.Err() != context.Canceled {
		t.Fatalf(`child err from parent cancel = %v, want Canceled`, child.Err())
	}
	if !child.Cancel() {
		t.Error(`first Cancel() should return true even when parent already canceled`)
	}
	if child.Err() != context.Canceled {
		t.Errorf(`after self Cancel(), err = %v, want Canceled (not parent-specific type)`, child.Err())
	}
}

func TestStdWithCancel(t *testing.T) {
	c, cancel := context.WithCancel(context.Background())
	if c.Err() != nil {
		t.Fatal(`initial Err() should be nil`)
	}
	cancel()
	err := c.Err()
	if err != context.Canceled {
		t.Errorf(`expected context.Canceled, got %v`, err)
	}
	cancel()
	err = c.Err()
	if err != context.Canceled {
		t.Errorf(`expected context.Canceled, got %v`, err)
	}
}

// Getting real cause from linked context is not supported.
// context.Canceled is returned instead.

// func waitLinkedDone(t *testing.T, ctx *CancelCtx) {
// 	t.Helper()
// 	select {
// 	case <-ctx.Done():
// 	case <-time.After(time.Second):
// 		t.Fatal(`linked context was not canceled in time`)
// 	}
// }

// // Linked cancel propagates the signal only, not the linked parent's Err()/Cause().
// func TestLinkedCancelCtxSignalOnly(t *testing.T) {
// 	timeoutParent, cancel := context.WithTimeout(context.Background(), time.Millisecond)
// 	defer cancel()
// 	time.Sleep(2 * time.Millisecond)
// 	if timeoutParent.Err() != context.DeadlineExceeded {
// 		t.Fatalf(`timeout parent err = %v, want DeadlineExceeded`, timeoutParent.Err())
// 	}

// 	root := NewCancelCtx(context.Background())
// 	linked := root.NewLinkedCancelCtx(timeoutParent)
// 	waitLinkedDone(t, linked)
// 	if linked.Err() != context.Canceled {
// 		t.Errorf(`linked err = %v, want Canceled`, linked.Err())
// 	}

// 	cause := errors.New(`boom`)
// 	causeParent, cancelCause := context.WithCancelCause(context.Background())
// 	cancelCause(cause)
// 	if context.Cause(causeParent) != cause {
// 		t.Fatalf(`parent cause = %v, want %v`, context.Cause(causeParent), cause)
// 	}

// 	root = NewCancelCtx(context.Background())
// 	linked = root.NewLinkedCancelCtx(causeParent)
// 	waitLinkedDone(t, linked)
// 	if linked.Err() != context.Canceled {
// 		t.Errorf(`linked err = %v, want Canceled`, linked.Err())
// 	}
// 	if context.Cause(linked) == cause {
// 		t.Errorf(`linked cause should not be propagated from linked parent`)
// 	}
// }
