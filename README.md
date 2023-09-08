# go-cancelContext
CancelContext and LinkedCancelCtx inspired by .Net

NewLinkedCancelCtx() creates a new context that links with multiple parents.
When any parent is done, the new context is canceled automatically.

This package makes use of internal go library, so that it is simple and efficient.