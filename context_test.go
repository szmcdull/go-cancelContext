package cancelContext

import (
	"context"
	"errors"
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

func waitDone(c CancelCtx, result *bool) {
	<-c.Done()
	*result = true
}

func TestDoneAfterCancel(t *testing.T) {
	c := NewCancelCtx(context.Background())
	c.Cancel()
	done := false
	go waitDone(c, &done)
	time.Sleep(500)
	if !done {
		t.Fail()
	}
}

func TestPassOn(t *testing.T) {
	c := NewCancelCtx(context.Background())

	cancel := func(ctx CancelCtx) {
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
		t.Errorf(`expected ContextDoneError, got %v`, c.Err())
	}
}

func TestCloseChan(t *testing.T) {
	c := NewCancelCtx(context.Background())
	c.Cancel()
	if c.Done() != ClosedChan() {
		t.Fail()
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

func TestStdWithCancelCause(t *testing.T) {
	c, cancel := context.WithCancelCause(context.Background())
	if c.Err() != nil {
		t.Fatal(`initial Err() should be nil`)
	}
	aaa := errors.New(`aaa`)
	cancel(aaa)
	err := c.Err()
	if err != context.Canceled {
		t.Errorf(`expected context.Canceled, got %v`, err)
	}
	if context.Cause(c) != aaa {
		t.Errorf(`expected cause %v, got %v`, aaa, err)
	}
	cancel(errors.New(`bbb`))
	if context.Cause(c) != aaa {
		t.Errorf(`cause should not changed after set`)
	}
}
