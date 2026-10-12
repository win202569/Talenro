package nodeissuer

import (
	"bytes"
	"context"
	"testing"
	"time"
)

// Catches every material-description mismatch, including padding/tail bytes and
// pathLen's absent/explicit-zero distinction. No public fabricated-summary input.
func TestIssuerSnapshots(t *testing.T) {
	for name := range fixtureGoldens(t) {
		in := fixtureVariant(t, name)
		s, e := evaluateLower(context.Background(), in)
		if e != nil {
			t.Fatal(e)
		}
		c, e := parseSnapshots(context.Background(), in, s)
		if e != nil || c.ca == nil || c.leaf == nil {
			t.Fatal("snapshot positive", name, e)
		}
		if !bytes.Equal(c.ca.Raw, in.CA.DER) || !bytes.Equal(c.leaf.Raw, in.Leaf.DER) {
			t.Fatal("raw ownership")
		}
	}
	in := fixtureInput(t)
	s, e := evaluateLower(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	mutations := []func(*lowerSummaries){
		func(s *lowerSummaries) { s.ca.CADERDigest[0] ^= 1 }, func(s *lowerSummaries) { s.ca.SPKIDERDigest[0] ^= 1 }, func(s *lowerSummaries) { s.ca.SubjectDERDigest[0] ^= 1 }, func(s *lowerSummaries) { s.ca.IssuerDERDigest[0] ^= 1 },
		func(s *lowerSummaries) { s.ca.SubjectKeyIdentifier[0] ^= 1 }, func(s *lowerSummaries) { s.ca.AuthorityKeyIdentifier[0] ^= 1 }, func(s *lowerSummaries) { s.ca.Serial[0] ^= 1 }, func(s *lowerSummaries) { s.ca.Serial[19] = 1 }, func(s *lowerSummaries) { s.ca.SerialLength++ },
		func(s *lowerSummaries) { s.ca.NotBefore = s.ca.NotBefore.Add(time.Second) }, func(s *lowerSummaries) { s.ca.NotAfter = s.ca.NotAfter.Add(time.Second) }, func(s *lowerSummaries) { s.ca.HasPathLen = true }, func(s *lowerSummaries) { s.ca.PathLen = 1 }, func(s *lowerSummaries) { s.ca.CRLSignAllowed = true },
		func(s *lowerSummaries) { s.leaf.Leaf.LeafDERDigest[0] ^= 1 }, func(s *lowerSummaries) { s.leaf.Leaf.SPKIDERDigest[0] ^= 1 }, func(s *lowerSummaries) { s.leaf.DeclaredIssuerSPKIDERDigest[0] ^= 1 },
		func(s *lowerSummaries) { s.leaf.SubjectKeyIdentifier[0] ^= 1 }, func(s *lowerSummaries) { s.leaf.AuthorityKeyIdentifier[0] ^= 1 }, func(s *lowerSummaries) { s.leaf.Leaf.Serial[0] ^= 1 }, func(s *lowerSummaries) { s.leaf.Leaf.Serial[19] = 1 }, func(s *lowerSummaries) { s.leaf.Leaf.SerialLength++ },
		func(s *lowerSummaries) { s.leaf.Leaf.NotBefore = s.leaf.Leaf.NotBefore.Add(time.Second) }, func(s *lowerSummaries) { s.leaf.Leaf.NotAfter = s.leaf.Leaf.NotAfter.Add(time.Second) },
		func(s *lowerSummaries) { s.leaf.Leaf.NodeID[0] ^= 1 }, func(s *lowerSummaries) { s.leaf.Leaf.TrustDomain = "other.test" },
	}
	for i, m := range mutations {
		bad := s
		m(&bad)
		c, e := parseSnapshots(context.Background(), in, bad)
		if e != ErrInvalidMaterial || c != (certificateSnapshots{}) {
			t.Errorf("mismatch %d: %v", i, e)
		}
	}
	malformed := in
	malformed.CA.DER = []byte{0x30, 0}
	if _, e := parseSnapshots(context.Background(), malformed, s); e != ErrInvalidMaterial {
		t.Fatal("malformed CA", e)
	}
	malformed = in
	malformed.Leaf.DER = []byte{0x30, 0}
	if _, e := parseSnapshots(context.Background(), malformed, s); e != ErrInvalidMaterial {
		t.Fatal("malformed leaf", e)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := parseSnapshots(canceled, in, s); e != context.Canceled {
		t.Fatal("snapshot cancellation", e)
	}
}
