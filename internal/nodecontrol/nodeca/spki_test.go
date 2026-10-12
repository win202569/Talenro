package nodeca

import (
	"bytes"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/big"
	"testing"
)

func TestCADeclaredSPKI(t *testing.T) {
	p := fixtureHex(rfcSPKIHex)
	if len(p) != 91 {
		t.Fatal("RFC fixture length")
	}
	parts := fixtureParts(t, p)
	compressed := append([]byte{0, 2}, parts[1][4:36]...)
	cases := []struct {
		name string
		b    []byte
		want error
	}{
		{"RFC", p, nil}, {"new subject", fixtureInput(false).DeclaredSPKI, nil}, {"empty", nil, ErrInvalidInput},
		{"tail", append(bytes.Clone(p), 0), ErrInvalidInput}, {"PEM", pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: p}), ErrInvalidInput},
		{"compressed", fixtureJoin([][]byte{parts[0], fixtureTLV(3, compressed)}), ErrInvalidInput},
		{"explicit", fixtureJoin([][]byte{fixtureJoin([][]byte{fixtureParts(t, parts[0])[0], []byte{0x30, 0}}), parts[1]}), ErrInvalidInput},
		{"invalid point", append(bytes.Clone(p[:27]), make([]byte, 64)...), ErrInvalidInput},
		{"nonminimal", append([]byte{0x30, 0x81, 0x59}, p[2:]...), ErrInvalidInput},
	}
	for _, curve := range []elliptic.Curve{elliptic.P384(), elliptic.P521(), elliptic.P224()} {
		b, e := x509.MarshalPKIXPublicKey(&fixtureKey(2, curve).PublicKey)
		if e == nil {
			cases = append(cases, struct {
				name string
				b    []byte
				want error
			}{curve.Params().Name, b, ErrUnsupportedCAProfile})
		}
	}
	ed, _ := x509.MarshalPKIXPublicKey(ed25519.PublicKey(make([]byte, 32)))
	cases = append(cases, struct {
		name string
		b    []byte
		want error
	}{"Ed25519", ed, ErrUnsupportedCAProfile})
	rsaDER, e := x509.MarshalPKIXPublicKey(&rsa.PublicKey{N: new(big.Int).SetBytes(bytes.Repeat([]byte{0x7f}, 256)), E: 65537})
	if e != nil {
		t.Fatal(e)
	}
	cases = append(cases, struct {
		name string
		b    []byte
		want error
	}{"RSA", rsaDER, ErrUnsupportedCAProfile})
	xKey, e := ecdh.X25519().NewPublicKey(bytes.Repeat([]byte{1}, 32))
	if e != nil {
		t.Fatal(e)
	}
	xDER, e := x509.MarshalPKIXPublicKey(xKey)
	if e != nil {
		t.Fatal(e)
	}
	cases = append(cases, struct {
		name string
		b    []byte
		want error
	}{"X25519", xDER, ErrUnsupportedCAProfile})
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := checkDeclaredSPKI(c.b); !errors.Is(got, c.want) {
				t.Fatalf("got %v want %v", got, c.want)
			}
		})
	}
}
func TestCAWireSPKI(t *testing.T) {
	in := fixtureInput(false)
	c := fixtureCert(t, in)
	raw, e := parseRawCA(in.DER)
	if e != nil {
		t.Fatal(e)
	}
	if e = checkCASPKI(c, raw); e != nil {
		t.Fatal(e)
	}
	bad := raw
	bad.spki.full = append(bytes.Clone(raw.spki.full), 0)
	if e = checkCASPKI(c, bad); !errors.Is(e, ErrInvalidMaterial) {
		t.Fatal("raw mismatch", e)
	}
	copied := *c
	copied.PublicKey = &fixtureKey(2, elliptic.P256()).PublicKey
	if e = checkCASPKI(&copied, raw); !errors.Is(e, ErrInvalidMaterial) {
		t.Fatal("parsed key mismatch", e)
	}
	if e = checkCASPKI(nil, raw); !errors.Is(e, ErrInvalidMaterial) {
		t.Fatal("nil parser", e)
	}
}
