package cancelContext

import (
	"context"
	"testing"
)

func BenchmarkCreateLinked1Parent(b *testing.B) {
	root := NewCancelCtx(context.Background())
	p1 := NewCancelCtx(context.Background())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = root.NewLinkedCancelCtx(p1)
	}
}

func BenchmarkCreateLinked3Parents(b *testing.B) {
	root := NewCancelCtx(context.Background())
	ps := []*CancelCtx{
		NewCancelCtx(context.Background()),
		NewCancelCtx(context.Background()),
		NewCancelCtx(context.Background()),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = root.NewLinkedCancelCtx(ps[0], ps[1], ps[2])
	}
}

func BenchmarkCancelViaLinkedParent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		root := NewCancelCtx(context.Background())
		p1 := NewCancelCtx(context.Background())
		linked := root.NewLinkedCancelCtx(p1)
		p1.Cancel()
		<-linked.Done()
	}
}

func BenchmarkCancelViaTreeParent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		root := NewCancelCtx(context.Background())
		child := NewCancelCtx(root)
		root.Cancel()
		<-child.Done()
	}
}
