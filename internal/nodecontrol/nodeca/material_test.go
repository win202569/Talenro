package nodeca

import (
	"testing"
	"time"
)

func epochTime() time.Time { return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC) }

func TestCAMaterialLifecycle(t *testing.T) {
	var nilMaterial *Material
	nilMaterial.Clear()
	if nilMaterial.Summary() != (Summary{}) {
		t.Fatal("nil")
	}
	zero := &Material{}
	zero.Clear()
	if zero.Summary() != (Summary{}) {
		t.Fatal("zero")
	}
	o := ownInput(Input{DER: []byte{1}, DeclaredSPKI: []byte{2}})
	m := &Material{owned: o}
	copied := *m
	want := Summary{Proof: NodeCAMaterialShapeAndDeclaredKeyBound, Pending: [6]Obligation{1, 2, 3, 4, 5, 6}, CADERDigest: [32]byte{1}, SPKIDERDigest: [32]byte{2}, SubjectDERDigest: [32]byte{3}, IssuerDERDigest: [32]byte{4}, SubjectKeyIdentifier: KeyIdentifier{5}, AuthorityKeyIdentifier: KeyIdentifier{6}, Serial: [20]byte{7}, SerialLength: 1, NotBefore: epochTime(), NotAfter: epochTime().Add(time.Hour), HasPathLen: true, PathLen: 255, CRLSignAllowed: true}
	o.summary = want
	if m.Summary() != (Summary{}) {
		t.Fatal("invalid owner visible")
	}
	o.valid = true
	detached := m.Summary()
	if detached != want || copied.Summary() != want {
		t.Fatal("value projection")
	}
	m.Clear()
	copied.Clear()
	m.Clear()
	if m.Summary() != (Summary{}) || copied.Summary() != (Summary{}) || o.summary != (Summary{}) || detached != want {
		t.Fatal("shared or detached lifecycle")
	}
}
