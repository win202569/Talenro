package statetimebinding

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"talenro.local/platform/internal/nodecontrol/statepolicy"
	"talenro.local/platform/internal/nodecontrol/statepolicybinding"
	"talenro.local/platform/internal/nodecontrol/statetime"
	"testing"
	"time"
)

func TestTimeBindingErrorMapping(t *testing.T) {
	pairs := [][2]error{
		{statepolicybinding.ErrInvalidInput, ErrInvalidInput},
		{statepolicybinding.ErrLimitExceeded, ErrLimitExceeded},
		{statepolicybinding.ErrIncompleteCoverage, ErrIncompleteCoverage},
		{statepolicybinding.ErrInvalidMaterial, ErrInvalidMaterial},
		{statepolicybinding.ErrMissingMaterial, ErrMissingMaterial},
		{statepolicybinding.ErrDuplicateMaterial, ErrDuplicateMaterial},
		{statepolicybinding.ErrBindingMismatch, ErrBindingMismatch},
		{statepolicybinding.ErrSignatureMismatch, ErrSignatureMismatch},
		{statepolicybinding.ErrRollback, ErrRollback},
		{statepolicybinding.ErrConflict, ErrConflict},
		{statepolicybinding.ErrClearedModel, ErrClearedModel},
		{statepolicybinding.ErrUnsupportedPolicyShape, ErrUnsupportedPolicyShape},
		{statepolicybinding.ErrPolicyRejected, ErrPolicyRejected},
		{statetime.ErrInvalidInput, ErrInvalidInput},
		{statetime.ErrBindingMismatch, ErrBindingMismatch},
		{statetime.ErrSignatureMismatch, ErrSignatureMismatch},
		{statetime.ErrUnauthorizedKey, ErrUnauthorizedKey},
		{statetime.ErrAssuranceMissing, ErrAssuranceMissing},
		{statetime.ErrInvalidTime, ErrInvalidTime},
		{statetime.ErrExpired, ErrExpired},
		{statetime.ErrCheckpointRollback, ErrCheckpointRollback},
		{statetime.ErrStaleEvidence, ErrStaleEvidence},
		{ErrUnsupportedBindingShape, ErrUnsupportedBindingShape},
	}
	for _, p := range pairs {
		if got := mapError(context.Background(), fmt.Errorf("sensitive: %w", p[0])); got != p[1] || !strings.HasPrefix(got.Error(), "statetimebinding: ") || strings.Contains(got.Error(), "sensitive") {
			t.Fatal("mapping", got)
		}
	}
	for code := 1; code <= 5; code++ {
		for reason := 1; reason <= 14; reason++ {
			low := &statepolicybinding.PolicyFailure{Code: statepolicy.Code(code), Reason: statepolicy.Reason(reason)}
			got := mapError(context.Background(), fmt.Errorf("secret %w", low))
			var high *PolicyFailure
			if !errors.As(got, &high) || high == nil || high.Code != low.Code || high.Reason != low.Reason || !errors.Is(got, ErrPolicyRejected) || high.Error() != ErrPolicyRejected.Error() {
				t.Fatal("diagnostic lost")
			}
		}
	}
	var low *statepolicybinding.PolicyFailure
	for _, e := range []error{low, &statepolicybinding.PolicyFailure{}, &statepolicybinding.PolicyFailure{Code: 6, Reason: 1}, &statepolicybinding.PolicyFailure{Code: 1, Reason: 15}, errors.New("sensitive raw/UUID/path")} {
		if mapError(context.Background(), e) != ErrInvalidInput {
			t.Fatal("unknown diagnosis")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if mapError(ctx, ErrInvalidInput) != context.Canceled {
		t.Fatal("ctx priority")
	}
	if mapError(nil, nil) != ErrInvalidInput || mapError(context.Background(), nil) != nil {
		t.Fatal("nil")
	}
	for _, e := range []error{context.Canceled, context.DeadlineExceeded} {
		if mapError(context.Background(), fmt.Errorf("wrapped: %w", e)) != e {
			t.Fatal("standard cancellation")
		}
	}
}

func TestTimeBindingWrappedNilAndDeadline(t *testing.T) {
	var pf *statepolicybinding.PolicyFailure
	if mapError(context.Background(), fmt.Errorf("wrapped: %w", pf)) != ErrInvalidInput {
		t.Fatal("wrapped typed nil")
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	defer cancel()
	if mapError(ctx, pf) != context.DeadlineExceeded {
		t.Fatal("deadline priority")
	}
}
