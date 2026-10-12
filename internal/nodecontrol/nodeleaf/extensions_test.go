package nodeleaf

import (
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"testing"
)

type extensionCase struct {
	name  string
	index int
	value []byte
}

func leafExtensionCases() []extensionCase {
	uri := []byte("spiffe://node.test/node/11111111-2222-4333-8444-555555555555")
	cases := []extensionCase{
		{"BC explicit false", 0, []byte{0x30, 3, 1, 1, 0}},
		{"BC CA", 0, []byte{0x30, 3, 1, 1, 255}},
		{"BC pathLen", 0, []byte{0x30, 3, 2, 1, 0}},
		{"KU extra", 1, []byte{3, 2, 6, 0xc0}},
		{"KU empty", 1, []byte{3, 1, 0}},
		{"KU wrong padding", 1, []byte{3, 2, 0, 0x80}},
		{"EKU server", 2, []byte{0x30, 10, 6, 8, 0x2b, 6, 1, 5, 5, 7, 3, 1}},
		{"EKU any", 2, []byte{0x30, 6, 6, 4, 0x55, 0x1d, 0x25, 0}},
		{"EKU duplicate", 2, fixtureTLV(0x30, bytes.Repeat([]byte{6, 8, 0x2b, 6, 1, 5, 5, 7, 3, 2}, 2))},
		{"SAN empty", 3, []byte{0x30, 0}},
		{"SAN zero URI", 3, []byte{0x30, 2, 0x86, 0}},
		{"SAN oversize", 3, fixtureTLV(0x30, fixtureTLV(0x86, bytes.Repeat([]byte{'a'}, 305)))},
		{"SAN Unicode", 3, fixtureTLV(0x30, fixtureTLV(0x86, []byte{0xc3, 0xa9}))},
		{"SAN constructed", 3, fixtureTLV(0x30, fixtureTLV(0xa6, uri))},
		{"SAN two URI", 3, fixtureTLV(0x30, bytes.Repeat(fixtureTLV(0x86, uri), 2))},
		{"SKI zero", 4, []byte{4, 0}},
		{"SKI long", 4, fixtureTLV(4, bytes.Repeat([]byte{1}, 65))},
		{"AKI zero", 5, []byte{0x30, 2, 0x80, 0}},
		{"AKI long", 5, fixtureTLV(0x30, fixtureTLV(0x80, bytes.Repeat([]byte{1}, 65)))},
		{"AKI constructed", 5, []byte{0x30, 3, 0xa0, 1, 1}},
		{"AKI duplicate", 5, []byte{0x30, 6, 0x80, 1, 1, 0x80, 1, 2}},
		{"AKI issuer", 5, []byte{0x30, 5, 0x80, 1, 1, 0xa1, 0}},
		{"AKI serial", 5, []byte{0x30, 6, 0x80, 1, 1, 0x82, 1, 1}},
	}
	for i, e := range fixtureExtensions() {
		cases = append(cases, extensionCase{"inner trailing", i, append(bytes.Clone(e.Value), 5, 0)}, extensionCase{"inner wrong tag", i, append([]byte{5}, e.Value[1:]...)}, extensionCase{"inner nonminimal", i, append([]byte{e.Value[0], 0x81, e.Value[1]}, e.Value[2:]...)})
	}
	return cases
}

func fixtureWithExtension(t testing.TB, index int, value []byte) Input {
	return leafWithTemplate(t, func(c *x509.Certificate) { c.ExtraExtensions[index].Value = bytes.Clone(value) })
}

func fixtureExtensionBlock(ext []pkix.Extension, explicitFalse bool) []byte {
	var parts [][]byte
	for _, e := range ext {
		oid, err := asn1.Marshal(e.Id)
		if err != nil {
			panic(err)
		}
		body := oid
		if e.Critical {
			body = append(body, 1, 1, 255)
		} else if explicitFalse {
			body = append(body, 1, 1, 0)
		}
		body = append(body, fixtureTLV(4, e.Value)...)
		parts = append(parts, fixtureTLV(0x30, body))
	}
	return fixtureTLV(0xa3, fixtureJoin(parts))
}

