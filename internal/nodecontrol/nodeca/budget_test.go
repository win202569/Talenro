package nodeca

import (
	"context"
	"errors"
	"testing"
)

func TestCAPreflight(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	expired, cancelDeadline := context.WithDeadline(context.Background(), epochTime())
	defer cancelDeadline()
	cases := []struct {
		name      string
		ctx       context.Context
		der, spki int
		want      error
	}{
		{"nil", nil, 65537, 8193, ErrInvalidInput}, {"canceled", canceled, 65537, 8193, context.Canceled},
		{"deadline", expired, 1, 1, context.DeadlineExceeded}, {"DERempty", context.Background(), 0, 1, ErrInvalidInput},
		{"SPKIempty", context.Background(), 1, 0, ErrInvalidInput}, {"one", context.Background(), 1, 1, nil},
		{"maxDER", context.Background(), 65536, 1, nil}, {"overDER", context.Background(), 65537, 1, ErrLimitExceeded},
		{"maxSPKI", context.Background(), 1, 8192, nil}, {"overSPKI", context.Background(), 1, 8193, ErrLimitExceeded},
		{"bothMax", context.Background(), 65536, 8192, nil}, {"DERoverEmptySPKI", context.Background(), 65537, 0, ErrLimitExceeded},
		{"SPKIoverEmptyDER", context.Background(), 0, 8193, ErrLimitExceeded},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := preflight(c.ctx, Input{DER: make([]byte, c.der), DeclaredSPKI: make([]byte, c.spki)}); !errors.Is(got, c.want) {
				t.Fatalf("got %v want %v", got, c.want)
			}
		})
	}
}

func TestCAFixedErrors(t *testing.T) {
	errs := []error{ErrInvalidInput, ErrLimitExceeded, ErrInvalidMaterial, ErrUnsupportedCAProfile, ErrBindingMismatch}
	wants := []string{"nodeca: invalid input", "nodeca: limit exceeded", "nodeca: invalid material", "nodeca: unsupported CA profile", "nodeca: binding mismatch"}
	for i, e := range errs {
		if e == nil || e.Error() != wants[i] {
			t.Fatalf("error %d: %v", i, e)
		}
		for j := 0; j < i; j++ {
			if errors.Is(e, errs[j]) {
				t.Fatal("sentinels aliased")
			}
		}
	}
}
