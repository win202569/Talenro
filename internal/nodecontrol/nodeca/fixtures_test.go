package nodeca

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"math/big"
	"testing"
)

const (
	selfCAHex       = "308201aa30820151a00302010202020080300a06082a8648ce3d0403023020311e301c060355040313156e6f6465636120676f6c64656e207375626a656374301e170d3236303130313030303030305a170d3336303130313030303030305a3020311e301c060355040313156e6f6465636120676f6c64656e207375626a6563743059301306072a8648ce3d020106082a8648ce3d030107034200046b17d1f2e12c4247f8bce6e563a440f277037d812deb33a0f4a13945d898c2964fe342e2fe1a7f9b8ee7eb4a7c0f9e162bce33576b315ececbb6406837bf51f5a37b3079300e0603551d0f0101ff040403020204300f0603551d130101ff040530030101ff30290603551d0e042204205cd252fb0ce8932436faf8ccd1040981b89ee4ad6b9fe9e2a2b7e71aacb27cd3302b0603551d230424302280205cd252fb0ce8932436faf8ccd1040981b89ee4ad6b9fe9e2a2b7e71aacb27cd3300a06082a8648ce3d040302034700304402205ae2644aa860aa2f1f51ce08abe696ce3177b9b9d3b340b95c4e704bf6be213d022059a382e0975fa4d86e7a23d7616e46e1c170a0a2805e9a8836533d6931d750ec"
	parentCAHex     = "308201aa30820150a00302010202020080300a06082a8648ce3d040302301f311d301b060355040313146e6f6465636120676f6c64656e20706172656e74301e170d3236303130313030303030305a170d3336303130313030303030305a3020311e301c060355040313156e6f6465636120676f6c64656e207375626a6563743059301306072a8648ce3d020106082a8648ce3d030107034200046b17d1f2e12c4247f8bce6e563a440f277037d812deb33a0f4a13945d898c2964fe342e2fe1a7f9b8ee7eb4a7c0f9e162bce33576b315ececbb6406837bf51f5a37b3079300e0603551d0f0101ff040403020204300f0603551d130101ff040530030101ff30290603551d0e042204205cd252fb0ce8932436faf8ccd1040981b89ee4ad6b9fe9e2a2b7e71aacb27cd3302b0603551d23042430228020dc0ce633dbcc913dafafa4b89ac44d8ce683fdfc3f60c8bdf21213b9f2b534ba300a06082a8648ce3d040302034800304502200affc6b941f40670d34501749d3832f06f2c13b6a30c878ef341181d50290a2e022100dd648176002caf39b715d7a3948a700be2381edaba6a1cdceaf714ba93dc8ba4"
	subjectSPKIHex  = "3059301306072a8648ce3d020106082a8648ce3d030107034200046b17d1f2e12c4247f8bce6e563a440f277037d812deb33a0f4a13945d898c2964fe342e2fe1a7f9b8ee7eb4a7c0f9e162bce33576b315ececbb6406837bf51f5"
	rfcSPKIHex      = "3059301306072a8648ce3d020106082a8648ce3d030107034200047f7f35a79794c950060b8029fc8f363a28f11159692d9d34e6ac948190434735f833b1a66652dc514337aff7f5c9c75d670c019d95a5d639b72744c64a9128bb"
	selfDERHash     = "3970d1ec0c4162e6087d7e50bbad778101dfcb8c9b0aa0d4746798807b7d140e"
	parentDERHash   = "2855dc11d66743ca939fda0326b761d019e080796ce2a03fa789bdd47c857460"
	subjectSPKIHash = "5cd252fb0ce8932436faf8ccd1040981b89ee4ad6b9fe9e2a2b7e71aacb27cd3"
	subjectNameHash = "d2db35c94e6d81fd8a560de3436b105edf829722d7bf67b20b4ffd47ef01e2ff"
	parentNameHash  = "afabbc957eb6dde91ccb5d672b0604358e2a6425a8e3befe5fc2d7ab19dc1858"
	rfcSPKIHash     = "6d20896ab8bd833b6b66554bd59b20225d8a75a296088148399d7bf763d57405"
	parentKeyHash   = "dc0ce633dbcc913dafafa4b89ac44d8ce683fdfc3f60c8bdf21213b9f2b534ba"
)

