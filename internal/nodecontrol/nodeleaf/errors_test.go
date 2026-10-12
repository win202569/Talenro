package nodeleaf

import (
	"errors"
	"fmt"
	"testing"
)

func TestLeafFixedErrors(t *testing.T) {
	cases := []struct {
		err  error
		text string
	}{
		{ErrInvalidInput, "node leaf: invalid input"},
		{ErrLimitExceeded, "node leaf: limit exceeded"},
		{ErrInvalidMaterial, "node leaf: invalid material"},
		{ErrUnsupportedLeafProfile, "node leaf: unsupported leaf profile"},
		{ErrBindingMismatch, "node leaf: binding mismatch"},
	}
	for i, c := range cases {
		if c.err == nil || c.err.Error() != c.text || !errors.Is(fmt.Errorf("outer: %w", c.err), c.err) {
			t.Fatal("opaque fixed error", i)
		}
		for j, d := range cases {
			if i != j && errors.Is(c.err, d.err) {
				t.Fatal("collapsed errors", i, j)
			}
		}
	}
}
