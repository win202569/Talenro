package trust

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestJCSMatchesOfficialRFC8785Vectors catches locale sorting, whitespace,
// escaping, number rendering, and UTF-16 property-order regressions.
func TestJCSMatchesOfficialRFC8785Vectors(t *testing.T) {
	vectors := []struct {
		name, inputHash, outputHash string
	}{
		{name: "arrays", inputHash: "e503b6d71d1afa595b1c74b1016445c944cd89f90418066b23de1aeda7d17563", outputHash: "099601b171cafed97c333f8878d68e7f8c8f795412adb34b2fdcf0e7c7beac42"},
		{name: "french", inputHash: "03676a951cd8753ac62589f72eb2105cc782c33425418cfe1d517c111f6e5d5a", outputHash: "d99d0ebdcb0033cb858cfa830ae46bc0fb3309413b271f1da828c89901a27ed5"},
		{name: "structures", inputHash: "d66893805be1784116af50af3110d08766c70a6b4aad93374723f72346e7aaa6", outputHash: "605f65004ec2db7692522a0852c22f1c989e036d547e88963d1a3143cf3195d5"},
		{name: "unicode", inputHash: "4621864e014d4a805a563f55b9ea20aba4a2d2dc09c7394f625496998c00702c", outputHash: "0d99aad92a125196ff887876643fd3206786a84ddce2cee52ba4ad256d2381d3"},
		{name: "values", inputHash: "c4a041b503d6bc236036ef44db4dac499272f60fc22c40dc3b7a54870ba6f1c3", outputHash: "2d5e01a318d0f0879ab568c4be289c8b1f64ef8921a53c6277d5e069978baacb"},
		{name: "weird", inputHash: "a3a905266bd4a49a969274ea69baa14ee0c4af0ead926d6fa2b7612b4af75387", outputHash: "6af595a9aa80110b964b4de3f82a05fa6ae7423005019bacfa2620dddc4e94d1"},
	}
	for _, vector := range vectors {
		t.Run(vector.name, func(t *testing.T) {
			// #nosec G304 -- name comes from the closed literal allowlist above.
			input, err := os.ReadFile(filepath.Join("..", "..", "testdata", "crypto", "rfc8785", "input", vector.name+".json"))
			if err != nil {
				t.Fatalf("read official input: %v", err)
			}
			// #nosec G304 -- name comes from the closed literal allowlist above.
			want, err := os.ReadFile(filepath.Join("..", "..", "testdata", "crypto", "rfc8785", "output", vector.name+".json"))
			if err != nil {
				t.Fatalf("read official output: %v", err)
			}
			if gotHash := fmt.Sprintf("%x", sha256.Sum256(input)); gotHash != vector.inputHash {
				t.Fatalf("official input SHA-256 = %s, want %s", gotHash, vector.inputHash)
			}
			if gotHash := fmt.Sprintf("%x", sha256.Sum256(want)); gotHash != vector.outputHash {
				t.Fatalf("official output SHA-256 = %s, want %s", gotHash, vector.outputHash)
			}
			got, err := CanonicalizeJSON(input, new(any))
			if err != nil {
				t.Fatalf("CanonicalizeJSON: %v", err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("canonical bytes = %q, want exact official bytes %q", got, want)
			}
		})
	}
}

// TestJCSRejectsAmbiguousAndNonIJSONBeforeTransform catches parser differentials
// caused by duplicate names, unpaired surrogates, non-finite values, bad UTF-8,
// trailing JSON, and excessive nesting.
func TestJCSRejectsAmbiguousAndNonIJSONBeforeTransform(t *testing.T) {
	tests := []struct {
		name string
		body []byte
	}{
		{name: "duplicate", body: []byte(`{"a":1,"a":2}`)},
		{name: "unpaired high surrogate", body: []byte(`{"a":"\ud800"}`)},
		{name: "unpaired low surrogate", body: []byte(`{"a":"\udc00"}`)},
		{name: "non-finite overflow", body: []byte(`{"a":1e400}`)},
		{name: "NaN", body: []byte(`{"a":NaN}`)},
		{name: "Infinity", body: []byte(`{"a":Infinity}`)},
		{name: "bad UTF-8", body: []byte{'{', '"', 'a', '"', ':', '"', 0xff, '"', '}'}},
		{name: "trailing", body: []byte(`{}[]`)},
		{name: "too deep", body: []byte(strings.Repeat("[", 17) + strings.Repeat("]", 17))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := CanonicalizeJSON(test.body, new(any)); err == nil {
				t.Fatal("CanonicalizeJSON accepted unsafe JSON")
			}
		})
	}
}

// TestJCSTypedPayloadReDecodesExactCanonicalBytes catches canonicalizing before
// schema validation or returning bytes that no longer decode to the exact type.
func TestJCSTypedPayloadReDecodesExactCanonicalBytes(t *testing.T) {
	value, err := DecodePayloadV1([]byte(validPayloadJSON))
	if err != nil {
		t.Fatalf("DecodePayloadV1: %v", err)
	}
	got, err := CanonicalizePayloadV1(value)
	if err != nil {
		t.Fatalf("CanonicalizePayloadV1: %v", err)
	}
	const want = `{"audience":"7fa85f64-5717-4562-b3fc-2c963f66afa6","bundle_id":"d64cc450-b7eb-4575-9d8a-8a304096e719","bundle_locator":"ERERERERERERERERERERERERERERERERERERERERERE","bundle_version":"1","expires_at":"2026-08-10T12:00:00Z","issued_at":"2026-08-09T12:00:00Z","not_before":"2026-08-09T12:00:00Z","policy_snapshot":{"mode":"standard"},"schema_version":"talenro-config-bundle/v1","test_config":{"message":"hello","sequence":"1"}}`
	if string(got) != want {
		t.Fatalf("canonical payload = %s, want %s", got, want)
	}
	decoded, err := DecodePayloadV1(got)
	if err != nil || decoded.BundleID != value.BundleID || string(decoded.PolicySnapshot) != `{"mode":"standard"}` {
		t.Fatalf("canonical re-decode = %#v, %v", decoded, err)
	}
}
