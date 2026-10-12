package nodeissuer

import (
	"bytes"
	"testing"
)

// Catches permissive/nonminimal or overflowing envelope lengths and cursors.
func TestIssuerDERTLV(t *testing.T) {
	for _, n := range []int{0, 1, 127, 128, 255, 256, 65530} {
		raw := fixtureTLV(4, make([]byte, n))
		withTail := append(bytes.Clone(raw), 0x55)
		v, tail, e := readElement(withTail)
		if e != nil || v.tag != 4 || len(v.content) != n || !bytes.Equal(v.full, raw) || !bytes.Equal(tail, []byte{0x55}) {
			t.Fatal("TLV", n, e)
		}
		if _, e = singleElement(raw, 4); e != nil {
			t.Fatal(e)
		}
		p := withTail
		if _, e = takeElement(&p, 4); e != nil || !bytes.Equal(p, []byte{0x55}) {
			t.Fatal("cursor")
		}
		if _, e = singleElement(withTail, 4); e != ErrInvalidMaterial {
			t.Fatal("tail")
		}
	}
	for _, raw := range [][]byte{nil, {4}, {4, 0x80, 0, 0}, {4, 0x81, 0}, {4, 0x81, 0x7f}, {4, 0x82, 0, 0x80}, {4, 0x82, 0xff, 0xff}, {4, 0xff}, {0x1f, 0}, {0, 0}, {4, 2, 0}, {4, 0x88, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}} {
		if _, _, e := readElement(raw); e != ErrInvalidMaterial {
			t.Fatal("malformed accepted", raw, e)
		}
	}
	raw := []byte{4, 0}
	original := bytes.Clone(raw)
	if _, e := takeElement(&raw, 2); e != ErrInvalidMaterial || !bytes.Equal(raw, original) {
		t.Fatal("wrong tag cursor")
	}
	for _, tc := range []struct {
		raw  []byte
		want bool
	}{{nil, false}, {[]byte{0}, true}, {[]byte{0x80}, true}, {[]byte{0xff}, true}, {[]byte{0, 0x80}, true}, {[]byte{0, 1}, false}, {[]byte{0xff, 0x80}, false}} {
		if minimalInteger(tc.raw) != tc.want {
			t.Fatal("minimal integer", tc)
		}
	}
}

// Catches accepting altered v3 order, extra/unique fields and incorrect raw ranges.
func TestIssuerEnvelope(t *testing.T) {
	in := fixtureInput(t)
	v := fixtureGoldens(t)["base"]
	raw, e := parseLeafEnvelope(in.Leaf.DER)
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range []struct {
		got  []byte
		name string
	}{{raw.tbs.full, "LeafTBS"}, {raw.issuer.full, "LeafIssuer"}, {raw.subject.full, "LeafSubject"}, {raw.spki.full, "LeafSPKI"}} {
		if !bytes.Equal(c.got, hexBytes(t, v[c.name])) {
			t.Fatal("raw field", c.name)
		}
	}
	if !bytes.Equal(raw.signature.content, append([]byte{0}, hexBytes(t, v["LeafSignature"])...)) {
		t.Fatal("signature range")
	}
	cert := fixtureParts(t, in.Leaf.DER)
	tbs := fixtureParts(t, cert[0])
	bad := [][]byte{append(bytes.Clone(in.Leaf.DER), 0), fixtureJoin(cert[:2]), fixtureJoin(append(append([][]byte{}, cert...), []byte{5, 0})), fixtureRewrite(t, in.Leaf.DER, "tbs", fixtureJoin(tbs[1:])), fixtureRewrite(t, in.Leaf.DER, "tbs", fixtureJoin(append(append([][]byte{}, tbs...), []byte{0x81, 1, 0})))}
	wrong := append([][]byte{}, tbs...)
	wrong[0] = []byte{0xa0, 3, 2, 1, 1}
	bad = append(bad, fixtureRewrite(t, in.Leaf.DER, "tbs", fixtureJoin(wrong)))
	for _, b := range bad {
		if _, e := parseLeafEnvelope(b); e != ErrInvalidMaterial {
			t.Fatal("bad envelope", e)
		}
	}
}
