package stateformat

import (
	"encoding/base64"
	"strings"
	"testing"
)

// Removing the exact decimal/range checks must reject these cases, not round them.
func TestPrimitiveUExact(t *testing.T) {
	for _, tc := range []struct {
		s    string
		want uint64
	}{
		{"1", 1}, {"9007199254740991", 9007199254740991}, {"9007199254740992", 9007199254740992},
		{"9007199254740993", 9007199254740993}, {"9223372036854775807", 9223372036854775807},
	} {
		if got, err := parseU(tc.s); err != nil || got != tc.want {
			t.Errorf("%s: got %d, %v", tc.s, got, err)
		}
	}
	for _, s := range []string{"", "0", "01", "+1", "-1", "1e2", "1.0", " 1", "1 ", "１", "9223372036854775808", "18446744073709551615", "100000000000000000000"} {
		if got, err := parseU(s); err != ErrInvalidFormat || got != 0 {
			t.Errorf("accepted %q: %d %v", s, got, err)
		}
	}
}

// Accepting noncanonical encodings or a repaired value breaks these assertions.
func TestPrimitiveEncodings(t *testing.T) {
	t.Run("digest", func(t *testing.T) {
		d, err := parseDigest(strings.Repeat("ab", 32))
		if err != nil || d[0] != 0xab || d[31] != 0xab {
			t.Fatal(d, err)
		}
		for _, s := range []string{"", strings.Repeat("00", 32), strings.Repeat("AB", 32), strings.Repeat("ab", 31), strings.Repeat("ab", 33), strings.Repeat("ag", 32)} {
			if d, err := parseDigest(s); err != ErrInvalidFormat || d != ([32]byte{}) {
				t.Errorf("accepted %q", s)
			}
		}
	})
	t.Run("key", func(t *testing.T) {
		raw := make([]byte, 32)
		raw[0] = 1
		good := base64.RawURLEncoding.EncodeToString(raw)
		if v, err := parseKey(good); err != nil || v[0] != 1 || v[31] != 0 {
			t.Fatal(v, err)
		}
		for _, s := range []string{"", good + "=", good[:42] + "B", good[:42], strings.Repeat("A", 43), "+" + good[1:], good + "\n"} {
			if v, err := parseKey(s); err != ErrInvalidFormat || v != ([32]byte{}) {
				t.Errorf("accepted %q", s)
			}
		}
	})
	t.Run("signature", func(t *testing.T) {
		raw := make([]byte, 64)
		raw[0] = 2
		good := base64.RawURLEncoding.EncodeToString(raw)
		if v, err := parseSignature(good); err != nil || v[0] != 2 || v[63] != 0 {
			t.Fatal(v, err)
		}
		for _, s := range []string{"", good + "==", good[:85] + "B", good[:85], strings.Repeat("A", 86), "/" + good[1:], good + "\r\n"} {
			if v, err := parseSignature(s); err != ErrInvalidFormat || v != ([64]byte{}) {
				t.Errorf("accepted %q", s)
			}
		}
	})
	t.Run("time", func(t *testing.T) {
		for _, s := range []string{"0001-01-01T00:00:01Z", "2024-02-29T23:59:59Z", "9999-12-31T23:59:59Z"} {
			if v, err := parseTime(s); err != nil || v.Format("2006-01-02T15:04:05Z") != s {
				t.Errorf("valid %q: %v", s, err)
			}
		}
		for _, s := range []string{"0001-01-01T00:00:00Z", "0000-01-01T00:00:00Z", "10000-01-01T00:00:00Z", "2023-02-29T00:00:00Z", "2024-01-01T00:00:60Z", "2024-01-01T00:00:00+00:00", "2024-01-01T00:00:00.0Z", "2024-1-01T00:00:00Z", "2024-01-01t00:00:00z"} {
			if v, err := parseTime(s); err != ErrInvalidFormat || !v.IsZero() {
				t.Errorf("accepted %q", s)
			}
		}
	})
	t.Run("uuid", func(t *testing.T) {
		for _, s := range []string{"12345678-1234-0234-8234-123456789abc", "12345678-1234-f234-b234-123456789abc"} {
			if v, err := parseUUID(s); err != nil || v.String() != s {
				t.Errorf("valid %q: %v", s, err)
			}
		}
		for _, s := range []string{"00000000-0000-0000-0000-000000000000", "12345678-1234-4234-7234-123456789abc", "12345678-1234-4234-c234-123456789abc", "12345678-1234-4234-8234-123456789ABC", "12345678123442348234123456789abc", "{12345678-1234-4234-8234-123456789abc}"} {
			if v, err := parseUUID(s); err != ErrInvalidFormat || v != ([16]byte{}) {
				t.Errorf("accepted %q", s)
			}
		}
	})
	t.Run("id_and_build", func(t *testing.T) {
		for _, s := range []string{"a", "A0._-", strings.Repeat("x", 64)} {
			if !validID(s, 64) {
				t.Errorf("valid ID %q", s)
			}
		}
		for _, s := range []string{"", "-a", ".a", "_a", "a/b", "a b", "é", strings.Repeat("x", 65)} {
			if validID(s, 64) {
				t.Errorf("accepted ID %q", s)
			}
		}
		if !validID(strings.Repeat("x", 128), 128) || validID(strings.Repeat("x", 129), 128) {
			t.Fatal("ID128 bound")
		}
		for _, s := range []string{"0.0.0", "1.2.3", "2147483647.2147483647.2147483647"} {
			if !validBuild(s) {
				t.Errorf("valid build %q", s)
			}
		}
		for _, s := range []string{"", "1.2", "1.2.3.4", "01.2.3", "1.02.3", "1.2.03", "v1.2.3", "1.2.3-alpha", "1.2.3+meta", "2147483648.0.0", "1.-2.3", "1.2.3 ", strings.Repeat("1", 65)} {
			if validBuild(s) {
				t.Errorf("accepted build %q", s)
			}
		}
	})
}
