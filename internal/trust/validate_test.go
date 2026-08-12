package trust

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

const validPayloadJSON = `{"schema_version":"talenro-config-bundle/v1","bundle_id":"d64cc450-b7eb-4575-9d8a-8a304096e719","bundle_locator":"ERERERERERERERERERERERERERERERERERERERERERE","bundle_version":"1","audience":"7fa85f64-5717-4562-b3fc-2c963f66afa6","issued_at":"2026-08-09T12:00:00Z","not_before":"2026-08-09T12:00:00Z","expires_at":"2026-08-10T12:00:00Z","policy_snapshot":{"mode":"standard"},"test_config":{"message":"hello","sequence":"1"}}`

// TestValidatePayloadRejectsStructuralBreaks catches accepting attacker-controlled
// members, ambiguous duplicate members, malformed UTF-8, trailing values, and
// numeric protocol integers that different implementations can round differently.
func TestValidatePayloadRejectsStructuralBreaks(t *testing.T) {
	tests := []struct {
		name string
		body []byte
	}{
		{name: "unknown top-level member", body: []byte(strings.Replace(validPayloadJSON, `"test_config":`, `"unknown":true,"test_config":`, 1))},
		{name: "unknown nested member", body: []byte(strings.Replace(validPayloadJSON, `"sequence":"1"`, `"sequence":"1","unknown":true`, 1))},
		{name: "duplicate member", body: []byte(strings.Replace(validPayloadJSON, `"bundle_id":`, `"schema_version":"other","bundle_id":`, 1))},
		{name: "numeric bundle version", body: []byte(strings.Replace(validPayloadJSON, `"bundle_version":"1"`, `"bundle_version":9007199254740993`, 1))},
		{name: "numeric sequence", body: []byte(strings.Replace(validPayloadJSON, `"sequence":"1"`, `"sequence":9007199254740993`, 1))},
		{name: "fractional timestamp", body: []byte(strings.Replace(validPayloadJSON, `2026-08-09T12:00:00Z`, `2026-08-09T12:00:00.000Z`, 1))},
		{name: "offset timestamp", body: []byte(strings.Replace(validPayloadJSON, `2026-08-09T12:00:00Z`, `2026-08-09T16:00:00+04:00`, 1))},
		{name: "trailing value", body: append([]byte(validPayloadJSON), []byte(` {}`)...)},
		{name: "invalid UTF-8", body: append(append([]byte(nil), []byte(validPayloadJSON[:len(validPayloadJSON)-1])...), 0xff, '}')},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := DecodePayloadV1(test.body); err == nil {
				t.Fatal("DecodePayloadV1 accepted structurally unsafe JSON")
			}
		})
	}
}

// TestValidateDirectDecodersRejectEscapedLoneSurrogates catches bypassing the
// raw I-JSON pass before encoding/json replaces invalid UTF-16 with RuneError.
func TestValidateDirectDecodersRejectEscapedLoneSurrogates(t *testing.T) {
	for _, surrogate := range []string{`\ud800`, `\udc00`} {
		payload := []byte(strings.Replace(validPayloadJSON, `"message":"hello"`, `"message":"`+surrogate+`"`, 1))
		if _, err := DecodePayloadV1(payload); err == nil {
			t.Fatalf("DecodePayloadV1 accepted lone surrogate %s", surrogate)
		}

		bundle := []byte(`{"payload_jcs":"` + surrogate + `","payload_sha256":"` + strings.Repeat("A", 43) +
			`","signer_key_id":"` + strings.Repeat("k", 22) + `","algorithm":"Ed25519","signature":"` +
			strings.Repeat("A", 86) + `","padding":""}`)
		if _, err := DecodeSignedBundleV1(bundle); err == nil {
			t.Fatalf("DecodeSignedBundleV1 accepted lone surrogate %s", surrogate)
		}
	}
}

