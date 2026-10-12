package nodeca

import (
	"bytes"
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"
)

func fixtureNameSize(n int) []byte {
	return fixtureTLV(0x30, fixtureTLV(0x31, fixtureJoin([][]byte{{6, 3, 0x55, 4, 3}, fixtureTLV(12, []byte(strings.Repeat("a", n-21)))})))
}
func TestCAFields(t *testing.T) {
	in := fixtureInput(false)
	c := fixtureCert(t, in)
	raw, e := parseRawCA(in.DER)
	if e != nil {
		t.Fatal(e)
	}
	got, e := checkCAFields(c, raw)
	if e != nil || got.SerialLength != 1 || got.Serial[0] != 128 || !got.NotBefore.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) || !got.NotAfter.Equal(time.Date(2036, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("projection", got, e)
	}
	for _, n := range []int{4096, 4097} {
		name := fixtureNameSize(n)
		if len(name) != n {
			t.Fatal("name fixture")
		}
		for _, path := range []string{"subject", "issuer"} {
			changed := fixtureRewrite(t, in, path, name)
			cert := fixtureCert(t, changed)
			r, e := parseRawCA(changed.DER)
			if e != nil {
				t.Fatal(e)
			}
			_, e = checkCAFields(cert, r)
			want := error(nil)
			if n == 4097 {
				want = ErrUnsupportedCAProfile
			}
			if !errors.Is(e, want) {
				t.Fatalf("name %s/%d %v", path, n, e)
			}
		}
	}
	for _, path := range []string{"subject", "issuer"} {
		changed := fixtureRewrite(t, in, path, []byte{0x30, 0})
		cert := fixtureCert(t, changed)
		r, e := parseRawCA(changed.DER)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = checkCAFields(cert, r); !errors.Is(e, ErrUnsupportedCAProfile) {
			t.Fatal("empty Name", e)
		}
	}
	copied := *c
	copied.Version = 2
	if _, e = checkCAFields(&copied, raw); !errors.Is(e, ErrUnsupportedCAProfile) {
		t.Fatal("version", e)
	}
	copied = *c
	copied.RawSubject = []byte{0x30, 0}
	if _, e = checkCAFields(&copied, raw); !errors.Is(e, ErrInvalidMaterial) {
		t.Fatal("Name mismatch", e)
	}
	for _, v := range []struct {
		b    []byte
		want error
	}{
		{[]byte{1}, nil}, {append([]byte{0x7f}, bytes.Repeat([]byte{0xff}, 19)...), nil},
		{append([]byte{0, 0x80}, bytes.Repeat([]byte{0xff}, 18)...), nil},
		{append([]byte{0, 0x80}, bytes.Repeat([]byte{0xff}, 19)...), ErrUnsupportedCAProfile},
		{[]byte{0}, ErrUnsupportedCAProfile}, {[]byte{255}, ErrUnsupportedCAProfile}, {[]byte{0, 1}, ErrInvalidMaterial},
	} {
		r := raw
		r.serial = derElement{tag: 2, full: fixtureTLV(2, v.b), content: v.b}
		cert := *c
		cert.SerialNumber = new(big.Int).SetBytes(v.b)
		if len(v.b) > 0 && v.b[0]&128 != 0 {
			cert.SerialNumber.SetInt64(-1)
		}
		_, e := checkCAFields(&cert, r)
		if !errors.Is(e, v.want) {
			t.Fatalf("serial %x got %v want %v", v.b, e, v.want)
		}
	}
	for _, year := range []int{0, 1, 1949, 1950, 2049, 2050, 9999, 10000} {
		dt := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		tag, layout := byte(0x18), "20060102150405Z"
		if year >= 1950 && year <= 2049 {
			tag = 0x17
			layout = "060102150405Z"
		}
		r := derElement{tag: tag, content: []byte(dt.Format(layout))}
		want := year >= 1 && year <= 9999
		if canonicalTime(r, dt) != want {
			t.Fatalf("year %d", year)
		}
		if want {
			r.tag ^= 1
			if canonicalTime(r, dt) {
				t.Fatal("wrong time tag")
			}
		}
	}
	for _, edit := range []func(*derElement, *time.Time){
		func(r *derElement, dt *time.Time) { r.content = []byte("260101000000+0000") },
		func(r *derElement, dt *time.Time) { r.content = []byte("260101000000.0Z") },
		func(r *derElement, dt *time.Time) { *dt = dt.Add(time.Nanosecond) },
		func(r *derElement, dt *time.Time) {
			*dt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.FixedZone("offset", 3600))
		},
	} {
		dt := c.NotBefore
		r := raw.notBefore
		edit(&r, &dt)
		if canonicalTime(r, dt) {
			t.Fatal("noncanonical time accepted")
		}
	}
	for _, na := range []time.Time{c.NotBefore, c.NotBefore.Add(-time.Second)} {
		changed := fixtureRewrite(t, in, "notAfter", fixtureTLV(0x17, []byte(na.Format("060102150405Z"))))
		cert := fixtureCert(t, changed)
		r, e := parseRawCA(changed.DER)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = checkCAFields(cert, r); !errors.Is(e, ErrUnsupportedCAProfile) {
			t.Fatal("window", e)
		}
	}
	copied = *c
	copied.SerialNumber = big.NewInt(129)
	if _, e = checkCAFields(&copied, raw); !errors.Is(e, ErrInvalidMaterial) {
		t.Fatal("serial mismatch", e)
	}
	copied = *c
	copied.NotBefore = c.NotBefore.Add(time.Second)
	if _, e = checkCAFields(&copied, raw); !errors.Is(e, ErrInvalidMaterial) {
		t.Fatal("time mismatch", e)
	}
}
func TestCANotLeafLifetime(t *testing.T) {
	in := fixtureInput(false)
	for _, na := range []string{"260101000001Z", "360101000000Z"} {
		changed := fixtureRewrite(t, in, "notAfter", fixtureTLV(0x17, []byte(na)))
		c := fixtureCert(t, changed)
		r, e := parseRawCA(changed.DER)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = checkCAFields(c, r); e != nil {
			t.Fatal("CA lifetime constrained like leaf", e)
		}
	}
	for _, pair := range [][2]string{{"00010101000000Z", "99991231235959Z"}, {"19491231235959Z", "500101000000Z"}, {"491231235959Z", "20500101000000Z"}} {
		nbTag, naTag := byte(0x17), byte(0x17)
		if len(pair[0]) == 15 {
			nbTag = 0x18
		}
		if len(pair[1]) == 15 {
			naTag = 0x18
		}
		changed := fixtureRewrite(t, in, "notBefore", fixtureTLV(nbTag, []byte(pair[0])))
		changed = fixtureRewrite(t, changed, "notAfter", fixtureTLV(naTag, []byte(pair[1])))
		c := fixtureCert(t, changed)
		r, e := parseRawCA(changed.DER)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = checkCAFields(c, r); e != nil {
			t.Fatal("canonical boundary", e)
		}
	}
}
