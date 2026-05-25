//go:build go1.21
// +build go1.21

package cancelContext

import (
	"context"
	"testing"
)

func countAllocs(n int, fn func()) int {
	return int(testing.AllocsPerRun(n, fn))
}

func TestAllocsReport(t *testing.T) {
	bg := context.Background()
	rootP := NewCancelCtx(bg)
	//rootV := NewCancelCtx2(bg)
	p1P := NewCancelCtx(bg)
	//p1V := NewCancelCtx2(bg)
	// w := NewCancelCtx2(bg)
	wp := NewCancelCtx(bg)

	log := func(name string, fn func()) {
		t.Logf("%-50s %d allocs/op", name, countAllocs(100, fn))
	}

	t.Log("=== create child ctx (parent pre-created) ===")
	log("NewCancelCtx(*CancelCtx parent)", func() { _ = NewCancelCtx(rootP) })
	//log("NewCancelCtx(CancelCtx parent)", func() { _ = NewCancelCtx(rootV) })
	// log("NewCancelCtx2(*CancelCtx parent)", func() { _ = NewCancelCtx2(rootP) })
	//log("NewCancelCtx2(CancelCtx value parent)", func() { _ = NewCancelCtx2(rootV) })

	t.Log("=== linked (parents pre-created) ===")
	log("NewLinkedCancelCtx (both *CancelCtx)", func() { _ = rootP.NewLinkedCancelCtx(p1P) })
	// log("NewLinkedCancelCtx2 (both CancelCtx value)", func() { _ = rootV.NewLinkedCancelCtx2(p1V) })
	// log("NewLinkedCancelCtx2 (* primary, CancelCtx value other)", func() { _ = rootP.NewLinkedCancelCtx2(p1V) })
	// log("NewLinkedCancelCtx2 (CancelCtx value primary, * other)", func() { _ = rootV.NewLinkedCancelCtx2(p1P) })
	// log("NewLinkedCancelCtx2 (both *CancelCtx)", func() { _ = rootP.NewLinkedCancelCtx2(p1P) })

	t.Log("=== innerCancelCtx ===")
	// log("innerCancelCtx(CancelCtx value)", func() { _ = innerCancelCtx(&w) })
	log("innerCancelCtx(*CancelCtx deref)", func() { _ = innerCancelCtx(wp) })
}