// TestValidateDirectDecodersRequireEmptyValuedMembers catches treating omitted
// message or padding members as their valid, but wire-distinct, empty values.
func TestValidateDirectDecodersRequireEmptyValuedMembers(t *testing.T) {
	emptyMessage := []byte(strings.Replace(validPayloadJSON, `"message":"hello"`, `"message":""`, 1))
	if _, err := DecodePayloadV1(emptyMessage); err != nil {
		t.Fatalf("DecodePayloadV1 rejected present empty message: %v", err)
	}
	for _, member := range []string{
		"schema_version", "bundle_id", "bundle_locator", "bundle_version", "audience",
		"issued_at", "not_before", "expires_at", "policy_snapshot", "test_config",
	} {
		if _, err := DecodePayloadV1(removeTestJSONMember(t, []byte(validPayloadJSON), member)); err == nil {
			t.Fatalf("DecodePayloadV1 accepted omitted %s", member)
		}
	}
	for _, member := range []string{"message", "sequence"} {
		if _, err := DecodePayloadV1(removeNestedTestJSONMember(t, []byte(validPayloadJSON), "test_config", member)); err == nil {
			t.Fatalf("DecodePayloadV1 accepted omitted test_config.%s", member)
		}
	}

	bundle := SignedBundleV1{
		PayloadJCS: validPayloadJSON, PayloadSHA256: strings.Repeat("A", 43),
		SignerKeyID: strings.Repeat("k", 22), Algorithm: "Ed25519",
		Signature: strings.Repeat("A", 86), Padding: "",
	}
	body, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("marshal bundle: %v", err)
	}
	if _, err := DecodeSignedBundleV1(body); err != nil {
		t.Fatalf("DecodeSignedBundleV1 rejected present empty padding: %v", err)
	}
	for _, member := range []string{"payload_jcs", "payload_sha256", "signer_key_id", "algorithm", "signature", "padding"} {
		if _, err := DecodeSignedBundleV1(removeTestJSONMember(t, body, member)); err == nil {
			t.Fatalf("DecodeSignedBundleV1 accepted omitted %s", member)
		}
	}
}

func removeTestJSONMember(t *testing.T, body []byte, member string) []byte {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		t.Fatalf("decode test JSON: %v", err)
	}
	delete(object, member)
	changed, err := json.Marshal(object)
	if err != nil {
		t.Fatalf("encode test JSON: %v", err)
	}
	return changed
}

func removeNestedTestJSONMember(t *testing.T, body []byte, objectName, member string) []byte {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(body, &object); err != nil {
		t.Fatalf("decode test JSON: %v", err)
	}
	object[objectName] = removeTestJSONMember(t, object[objectName], member)
	changed, err := json.Marshal(object)
	if err != nil {
		t.Fatalf("encode test JSON: %v", err)
	}
	return changed
}

// TestValidatePayloadEnforcesFiniteDevicePolicyV1 catches widening the policy
// allowlist beyond the two exact device-policy-v1 shapes used by deviceauth.
func TestValidatePayloadEnforcesFiniteDevicePolicyV1(t *testing.T) {
	validPolicies := []string{
		`{"mode":"standard"}`,
		`{"mode":"trial_restricted","expires_at":"2026-08-10T12:00:00Z","max_devices":"1"}`,
	}
	for _, policy := range validPolicies {
		body := []byte(strings.Replace(validPayloadJSON, `{"mode":"standard"}`, policy, 1))
		if _, err := DecodePayloadV1(body); err != nil {
			t.Fatalf("valid policy %s: %v", policy, err)
		}
	}

	invalidPolicies := []string{
		`{}`,
		`{"mode":"other"}`,
		`{"mode":"standard","max_devices":"1"}`,
		`{"mode":"trial_restricted","expires_at":"2026-08-10T12:00:00Z"}`,
		`{"mode":"trial_restricted","expires_at":"2026-08-10T12:00:00Z","max_devices":1}`,
		`{"mode":"trial_restricted","expires_at":"2026-08-10T12:00:00.1Z","max_devices":"1"}`,
		`{"mode":"trial_restricted","expires_at":"2026-08-10T12:00:00Z","max_devices":"1","extra":true}`,
		`{"mode":"trial_restricted","mode":"standard","expires_at":"2026-08-10T12:00:00Z","max_devices":"1"}`,
	}
	for _, policy := range invalidPolicies {
		body := []byte(strings.Replace(validPayloadJSON, `{"mode":"standard"}`, policy, 1))
		if _, err := DecodePayloadV1(body); err == nil {
			t.Fatalf("accepted widened policy %s", policy)
		}
	}
}

