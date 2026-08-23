package contracts

import (
	"errors"
	"math"
	"testing"
)

func TestCompareVersionedDigestClosedOutcomes(t *testing.T) {
	digestA := Digest{1}
	digestB := Digest{2}
	cases := []struct {
		name      string
		current   VersionedDigest
		candidate VersionedDigest
		want      Comparison
		wantErr   error
	}{
		{name: "same", current: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, candidate: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, want: ComparisonSame},
		{name: "advance", current: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, candidate: VersionedDigest{Version: 5, AuthoritySequence: 9, Digest: digestB}, want: ComparisonAdvance},
		{name: "version rollback", current: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, candidate: VersionedDigest{Version: 3, AuthoritySequence: 9, Digest: digestB}, want: ComparisonRollback},
		{name: "sequence rollback", current: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, candidate: VersionedDigest{Version: 5, AuthoritySequence: 7, Digest: digestB}, want: ComparisonRollback},
		{name: "same coordinate fork", current: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, candidate: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestB}, want: ComparisonFork},
		{name: "crossed coordinate fork", current: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, candidate: VersionedDigest{Version: 4, AuthoritySequence: 9, Digest: digestB}, want: ComparisonFork},
		{name: "zero digest", current: VersionedDigest{Version: 4, AuthoritySequence: 8, Digest: digestA}, candidate: VersionedDigest{Version: 5, AuthoritySequence: 9}, wantErr: ErrInvalidAuthorityValue},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := CompareVersionedDigest(test.current, test.candidate)
			if errors.Is(err, test.wantErr) == false || got != test.want {
				t.Fatalf("comparison = %q, %v; want %q, %v", got, err, test.want, test.wantErr)
			}
		})
	}
}

func TestAuthorityVersionValidateRejectsZeroAndValuesAbovePostgresBigint(t *testing.T) {
	cases := []struct {
		name  string
		value AuthorityVersion
		want  error
	}{
		{name: "zero epoch", value: AuthorityVersion{Epoch: 0, Sequence: 1}, want: ErrInvalidAuthorityValue},
		{name: "zero sequence", value: AuthorityVersion{Epoch: 1, Sequence: 0}, want: ErrInvalidAuthorityValue},
		{name: "max epoch accepted", value: AuthorityVersion{Epoch: math.MaxInt64, Sequence: 1}},
		{name: "max sequence accepted", value: AuthorityVersion{Epoch: 1, Sequence: math.MaxInt64}},
		{name: "epoch above max rejected", value: AuthorityVersion{Epoch: math.MaxInt64 + 1, Sequence: 1}, want: ErrInvalidAuthorityValue},
		{name: "sequence above max rejected", value: AuthorityVersion{Epoch: 1, Sequence: math.MaxInt64 + 1}, want: ErrInvalidAuthorityValue},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := test.value.Validate()
			if !errors.Is(err, test.want) {
				t.Fatalf("validation error = %v; want %v", err, test.want)
			}
		})
	}
}

func TestCompareVersionedDigestRejectsInvalidCoordinates(t *testing.T) {
	digest := Digest{1}
	validCurrent := VersionedDigest{Version: 1, AuthoritySequence: 1, Digest: digest}
	validCandidate := VersionedDigest{Version: 2, AuthoritySequence: 2, Digest: digest}
	cases := []struct {
		name      string
		current   VersionedDigest
		candidate VersionedDigest
	}{
		{name: "current zero version", current: VersionedDigest{AuthoritySequence: 1, Digest: digest}, candidate: validCandidate},
		{name: "current zero sequence", current: VersionedDigest{Version: 1, Digest: digest}, candidate: validCandidate},
		{name: "current version above max", current: VersionedDigest{Version: math.MaxInt64 + 1, AuthoritySequence: 1, Digest: digest}, candidate: validCandidate},
		{name: "current sequence above max", current: VersionedDigest{Version: 1, AuthoritySequence: math.MaxInt64 + 1, Digest: digest}, candidate: validCandidate},
		{name: "candidate zero version", current: validCurrent, candidate: VersionedDigest{AuthoritySequence: 2, Digest: digest}},
		{name: "candidate zero sequence", current: validCurrent, candidate: VersionedDigest{Version: 2, Digest: digest}},
		{name: "candidate version above max", current: validCurrent, candidate: VersionedDigest{Version: math.MaxInt64 + 1, AuthoritySequence: 2, Digest: digest}},
		{name: "candidate sequence above max", current: validCurrent, candidate: VersionedDigest{Version: 2, AuthoritySequence: math.MaxInt64 + 1, Digest: digest}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := CompareVersionedDigest(test.current, test.candidate)
			if !errors.Is(err, ErrInvalidAuthorityValue) || got != "" {
				t.Fatalf("comparison = %q, %v; want empty comparison and %v", got, err, ErrInvalidAuthorityValue)
			}
		})
	}
}

func TestCompareVersionedDigestAcceptsMaxPostgresBigintCoordinates(t *testing.T) {
	digest := Digest{1}
	cases := []struct {
		name      string
		current   VersionedDigest
		candidate VersionedDigest
		want      Comparison
	}{
		{name: "current version", current: VersionedDigest{Version: math.MaxInt64, AuthoritySequence: 1, Digest: digest}, candidate: VersionedDigest{Version: math.MaxInt64, AuthoritySequence: 2, Digest: digest}, want: ComparisonFork},
		{name: "current sequence", current: VersionedDigest{Version: 1, AuthoritySequence: math.MaxInt64, Digest: digest}, candidate: VersionedDigest{Version: 2, AuthoritySequence: math.MaxInt64, Digest: digest}, want: ComparisonFork},
		{name: "candidate version", current: VersionedDigest{Version: 1, AuthoritySequence: 1, Digest: digest}, candidate: VersionedDigest{Version: math.MaxInt64, AuthoritySequence: 1, Digest: digest}, want: ComparisonFork},
		{name: "candidate sequence", current: VersionedDigest{Version: 1, AuthoritySequence: 1, Digest: digest}, candidate: VersionedDigest{Version: 1, AuthoritySequence: math.MaxInt64, Digest: digest}, want: ComparisonFork},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := CompareVersionedDigest(test.current, test.candidate)
			if err != nil || got != test.want {
				t.Fatalf("comparison = %q, %v; want %q, nil", got, err, test.want)
			}
		})
	}
}
