package nodeissuer

import (
	"bytes"
	"context"
	"testing"
)

// Catches checking declaration variables instead of actual CA bytes, normalizing
// Names, ignoring identifiers, or accidentally binding the CA parent AKI.
func TestIssuerRelation(t *testing.T) {
	in := fixtureInput(t)
	s, e := evaluateLower(context.Background(), in)
	if e != nil {
		t.Fatal(e)
	}
	certs, e := parseSnapshots(context.Background(), in, s)
	if e != nil {
		t.Fatal(e)
	}
	if e = checkIssuerRelation(in, s, certs); e != nil {
		t.Fatal(e)
	}
	badIn := in
	badIn.CA.DeclaredSPKI = in.Leaf.ExpectedSPKI
	if checkIssuerRelation(badIn, s, certs) != ErrIssuerRelationMismatch {
		t.Fatal("actual key")
	}
	for i := 0; i < 6; i++ {
		ca, leaf := *certs.ca, *certs.leaf
		bad := s
		switch i {
		case 0:
			leaf.RawIssuer = []byte{0x30, 0}
		case 1:
			leaf.AuthorityKeyId = make([]byte, 32)
		case 2:
			ca.SubjectKeyId = make([]byte, 32)
		case 3:
			bad.ca.SubjectKeyIdentifier[0] ^= 1
		case 4:
			bad.leaf.AuthorityKeyIdentifier[0] ^= 1
		case 5:
			ca.SubjectKeyId = bytes.Repeat([]byte{1}, 31)
		}
		if checkIssuerRelation(in, bad, certificateSnapshots{ca: &ca, leaf: &leaf}) != ErrIssuerRelationMismatch {
			t.Fatal("relation", i)
		}
	}
	ca := *certs.ca
	ca.AuthorityKeyId = make([]byte, 32)
	if checkIssuerRelation(in, s, certificateSnapshots{ca: &ca, leaf: certs.leaf}) != nil {
		t.Fatal("CA parent AKI compared to subject")
	}
	leaf := *certs.leaf
	leaf.RawIssuer = []byte{0x30, 0}
	leaf.AuthorityKeyId = make([]byte, 32)
	if checkIssuerRelation(in, s, certificateSnapshots{ca: certs.ca, leaf: &leaf}) != ErrIssuerRelationMismatch {
		t.Fatal("Name plus ID")
	}
}