func fixtureHex(s string) []byte {
	v, e := hex.DecodeString(s)
	if e != nil {
		panic(e)
	}
	return v
}
func fixtureInput(parentSigned bool) Input {
	s := selfCAHex
	if parentSigned {
		s = parentCAHex
	}
	return Input{DER: fixtureHex(s), DeclaredSPKI: fixtureHex(subjectSPKIHex)}
}
func fixtureKey(n int64, curve elliptic.Curve) *ecdsa.PrivateKey {
	d := big.NewInt(n)
	x, y := curve.ScalarBaseMult(d.Bytes())
	return &ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{Curve: curve, X: x, Y: y}, D: d}
}
func fixtureTLV(tag byte, b []byte) []byte {
	head := []byte{tag}
	n := len(b)
	if n < 128 {
		head = append(head, byte(n))
	} else {
		var size []byte
		for x := n; x > 0; x >>= 8 {
			size = append([]byte{byte(x)}, size...)
		}
		head = append(head, 0x80|byte(len(size)))
		head = append(head, size...)
	}
	return append(head, b...)
}
func fixtureJoin(parts [][]byte) []byte { return fixtureTLV(0x30, bytes.Join(parts, nil)) }
func fixtureParts(t testing.TB, b []byte) [][]byte {
	t.Helper()
	var v asn1.RawValue
	tail, e := asn1.Unmarshal(b, &v)
	if e != nil || len(tail) != 0 {
		t.Fatalf("fixture container %v", e)
	}
	var out [][]byte
	for rest := v.Bytes; len(rest) > 0; {
		var child asn1.RawValue
		next, e := asn1.Unmarshal(rest, &child)
		if e != nil {
			t.Fatal("fixture child", e)
		}
		out = append(out, bytes.Clone(child.FullBytes))
		rest = next
	}
	return out
}
func fixtureRewrite(t testing.TB, in Input, path string, b []byte) Input {
	t.Helper()
	in.DeclaredSPKI = bytes.Clone(in.DeclaredSPKI)
	cert := fixtureParts(t, in.DER)
	if path == "certificate" {
		in.DER = bytes.Clone(b)
		return in
	}
	if path == "signature" {
		cert[2] = b
		in.DER = fixtureJoin(cert)
		return in
	}
	if path == "outerAlgorithm" {
		cert[1] = b
		in.DER = fixtureJoin(cert)
		return in
	}
	if path == "tbs" {
		cert[0] = b
		in.DER = fixtureJoin(cert)
		return in
	}
	fields := fixtureParts(t, cert[0])
	index := -1
	switch path {
	case "version":
		index = 0
	case "serial":
		index = 1
	case "innerAlgorithm":
		index = 2
	case "issuer":
		index = 3
	case "subject":
		index = 5
	case "spki":
		index = 6
	case "extensions":
		index = 7
	case "notBefore", "notAfter":
		times := fixtureParts(t, fields[4])
		j := 0
		if path == "notAfter" {
			j = 1
		}
		times[j] = b
		fields[4] = fixtureJoin(times)
	default:
		t.Fatal("fixture path", path)
	}
	if index >= 0 {
		if b == nil {
			fields = append(fields[:index], fields[index+1:]...)
		} else {
			fields[index] = b
		}
	}
	cert[0] = fixtureJoin(fields)
	in.DER = fixtureJoin(cert)
	return in
}
func fixtureExts(t testing.TB, in Input) []pkix.Extension {
	t.Helper()
	fields := fixtureParts(t, fixtureParts(t, in.DER)[0])
	var wrapper asn1.RawValue
	_, err := asn1.Unmarshal(fields[7], &wrapper)
	if err != nil {
		t.Fatal(err)
	}
	var e []pkix.Extension
	tail, err := asn1.Unmarshal(wrapper.Bytes, &e)
	if err != nil || len(tail) != 0 {
		t.Fatal("fixture extensions", err)
	}
	return e
}
func fixtureSetExts(t testing.TB, in Input, e []pkix.Extension) Input {
	t.Helper()
	b, err := asn1.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	return fixtureRewrite(t, in, "extensions", fixtureTLV(0xa3, b))
}
func fixtureExt(t testing.TB, in Input, last int, edit func(*pkix.Extension)) Input {
	t.Helper()
	e := fixtureExts(t, in)
	for i := range e {
		if e[i].Id.Equal(asn1.ObjectIdentifier{2, 5, 29, last}) {
			edit(&e[i])
			return fixtureSetExts(t, in, e)
		}
	}
	t.Fatal("fixture missing extension")
	return Input{}
}
func fixtureCert(t testing.TB, in Input) *x509.Certificate {
	t.Helper()
	c, e := x509.ParseCertificate(in.DER)
	if e != nil {
		t.Fatal("fixture parser", e)
	}
	return c
}
