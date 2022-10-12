package app

import (
	"context"
	"errors"
	"testing"
)

func TestError_WithReqID(t *testing.T) {
	e1 := ErrNotFound
	e2 := ErrNotFound.WithReqID(context.Background())
	if !errors.Is(e2, e1) {
		t.Error("errors should be equal")
	}
}
