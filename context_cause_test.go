//go:build go1.20
// +build go1.20

package cancelContext

import (
	"context"
	"errors"
	"testing"
)

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
