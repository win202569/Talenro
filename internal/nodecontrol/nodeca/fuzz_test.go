package nodeca

import (
	"bytes"
	"context"
	"testing"
)

func FuzzCADER(f *testing.F) {
	f.Add(fixtureHex(selfCAHex))
	f.Add(fixtureHex(parentCAHex))
	f.Add([]byte{})
	f.Add([]byte{0x30, 0x88, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	f.Add(bytes.Repeat([]byte{1}, 65537))
	f.Fuzz(func(t *testing.T, b []byte) {
		m, e := Parse(context.Background(), Input{DER: b, DeclaredSPKI: fixtureHex(subjectSPKIHex)})
		if len(b) > 65536 && e != ErrLimitExceeded {
			t.Fatal("budget before clone", e)
		}
		if e != nil {
			if m != nil {
				m.Clear()
				t.Fatal("material on error")
			}
			return
		}
		if m == nil {
			t.Fatal("nil success")
		}
		got := m.Summary()
		if got.Proof != NodeCAMaterialShapeAndDeclaredKeyBound || got.Pending != expectedPending() {
			t.Fatal("strength")
		}
		m.Clear()
		if m.Summary() != (Summary{}) {
			t.Fatal("clear")
		}
	})
}
func FuzzCASPKI(f *testing.F) {
	f.Add(fixtureHex(rfcSPKIHex))
	f.Add(fixtureHex(subjectSPKIHex))
	f.Add([]byte{})
	f.Add([]byte{0x30, 0x80, 0, 0})
	f.Add(bytes.Repeat([]byte{1}, 8193))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 8192 {
			if e := preflight(context.Background(), Input{DER: []byte{1}, DeclaredSPKI: b}); e != ErrLimitExceeded {
				t.Fatal("budget", e)
			}
			return
		}
		e := checkDeclaredSPKI(b)
		if e != nil && e != ErrInvalidInput && e != ErrUnsupportedCAProfile {
			t.Fatal("dynamic classification", e)
		}
		if e == nil && len(b) != 91 {
			t.Fatal("noncanonical success")
		}
	})
}
