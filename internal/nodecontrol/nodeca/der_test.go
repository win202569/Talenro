package nodeca

import (
	"bytes"
	"errors"
	"testing"
)

func TestCADERTLV(t *testing.T) {
	for _, n := range []int{0, 1, 127, 128, 255, 256, 65536} {
		b := fixtureTLV(4, make([]byte, n))
		v, tail, e := readElement(append(b, 0xff))
		if e != nil || len(v.content) != n || !bytes.Equal(tail, []byte{0xff}) || !bytes.Equal(v.full, b) {
			t.Fatalf("length %d %v", n, e)
		}
	}
	for _, b := range [][]byte{nil, {4}, {0x1f, 0}, {4, 0x80, 0, 0}, {4, 0x81, 1, 0}, {4, 0x82, 0, 0x80}, {4, 0x89, 1, 0, 0, 0, 0, 0, 0, 0, 0}, {4, 0x88, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, {4, 2, 1}} {
		if _, _, e := readElement(b); !errors.Is(e, ErrInvalidMaterial) {
			t.Fatalf("accepted %x %v", b, e)
		}
	}
	if _, e := singleElement([]byte{4, 0, 0}, 4); !errors.Is(e, ErrInvalidMaterial) {
		t.Fatal("tail", e)
	}
	if _, e := singleElement([]byte{2, 0}, 4); !errors.Is(e, ErrInvalidMaterial) {
		t.Fatal("tag", e)
	}
	if !minimalInteger([]byte{0, 128}) || !minimalInteger([]byte{255, 127}) || minimalInteger(nil) || minimalInteger([]byte{0, 1}) || minimalInteger([]byte{255, 128}) {
		t.Fatal("integer minimality")
	}
	cursor := []byte{4, 0, 2, 1, 1}
	if _, e := takeElement(&cursor, 4); e != nil || !bytes.Equal(cursor, []byte{2, 1, 1}) {
		t.Fatal("cursor")
	}
	before := bytes.Clone(cursor)
	if _, e := takeElement(&cursor, 4); !errors.Is(e, ErrInvalidMaterial) || !bytes.Equal(cursor, before) {
		t.Fatal("failure consumed input")
	}
}
