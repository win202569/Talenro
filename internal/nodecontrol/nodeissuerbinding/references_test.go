package nodeissuerbinding

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"talenro.local/platform/internal/nodecontrol/contracts"
	"talenro.local/platform/internal/nodecontrol/nodeleafbinding"
	"testing"
)

func TestIssuerBindingReferenceVectors(t *testing.T) {
	if len(issuerReferenceDomain) != 36 || len(serialReferenceDomain) != 36 {
		t.Fatal("domain size")
	}
	for _, name := range []string{"base", "leaf-first", "ca-first", "serial-one-high", "serial-twenty", "ca-renewal", "other-ca-same-serial"} {
		s := assertIssuerBindingLower(t, issuerBindingFixture(t, name))
		want := issuerBindingExpected(t, name)
		var scratch referenceScratch
		i, r, e := deriveReferences(context.Background(), s, &scratch)
		if e != nil || i != want.Issuer || r != want.Serial || scratch != (referenceScratch{}) {
			t.Fatal("frozen vector", name, e)
		}
		ip, e := hex.DecodeString(want.IssuerPreimage)
		if e != nil {
			t.Fatal(e)
		}
		rp, e := hex.DecodeString(want.SerialPreimage)
		if e != nil {
			t.Fatal(e)
		}
		if len(ip) != 68 || len(rp) != 69+int(s.Evidence.Certificate.Leaf.Leaf.SerialLength) || len(rp) < 70 || len(rp) > 89 {
			t.Fatal("preimage sizes")
		}
		for _, wrong := range [][]byte{
			ip[:len(ip)-1],
			append([]byte("talenro.main-selected-serial-ref.v1\x00"), ip[36:]...),
			[]byte(hex.EncodeToString(ip)),
			append([]byte("talenro.main-selected-issuer-ref.v2\x00"), ip[36:]...),
		} {
			if contracts.Digest(sha256.Sum256(wrong)) == want.Issuer {
				t.Fatal("nondiscriminating issuer alternative")
			}
		}
		for _, wrong := range [][]byte{rp[:len(rp)-1], []byte(hex.EncodeToString(rp)), append(rp[:68:68], rp[69:]...), append([]byte("talenro.main-selected-issuer-ref.v1\x00"), rp[36:]...)} {
			if contracts.Digest(sha256.Sum256(wrong)) == want.Serial {
				t.Fatal("nondiscriminating serial alternative")
			}
		}
	}
}
func TestIssuerBindingCanonicalSerial(t *testing.T) {
	valid := assertIssuerBindingLower(t, issuerBindingFixture(t, "base"))
	for _, mutate := range []func(*nodeleafbinding.Summary){
		func(s *nodeleafbinding.Summary) { s.Evidence.Certificate.CA.CADERDigest = contracts.Digest{} },
		func(s *nodeleafbinding.Summary) { s.Evidence.Certificate.Leaf.Leaf.SerialLength = 0 },
		func(s *nodeleafbinding.Summary) { s.Evidence.Certificate.Leaf.Leaf.SerialLength = 21 },
		func(s *nodeleafbinding.Summary) { s.Evidence.Certificate.Leaf.Leaf.Serial[0] = 0 },
		func(s *nodeleafbinding.Summary) {
			s.Evidence.Certificate.Leaf.Leaf.SerialLength = 1
			s.Evidence.Certificate.Leaf.Leaf.Serial[1] = 1
		},
	} {
		s := valid
		mutate(&s)
		var scratch referenceScratch
		i, r, e := deriveReferences(context.Background(), s, &scratch)
		if e != ErrInvalidEvidence || i != (contracts.Digest{}) || r != (contracts.Digest{}) || scratch != (referenceScratch{}) {
			t.Fatal("invalid canonical serial", e)
		}
	}
	// Frozen independently with .NET SHA256, not by the product helper.
	var high nodeleafbinding.Summary
	high.Evidence.Certificate.CA.CADERDigest[0] = 1
	high.Evidence.Certificate.Leaf.Leaf.SerialLength = 20
	high.Evidence.Certificate.Leaf.Leaf.Serial[0] = 0x80
	high.Evidence.Certificate.Leaf.Leaf.Serial[19] = 1
	var scratch referenceScratch
	i, r, e := deriveReferences(context.Background(), high, &scratch)
	if e != nil || hex.EncodeToString(i[:]) != "c34f2f409acc65b77c1ff89367aeb667324849c6312e8e1d6767ae59814ae588" || hex.EncodeToString(r[:]) != "3c593c46fb864ab118666092b9758ab7eacc60ebe2ba1f2a2d23fc6226e6f537" {
		t.Fatal("20-byte high magnitude private encoding", e)
	}
}
func fillReferenceScratch(s *referenceScratch) {
	for j := range s.issuer {
		s.issuer[j] = 0xff
	}
	for j := range s.serial {
		s.serial[j] = 0xff
	}
}
func TestIssuerBindingReferenceCleanup(t *testing.T) {
	s := assertIssuerBindingLower(t, issuerBindingFixture(t, "base"))
	before := ownedTestClone(t, s)
	if _, _, e := deriveReferences(context.Background(), s, nil); e != ErrInvalidEvidence {
		t.Fatal("nil scratch", e)
	}
	var scratch referenceScratch
	fillReferenceScratch(&scratch)
	if _, _, e := deriveReferences(nil, s, &scratch); e != ErrInvalidInput || scratch != (referenceScratch{}) {
		t.Fatal("nil context cleanup", e)
	}
	probe := &ownershipFailureContext{Context: context.Background()}
	if _, _, e := deriveReferences(probe, s, &scratch); e != nil {
		t.Fatal(e)
	}
	for _, failure := range []error{context.Canceled, context.DeadlineExceeded} {
		for stop := 1; stop <= probe.calls; stop++ {
			fillReferenceScratch(&scratch)
			ctx := &ownershipFailureContext{Context: context.Background(), stop: stop, failure: failure}
			i, r, e := deriveReferences(ctx, s, &scratch)
			if e != failure || i != (contracts.Digest{}) || r != (contracts.Digest{}) || scratch != (referenceScratch{}) {
				t.Fatal("failure cleanup", stop, e)
			}
		}
	}
	for stop := 1; stop <= probe.calls; stop++ {
		func() {
			fillReferenceScratch(&scratch)
			defer func() {
				if recover() != "ownership test panic" || scratch != (referenceScratch{}) {
					t.Fatal("panic cleanup", stop)
				}
			}()
			deriveReferences(&ownershipFailureContext{Context: context.Background(), stop: stop, panicAt: true}, s, &scratch)
		}()
	}
	if !reflect.DeepEqual(s, before) {
		t.Fatal("evidence mutated")
	}
}