// TestValidatePayloadEnforcesAllFieldAndByteBounds catches empty identifiers,
// non-canonical integers, invalid time ordering, and byte-counting message bugs.
func TestValidatePayloadEnforcesAllFieldAndByteBounds(t *testing.T) {
	base, err := DecodePayloadV1([]byte(validPayloadJSON))
	if err != nil {
		t.Fatalf("decode base: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*PayloadV1)
	}{
		{name: "schema", mutate: func(value *PayloadV1) { value.SchemaVersion = "talenro-config-bundle/v2" }},
		{name: "bundle ID", mutate: func(value *PayloadV1) { value.BundleID = "" }},
		{name: "locator length", mutate: func(value *PayloadV1) { value.BundleLocator = strings.Repeat("A", 42) }},
		{name: "locator alphabet", mutate: func(value *PayloadV1) { value.BundleLocator = strings.Repeat("+", 43) }},
		{name: "zero version", mutate: func(value *PayloadV1) { value.BundleVersion = "0" }},
		{name: "leading-zero version", mutate: func(value *PayloadV1) { value.BundleVersion = "01" }},
		{name: "overflow version", mutate: func(value *PayloadV1) { value.BundleVersion = "18446744073709551616" }},
		{name: "empty audience", mutate: func(value *PayloadV1) { value.Audience = "" }},
		{name: "not-before before issuance", mutate: func(value *PayloadV1) { value.NotBefore = "2026-08-09T11:59:59Z" }},
		{name: "expiry at not-before", mutate: func(value *PayloadV1) { value.ExpiresAt = value.NotBefore }},
		{name: "message over 256 bytes", mutate: func(value *PayloadV1) { value.TestConfig.Message = strings.Repeat("é", 129) }},
		{name: "leading-zero sequence", mutate: func(value *PayloadV1) { value.TestConfig.Sequence = "01" }},
		{name: "policy over 4096 bytes", mutate: func(value *PayloadV1) {
			value.PolicySnapshot = json.RawMessage(strings.Repeat(" ", 4096) + `{"mode":"standard"}`)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := base
			value.PolicySnapshot = bytes.Clone(base.PolicySnapshot)
			test.mutate(&value)
			if err := ValidatePayloadV1(value); err == nil {
				t.Fatal("ValidatePayloadV1 accepted invalid bounded field")
			}
		})
	}

	base.TestConfig.Message = strings.Repeat("é", 128)
	if err := ValidatePayloadV1(base); err != nil {
		t.Fatalf("256-byte message rejected: %v", err)
	}
	base.TestConfig.Message = ""
	base.TestConfig.Sequence = "0"
	if err := ValidatePayloadV1(base); err != nil {
		t.Fatalf("empty message and zero unsigned sequence rejected: %v", err)
	}
	base.TestConfig.Sequence = "18446744073709551615"
	if err := ValidatePayloadV1(base); err != nil {
		t.Fatalf("maximum uint64 sequence rejected: %v", err)
	}
}

// TestValidateSignedBundleRejectsMalformedEncodings catches accepting signed
// frames whose digest, signature, key identifier, payload, or padding can exceed
// the fixed wire/storage limits.
func TestValidateSignedBundleRejectsMalformedEncodings(t *testing.T) {
	valid := SignedBundleV1{
		PayloadJCS: validPayloadJSON, PayloadSHA256: strings.Repeat("A", 43),
		SignerKeyID: strings.Repeat("k", 22), Algorithm: "Ed25519",
		Signature: strings.Repeat("A", 86), Padding: "",
	}
	tests := []struct {
		name   string
		mutate func(*SignedBundleV1)
	}{
		{name: "empty payload", mutate: func(v *SignedBundleV1) { v.PayloadJCS = "" }},
		{name: "oversized payload", mutate: func(v *SignedBundleV1) { v.PayloadJCS = strings.Repeat("a", 65537) }},
		{name: "digest length", mutate: func(v *SignedBundleV1) { v.PayloadSHA256 = strings.Repeat("A", 42) }},
		{name: "short key ID", mutate: func(v *SignedBundleV1) { v.SignerKeyID = strings.Repeat("k", 15) }},
		{name: "long key ID", mutate: func(v *SignedBundleV1) { v.SignerKeyID = strings.Repeat("k", 65) }},
		{name: "key ID alphabet", mutate: func(v *SignedBundleV1) { v.SignerKeyID = strings.Repeat("+", 22) }},
		{name: "algorithm", mutate: func(v *SignedBundleV1) { v.Algorithm = "ed25519" }},
		{name: "signature length", mutate: func(v *SignedBundleV1) { v.Signature = strings.Repeat("A", 85) }},
		{name: "padding alphabet", mutate: func(v *SignedBundleV1) { v.Padding = "+" }},
		{name: "padding over bound", mutate: func(v *SignedBundleV1) { v.Padding = strings.Repeat("A", 65540) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := valid
			test.mutate(&value)
			if err := ValidateSignedBundleV1(value); err == nil {
				t.Fatal("ValidateSignedBundleV1 accepted malformed wire field")
			}
		})
	}

	unknown := []byte(`{"payload_jcs":"{}","payload_sha256":"` + strings.Repeat("A", 43) + `","signer_key_id":"` + strings.Repeat("k", 22) + `","algorithm":"Ed25519","signature":"` + strings.Repeat("A", 86) + `","padding":"","unknown":true}`)
	if _, err := DecodeSignedBundleV1(unknown); err == nil {
		t.Fatal("DecodeSignedBundleV1 accepted unknown member")
	}
	validJSON, err := json.Marshal(valid)
	if err != nil {
		t.Fatalf("marshal valid signed bundle: %v", err)
	}
	duplicate := bytes.Replace(validJSON, []byte(`"padding":""`), []byte(`"padding":"","padding":""`), 1)
	if _, err := DecodeSignedBundleV1(duplicate); err == nil {
		t.Fatal("DecodeSignedBundleV1 accepted duplicate member")
	}
}
