package nodeissuerbinding

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/statecontract"
	"testing"
)

func issuerBindingMismatch(t testing.TB, in Input) {
	t.Helper()
	assertIssuerBindingLower(t, in)
	before := issuerBindingJSONHash(t, in)
	m, e := Evaluate(context.Background(), in)
	if m != nil || e != ErrBindingMismatch {
		t.Fatal("reference mismatch", e)
	}
	if issuerBindingJSONHash(t, in) != before {
		t.Fatal("input repaired")
	}
}
func TestIssuerBindingProjection(t *testing.T) {
	in := issuerBindingFixture(t, "base")
	evidence := assertIssuerBindingLower(t, in)
	p, e := bindProjection(context.Background(), in, evidence)
	want := issuerBindingExpected(t, "base")
	if e != nil || p != (Projection{Leaf: evidence.Proposed, IssuerID: want.Issuer, Serial: want.Serial}) {
		t.Fatal("projection", e)
	}
}
func TestIssuerBindingProjectionControls(t *testing.T) {
	for _, mutate := range []func(*statecontract.IdentitySection){
		func(i *statecontract.IdentitySection) { i.IssuerID[0] ^= 0x80 },
		func(i *statecontract.IdentitySection) { i.Serial[0] ^= 1 },
		func(i *statecontract.IdentitySection) { i.IssuerID[0] ^= 0x80; i.Serial[0] ^= 1 },
	} {
		in := issuerBindingFixture(t, "base")
		mutate(&in.Time.Policy.Binding.Main.Previous.Identity)
		mutate(&in.Time.Policy.Binding.Main.Proposed.Identity)
		issuerBindingMismatch(t, in)
	}
	for _, mutate := range []func(*Input){
		func(in *Input) {
			in.Time.Policy.Binding.Main.Previous.Identity.IssuerID = contracts.Digest{}
			in.Time.Policy.Binding.Main.Proposed.Identity.IssuerID = contracts.Digest{}
		},
		func(in *Input) { in.Time.Policy.Binding.Main.Proposed.Identity.Serial[0] ^= 1 },
	} {
		in := issuerBindingFixture(t, "base")
		mutate(&in)
		m, e := Evaluate(context.Background(), in)
		if m != nil || e == nil || e == ErrBindingMismatch {
			t.Fatal("lower prerequisite rejection", e)
		}
	}
}
func TestIssuerBindingSameKeyDifferentCAReferences(t *testing.T) {
	old := issuerBindingFixture(t, "serial-one-high")
	renewal := issuerBindingFixture(t, "ca-renewal")
	assertIssuerBindingLower(t, old)
	assertIssuerBindingLower(t, renewal)
	renewal.Time.Policy.Binding.Main.Previous.Identity.IssuerID = old.Time.Policy.Binding.Main.Previous.Identity.IssuerID
	renewal.Time.Policy.Binding.Main.Proposed.Identity.IssuerID = old.Time.Policy.Binding.Main.Proposed.Identity.IssuerID
	renewal.Time.Policy.Binding.Main.Previous.Identity.Serial = old.Time.Policy.Binding.Main.Previous.Identity.Serial
	renewal.Time.Policy.Binding.Main.Proposed.Identity.Serial = old.Time.Policy.Binding.Main.Proposed.Identity.Serial
	issuerBindingMismatch(t, renewal)
}
func TestIssuerBindingSameSerialDifferentIssuer(t *testing.T) {
	old := issuerBindingFixture(t, "serial-one-high")
	other := issuerBindingFixture(t, "other-ca-same-serial")
	assertIssuerBindingLower(t, old)
	assertIssuerBindingLower(t, other)
	other.Time.Policy.Binding.Main.Previous.Identity.Serial = old.Time.Policy.Binding.Main.Previous.Identity.Serial
	other.Time.Policy.Binding.Main.Proposed.Identity.Serial = old.Time.Policy.Binding.Main.Proposed.Identity.Serial
	issuerBindingMismatch(t, other)
}
func TestIssuerBindingProjectionAlternatives(t *testing.T) {
	base := issuerBindingFixture(t, "base")
	want := issuerBindingExpected(t, "base")
	ca, e := x509.ParseCertificate(base.Certificate.CA.DER)
	if e != nil {
		t.Fatal(e)
	}
	leaf, e := x509.ParseCertificate(base.Certificate.Leaf.DER)
	if e != nil {
		t.Fatal(e)
	}
	issuerAlternatives := []contracts.Digest{
		contracts.Digest(sha256.Sum256(ca.Raw)),
		contracts.Digest(sha256.Sum256(ca.RawSubjectPublicKeyInfo)),
		contracts.Digest(sha256.Sum256(ca.RawSubject)),
		contracts.Digest(sha256.Sum256(leaf.SubjectKeyId)),
		contracts.Digest(sha256.Sum256(leaf.AuthorityKeyId)),
	}
	for _, wrong := range issuerAlternatives {
		if wrong == want.Issuer {
			t.Fatal("alternative control")
		}
		in := issuerBindingFixture(t, "base")
		in.Time.Policy.Binding.Main.Previous.Identity.IssuerID = wrong
		in.Time.Policy.Binding.Main.Proposed.Identity.IssuerID = wrong
		issuerBindingMismatch(t, in)
	}
	ip := append([]byte("talenro.main-selected-serial-ref.v1\x00"), want.Issuer[:]...)
	caSerial := append(ip, byte(len(ca.SerialNumber.Bytes())))
	caSerial = append(caSerial, ca.SerialNumber.Bytes()...)
	serialAlternatives := []contracts.Digest{contracts.Digest(sha256.Sum256(leaf.SerialNumber.Bytes())), contracts.Digest(sha256.Sum256(caSerial))}
	for _, wrong := range serialAlternatives {
		if wrong == want.Serial {
			t.Fatal("serial alternative control")
		}
		in := issuerBindingFixture(t, "base")
		in.Time.Policy.Binding.Main.Previous.Identity.Serial = wrong
		in.Time.Policy.Binding.Main.Proposed.Identity.Serial = wrong
		issuerBindingMismatch(t, in)
	}
}
