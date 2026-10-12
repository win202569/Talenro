package nodeleafkeys

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"math/big"
	"testing"
)

func keysSPKIVariants(t testing.TB) map[string][]byte {
	t.Helper()
	ec := keysIssuerSPKI(t, "p256")
	ed := keysIssuerSPKI(t, "ed25519")
	rsaDER := keysIssuerSPKI(t, "rsa")
	ep := keysParts(t, ec)
	alg := keysParts(t, ep[0])
	explicit := append([][]byte(nil), alg...)
	explicit[1] = []byte{0x30, 0}
	extraAlg := keysJoin(append(append([][]byte(nil), alg...), []byte{5, 0}))
	compressed := elliptic.MarshalCompressed(elliptic.P256(), keysTestEC(elliptic.P256(), 1).X, keysTestEC(elliptic.P256(), 1).Y)
	var bits asn1.BitString
	if _, err := asn1.Unmarshal(ep[1], &bits); err != nil {
		t.Fatal(err)
	}
	badPoint := bytes.Clone(bits.Bytes)
	clear(badPoint[1:])
	rp := keysParts(t, rsaDER)
	var rbits asn1.BitString
	if _, err := asn1.Unmarshal(rp[1], &rbits); err != nil {
		t.Fatal(err)
	}
	ri := keysParts(t, rbits.Bytes)
	rsaInner := keysJoin(append(ri, []byte{5, 0}))
	rsaTrailing := keysJoin([][]byte{rp[0], keysTLV(3, append([]byte{0}, rsaInner...))})
	edp := keysParts(t, ed)
	eda := keysParts(t, edp[0])
	unused := bytes.Clone(ep[1]) // BIT STRING unused-bits byte follows tag + one-byte length
	unused[2] = 1
	unused[len(unused)-1] &= 0xfe
	nonminimal := append([]byte{0x30, 0x81, ec[1]}, ec[2:]...)
	pk, err := x509.ParsePKIXPublicKey(rsaDER)
	if err != nil {
		t.Fatal(err)
	}
	return map[string][]byte{
		"trailing": append(bytes.Clone(ec), 0), "concatenated": append(bytes.Clone(ec), ec...),
		"pem":                    pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: ec}),
		"pkcs1":                  x509.MarshalPKCS1PublicKey(pk.(*rsa.PublicKey)),
		"nonminimal-length":      nonminimal,
		"ec-parameters-trailing": keysJoin([][]byte{extraAlg, ep[1]}),
		"explicit-ec":            keysJoin([][]byte{keysJoin(explicit), ep[1]}),
		"compressed-ec":          keysJoin([][]byte{ep[0], keysTLV(3, append([]byte{0}, compressed...))}),
		"bad-point":              keysJoin([][]byte{ep[0], keysTLV(3, append([]byte{0}, badPoint...))}),
		"rsa-inner-trailing":     rsaTrailing,
		"ed-parameters":          keysJoin([][]byte{keysJoin(append(eda, []byte{5, 0})), edp[1]}),
		"ed-short":               keysJoin([][]byte{edp[0], keysTLV(3, append([]byte{0}, make([]byte, 31)...))}),
		"unused-bits":            keysJoin([][]byte{ep[0], unused}),
	}
}
func keysUnsupportedSPKIs(t testing.TB) map[string][]byte {
	t.Helper()
	p224, err := x509.MarshalPKIXPublicKey(&keysTestEC(elliptic.P224(), 1).PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	xkey, err := ecdh.X25519().NewPrivateKey(make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	xraw, err := x509.MarshalPKIXPublicKey(xkey.PublicKey())
	if err != nil {
		t.Fatal(err)
	}
	oid, err := asn1.Marshal(asn1.ObjectIdentifier{1, 2, 840, 10040, 4, 1})
	if err != nil {
		t.Fatal(err)
	}
	params := keysJoin([][]byte{{2, 1, 23}, {2, 1, 11}, {2, 1, 2}})
	dsaRaw := keysJoin([][]byte{keysJoin([][]byte{oid, params}), keysTLV(3, []byte{0, 2, 1, 4})})
	unknownOID, err := asn1.Marshal(asn1.ObjectIdentifier{1, 2, 3, 4})
	if err != nil {
		t.Fatal(err)
	}
	unknown := keysJoin([][]byte{keysJoin([][]byte{unknownOID}), keysTLV(3, []byte{0, 1})})
	return map[string][]byte{"p224": p224, "x25519": xraw, "dsa": dsaRaw, "unknown": unknown}
}
func TestKeysIssuerKinds(t *testing.T) {
	for _, kind := range []string{"p256", "p384", "p521", "rsa", "ed25519"} {
		raw := keysIssuerSPKI(t, kind)
		before := bytes.Clone(raw)
		if err := checkDeclaredIssuerSPKI(raw); err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		if !bytes.Equal(raw, before) {
			t.Fatal("input changed")
		}
	}
	for name, raw := range keysUnsupportedSPKIs(t) {
		if name != "unknown" {
			if _, err := x509.ParsePKIXPublicKey(raw); err != nil {
				t.Fatalf("unsupported fixture %s: %v", name, err)
			}
		}
		if checkDeclaredIssuerSPKI(raw) != ErrInvalidInput {
			t.Fatal("unsupported accepted", name)
		}
	}
}
func TestKeysIssuerCanonical(t *testing.T) {
	variants := keysSPKIVariants(t)
	raw := variants["rsa-inner-trailing"]
	pk, err := x509.ParsePKIXPublicKey(raw)
	if err != nil {
		t.Fatal("Go-accepted control", err)
	}
	canonical, err := x509.MarshalPKIXPublicKey(pk)
	if err != nil || bytes.Equal(raw, canonical) {
		t.Fatal("canonical difference control", err)
	}
	t.Logf("fixed Go parser accepts RSA inner trailing; raw=%d canonical=%d, relation checker must reject", len(raw), len(canonical))
	for name, b := range variants {
		before := bytes.Clone(b)
		if err := checkDeclaredIssuerSPKI(b); err != ErrInvalidInput {
			t.Fatalf("%s: %v", name, err)
		}
		if !bytes.Equal(b, before) {
			t.Fatal("canonical repair", name)
		}
	}
}
func TestKeysIssuerBudgetEdge(t *testing.T) {
	magnitude := make([]byte, 8154)
	magnitude[0] = 0x80
	magnitude[len(magnitude)-1] = 1
	key := &rsa.PublicKey{N: new(big.Int).SetBytes(magnitude), E: 65537}
	raw, err := x509.MarshalPKIXPublicKey(key)
	if err != nil || len(raw) != 8192 {
		t.Fatalf("edge bytes=%d err=%v", len(raw), err)
	}
	parsed, err := x509.ParsePKIXPublicKey(raw)
	if err != nil {
		t.Fatal(err)
	}
	round, err := x509.MarshalPKIXPublicKey(parsed)
	if err != nil || !bytes.Equal(raw, round) {
		t.Fatal("edge control", err)
	}
	in := keysBudgetInput()
	in.DeclaredIssuerSPKI = raw
	if preflight(context.Background(), in) != nil || checkDeclaredIssuerSPKI(raw) != nil {
		t.Fatal("8192 edge rejected")
	}
}