func TestLeafExtensionInventory(t *testing.T) {
	in := leafFixture(t)
	cert, e := x509.ParseCertificate(in.DER)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := parseRawLeaf(in.DER)
	if e != nil {
		t.Fatal(e)
	}
	// All permutations, not just one reversal, must be accepted.
	var permute func([]rawExtension, int)
	permute = func(items []rawExtension, n int) {
		if n == len(items) {
			if _, e := checkExtensions(items, cert); e != nil {
				t.Fatal("permutation", e)
			}
			return
		}
		for i := n; i < len(items); i++ {
			items[n], items[i] = items[i], items[n]
			permute(items, n+1)
			items[n], items[i] = items[i], items[n]
		}
	}
	permute(append([]rawExtension(nil), raw.extensions...), 0)
	for i := range raw.extensions {
		missing := append([]rawExtension(nil), raw.extensions[:i]...)
		missing = append(missing, raw.extensions[i+1:]...)
		duplicate := append(append([]rawExtension(nil), raw.extensions...), raw.extensions[i])
		wrong := append([]rawExtension(nil), raw.extensions...)
		wrong[i].critical = !wrong[i].critical
		for _, items := range [][]rawExtension{missing, duplicate, wrong} {
			if _, e := checkExtensions(items, cert); !errors.Is(e, ErrUnsupportedLeafProfile) {
				t.Fatal("inventory accepted", i, e)
			}
		}
	}
	for _, critical := range []bool{false, true} {
		unknown := rawExtension{id: asn1.ObjectIdentifier{1, 2, 3, 4}, critical: critical, value: []byte{5, 0}}
		for _, items := range [][]rawExtension{append(append([]rawExtension(nil), raw.extensions...), unknown), append([]rawExtension{unknown}, raw.extensions[1:]...)} {
			if _, e := checkExtensions(items, cert); !errors.Is(e, ErrUnsupportedLeafProfile) {
				t.Fatal("unknown extension", e)
			}
		}
	}
	bad := rewriteFixture(t, in, "extensions", fixtureExtensionBlock(fixtureExtensions(), true))
	if _, e := parseRawLeaf(bad.DER); !errors.Is(e, ErrInvalidMaterial) {
		t.Fatal("explicit default critical", e)
	}
}

func TestLeafExtensionValues(t *testing.T) {
	in := leafFixture(t)
	cert, e := x509.ParseCertificate(in.DER)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := parseRawLeaf(in.DER)
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range leafExtensionCases() {
		t.Run(tc.name+string(rune('0'+tc.index)), func(t *testing.T) {
			items := append([]rawExtension(nil), raw.extensions...)
			items[tc.index].value = tc.value
			// Some variants are rejected by Go first; this unit assertion isolates our raw rule.
			cc := *cert
			cc.Extensions = append([]pkix.Extension(nil), cert.Extensions...)
			cc.Extensions[tc.index].Value = tc.value
			if _, e := checkExtensions(items, &cc); e == nil {
				t.Fatal("invalid value accepted")
			}
		})
	}
	for _, n := range []int{1, 64} {
		good := leafWithTemplate(t, func(c *x509.Certificate) {
			c.ExtraExtensions[4].Value = fixtureTLV(4, bytes.Repeat([]byte{7}, n))
			c.ExtraExtensions[5].Value = fixtureTLV(0x30, fixtureTLV(0x80, bytes.Repeat([]byte{8}, n)))
		})
		c, e := x509.ParseCertificate(good.DER)
		if e != nil {
			t.Fatal(e)
		}
		r, e := parseRawLeaf(good.DER)
		if e != nil {
			t.Fatal(e)
		}
		if _, e := checkExtensions(r.extensions, c); e != nil {
			t.Fatal("key ID boundary", n, e)
		}
	}
	cc := *cert
	cc.BasicConstraintsValid = false
	if _, e := checkExtensions(raw.extensions, &cc); e == nil {
		t.Fatal("parsed BC ignored")
	}
	cc = *cert
	cc.IsCA = true
	if _, e := checkExtensions(raw.extensions, &cc); e == nil {
		t.Fatal("parsed CA ignored")
	}
	cc = *cert
	cc.KeyUsage |= x509.KeyUsageKeyEncipherment
	if _, e := checkExtensions(raw.extensions, &cc); e == nil {
		t.Fatal("parsed KU ignored")
	}
	cc = *cert
	cc.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	if _, e := checkExtensions(raw.extensions, &cc); e == nil {
		t.Fatal("parsed EKU ignored")
	}
	cc = *cert
	cc.SubjectKeyId = []byte{99}
	if _, e := checkExtensions(raw.extensions, &cc); e == nil {
		t.Fatal("parsed SKI ignored")
	}
	cc = *cert
	cc.AuthorityKeyId = []byte{99}
	if _, e := checkExtensions(raw.extensions, &cc); e == nil {
		t.Fatal("parsed AKI ignored")
	}
}

func TestLeafIgnoredSANArm(t *testing.T) {
	uri := fixtureParts(t, fixtureExtensions()[3].Value)[0]
	for _, tc := range []struct {
		name    string
		arm     []byte
		ignored bool
	}{
		{"otherName", []byte{0xa0, 7, 6, 2, 0x2a, 3, 0xa0, 1, 0}, true},
		{"registeredID", []byte{0x88, 2, 0x2a, 3}, true},
		{"DNS", fixtureTLV(0x82, []byte("node.test")), false},
		{"IP", []byte{0x87, 4, 127, 0, 0, 1}, false},
		{"email", fixtureTLV(0x81, []byte("a@node.test")), false},
		{"secondURI", uri, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := fixtureWithExtension(t, 3, fixtureTLV(0x30, append(bytes.Clone(uri), tc.arm...)))
			c, e := x509.ParseCertificate(in.DER)
			if e != nil {
				t.Fatal("Go control", e)
			}
			if tc.ignored && len(c.URIs) != 1 {
				t.Fatal("ignored arm control")
			}
			r, e := parseRawLeaf(in.DER)
			if e != nil {
				t.Fatal(e)
			}
			if _, e := checkExtensions(r.extensions, c); !errors.Is(e, ErrUnsupportedLeafProfile) {
				t.Fatal("additional GeneralName accepted", e)
			}
		})
	}
}
