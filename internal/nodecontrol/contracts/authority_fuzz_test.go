package contracts

import "testing"

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
