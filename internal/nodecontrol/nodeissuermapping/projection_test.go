package nodeissuermapping

import (
	"context"
	"talenro.local/platform/internal/nodecontrol/nodeissuerbinding"
	"testing"
)

func TestIssuerMappingProjection(t *testing.T) {
	in := issuerMappingFixture(t, "serial-one-high")
	s := assertIssuerMappingLower(t, in)
	want := issuerMappingExpected(t, "serial-one-high").Proposed
	if out, e := bindProjection(context.Background(), in.Claims, s); e != nil || out != want {
		t.Fatal(out, e)
	}
	for _, mutate := range []func(*Claims){
		func(c *Claims) { c.Catalog.Entries[0].CADERDigest[0] ^= 1 },
		func(c *Claims) { c.Association.NodeID = "33333333-3333-4333-8333-333333333333" },
		func(c *Claims) { c.Association.LeafDER[0] ^= 1 },
		func(c *Claims) { c.Association.PublicKey[0] ^= 1 },
		func(c *Claims) { c.Association.SerialBytes[0] ^= 1 },
		func(c *Claims) { c.Association.SerialBytes = append(c.Association.SerialBytes, 1) },
	} {
		c := ownedTestClone(t, in.Claims)
		mutate(&c)
		if out, e := bindProjection(context.Background(), c, s); e != ErrClaimConflict || out != (Projection{}) {
			t.Fatal("correspondence", e)
		}
	}
}
func TestIssuerMappingSameKeyRenewal(t *testing.T) {
	one := issuerMappingFixture(t, "serial-one-high")
	renewal := issuerMappingFixture(t, "ca-renewal")
	a, b := assertIssuerMappingLower(t, one), assertIssuerMappingLower(t, renewal)
	c := one.Claims
	r := renewal.Claims.Catalog.Entries[0]
	r.CAVersion = 4
	c.Catalog.Entries = append(c.Catalog.Entries, r)
	if _, e := bindProjection(context.Background(), c, a); e != nil {
		t.Fatal(e)
	}
	if _, e := bindProjection(context.Background(), c, b); e != ErrClaimConflict {
		t.Fatal("wrong explicit version", e)
	}
	c.Association.CAVersion = 4
	if _, e := bindProjection(context.Background(), c, b); e != nil {
		t.Fatal("explicit coherent alternate", e)
	}
	if a.Proposed.IssuerID == b.Proposed.IssuerID || a.Proposed.Serial == b.Proposed.Serial {
		t.Fatal("different exact CA")
	}
}
func TestIssuerMappingSameSerialDifferentIssuer(t *testing.T) {
	one := issuerMappingFixture(t, "serial-one-high")
	other := issuerMappingFixture(t, "other-ca-same-serial")
	a, b := assertIssuerMappingLower(t, one), assertIssuerMappingLower(t, other)
	if _, e := bindProjection(context.Background(), one.Claims, b); e != ErrClaimConflict {
		t.Fatal(e)
	}
	if _, e := bindProjection(context.Background(), other.Claims, b); e != nil {
		t.Fatal(e)
	}
	if a.Proposed.Serial == b.Proposed.Serial {
		t.Fatal("CA-bound serial")
	}
}
func TestIssuerMappingSerialEncoding(t *testing.T) {
	for _, name := range []string{"serial-one-high", "serial-twenty"} {
		in := issuerMappingFixture(t, name)
		s := assertIssuerMappingLower(t, in)
		if _, e := bindProjection(context.Background(), in.Claims, s); e != nil {
			t.Fatal(e)
		}
		for _, prefix := range []bool{false, true} {
			c := ownedTestClone(t, in.Claims)
			if prefix {
				c.Association.SerialBytes = append([]byte{0}, c.Association.SerialBytes...)
			} else {
				c.Association.SerialBytes = append(c.Association.SerialBytes, 0)
			}
			if _, e := bindProjection(context.Background(), c, s); e != ErrClaimConflict {
				t.Fatal("padding", e)
			}
		}
	}
	// Twenty-byte high-bit magnitude is defensive-only; real 21-byte INTEGER content remains rejected.
	s := assertIssuerMappingLower(t, issuerMappingFixture(t, "serial-twenty"))
	s.Evidence.Evidence.Certificate.Leaf.Leaf.Serial[0] = 0x80
	if e := validateEvidence(s); e != nil {
		t.Fatal(e)
	}
	var packets map[string]Input
	issuerMappingDecode(t, issuerMappingFrozenInputs, &packets)
	found := false
	for name, in := range packets {
		if name == "serial-twenty-high-lower-rejected" {
			m, e := nodeissuerbinding.Evaluate(context.Background(), in.Material)
			if m != nil {
				m.Clear()
			}
			if e == nil {
				t.Fatal("21-byte DER accepted")
			}
			found = true
		}
	}
	if !found {
		t.Fatal("missing 21-byte DER control")
	}
}
