package contracts

import (
	"errors"
	"math"
	"testing"
)

func FuzzCompareVersionedDigestNeverInventsOutcome(f *testing.F) {
	f.Add(uint64(1), uint64(1), uint64(1), uint64(1), byte(1), byte(1))
	f.Fuzz(func(t *testing.T, cv, cs, nv, ns uint64, cd, nd byte) {
		current := VersionedDigest{Version: cv, AuthoritySequence: cs, Digest: Digest{cd}}
		candidate := VersionedDigest{Version: nv, AuthoritySequence: ns, Digest: Digest{nd}}
		got, err := CompareVersionedDigest(current, candidate)
		if err != nil {
			return
		}
		switch got {
		case ComparisonSame, ComparisonAdvance, ComparisonRollback, ComparisonFork:
			return
		default:
			t.Fatalf("unknown comparison %q", got)
		}
	})
}

func FuzzCompareLocalVersionedDigestNeverInventsOutcome(f *testing.F) {
	f.Add(uint64(1), uint64(1), byte(1), byte(1))
	f.Add(uint64(1), uint64(2), byte(1), byte(2))
	f.Add(uint64(2), uint64(1), byte(1), byte(2))
	f.Add(uint64(1), uint64(1), byte(1), byte(2))
	f.Add(uint64(0), uint64(1), byte(1), byte(1))
	f.Add(uint64(1), uint64(0), byte(1), byte(1))
	f.Add(uint64(math.MaxInt64+1), uint64(1), byte(1), byte(1))
	f.Add(uint64(1), uint64(math.MaxInt64+1), byte(1), byte(1))
	f.Add(uint64(1), uint64(1), byte(0), byte(1))
	f.Add(uint64(1), uint64(1), byte(1), byte(0))
	f.Fuzz(func(t *testing.T, cv, nv uint64, cd, nd byte) {
		current := LocalVersionedDigestV1{Version: cv, Digest: Digest{cd}}
		candidate := LocalVersionedDigestV1{Version: nv, Digest: Digest{nd}}
		got, err := CompareLocalVersionedDigest(current, candidate)
		validCurrent := cv > 0 && cv <= math.MaxInt64 && cd != 0
		validCandidate := nv > 0 && nv <= math.MaxInt64 && nd != 0
		if !validCurrent || !validCandidate {
			if !errors.Is(err, ErrInvalidAuthorityValue) || got != "" {
				t.Fatalf("invalid comparison = %q, %v; want empty comparison and %v", got, err, ErrInvalidAuthorityValue)
			}
			return
		}

		var want Comparison
		switch {
		case nv < cv:
			want = ComparisonRollback
		case nv > cv:
			want = ComparisonAdvance
		case nd == cd:
			want = ComparisonSame
		default:
			want = ComparisonFork
		}
		if err != nil || got != want {
			t.Fatalf("comparison = %q, %v; want %q, nil", got, err, want)
		}
	})
}
