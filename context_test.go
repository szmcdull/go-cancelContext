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
	time.Sleep(500)
	if !done {
		t.Fail()
	}
}
