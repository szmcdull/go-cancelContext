# go-cancelContext

CancelContext and LinkedCancelCtx inspired by .NET `CancellationTokenSource`.

## NewLinkedCancelCtx

`NewLinkedCancelCtx()` creates a new context linked to multiple parents (similar to `CancellationTokenSource.CreateLinkedTokenSource`). When **any** parent is done, the linked context is canceled automatically.

### Cancel signal only

Like .NET's linked cancellation token, `NewLinkedCancelCtx` only guarantees that you receive a **cancel notification** (`Done()` closes, `Canceled()` returns true).

It does **not** propagate the linked parent's specific cancel reason:

- `linked.Err()` is always `context.Canceled`, even if a linked parent timed out (`context.DeadlineExceeded`) or was canceled with a custom cause (`context.WithCancelCause`).
- `context.Cause(linked)` is not preserved from linked parents.

This differs from a direct child created with `NewCancelCtx(parent)`, which follows the standard Go context tree and preserves `Err()` / `Cause()`.

If you need to know **why** or **which parent** triggered the cancel, keep references to the original parents and inspect them after `<-linked.Done()` (for example, check each parent's `Err()` or `context.Cause()`).

## Implementation notes

Go 1.18–1.22 use internal `context` APIs via [go-forceexport](https://github.com/szmcdull/go-forceexport) for efficient propagation.

Go 1.23+ uses `context.AfterFunc` as a workaround ([golang/go#67401](https://github.com/golang/go/issues/67401)); see `impl_1.23.go` for performance caveats.
