package deviceauth

import (
	"bytes"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"testing"
)

func TestPoPTranscriptMatchesIndependentLiteral(t *testing.T) {
	t.Parallel()

	input := ProofInput{
		ProtocolVersion:  "device-pop-v1",
		Challenge:        [32]byte{1},
		GrantDigest:      [32]byte{2},
		SigningPublicKey: [32]byte{3},
		HPKEPublicKey:    [32]byte{4},
		Operation:        "register_device",
		Audience:         "https://api.example.test",
		RequestNonce:     [32]byte{5},
	}
	want, err := hex.DecodeString("54414c454e524f2d4445564943452d504f502d5631000000000d6465766963652d706f702d763101000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000000000f72656769737465725f6465766963650000001868747470733a2f2f6170692e6578616d706c652e746573740500000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	got := ProofBytes(input)
	if got == nil {
		t.Fatal("ProofBytes rejected valid input")
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("proof transcript = %x, want independent literal %x", got, want)
	}
}

func TestPoPVerificationBindsEveryFieldAndSignature(t *testing.T) {
	t.Parallel()

	publicSigning, privateSigning, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	defer clear(privateSigning)
	hpkePrivate, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	hpkePrivateBytes := hpkePrivate.Bytes()
	defer clear(hpkePrivateBytes)

	input := ProofInput{
		ProtocolVersion: "device-pop-v1",
		Challenge:       [32]byte{1, 2, 3},
		GrantDigest:     [32]byte{4, 5, 6},
		Operation:       "register_device",
		Audience:        "https://api.example.test",
		RequestNonce:    [32]byte{7, 8, 9},
	}
	copy(input.SigningPublicKey[:], publicSigning)
	copy(input.HPKEPublicKey[:], hpkePrivate.PublicKey().Bytes())
	proof := ProofBytes(input)
	if proof == nil {
		t.Fatal("ProofBytes rejected valid generated keys")
	}
	signed := ed25519.Sign(privateSigning, proof)
	clear(proof)
	var signature [64]byte
	copy(signature[:], signed)
	clear(signed)
	if err := verifyProof(input, signature, "https://api.example.test"); err != nil {
		t.Fatalf("valid proof: %v", err)
	}

	mutations := []struct {
		name   string
		mutate func(*ProofInput)
	}{
		{name: "protocol", mutate: func(value *ProofInput) { value.ProtocolVersion = "device-pop-v2" }},
		{name: "challenge", mutate: func(value *ProofInput) { value.Challenge[0] ^= 0xff }},
		{name: "grant digest", mutate: func(value *ProofInput) { value.GrantDigest[0] ^= 0xff }},
		{name: "signing public key", mutate: func(value *ProofInput) { value.SigningPublicKey[0] ^= 0xff }},
		{name: "HPKE public key", mutate: func(value *ProofInput) { value.HPKEPublicKey[0] ^= 0xff }},
		{name: "operation", mutate: func(value *ProofInput) { value.Operation = "rotate_device" }},
		{name: "audience", mutate: func(value *ProofInput) { value.Audience = "https://other.example.test" }},
		{name: "request nonce", mutate: func(value *ProofInput) { value.RequestNonce[0] ^= 0xff }},
	}
	for _, test := range mutations {
		t.Run(test.name, func(t *testing.T) {
			mutation := input
			test.mutate(&mutation)
			if err := verifyProof(mutation, signature, "https://api.example.test"); !errors.Is(err, ErrInvalidProof) {
				t.Fatalf("mutated proof error = %v, want fixed invalid-proof error", err)
			}
		})
	}
	badSignature := signature
	badSignature[0] ^= 0xff
	if err := verifyProof(input, badSignature, "https://api.example.test"); !errors.Is(err, ErrInvalidProof) {
		t.Fatalf("substituted signature error = %v, want fixed invalid-proof error", err)
	}
}

func TestPoPRejectsMalformedStringsAndNonCanonicalAudience(t *testing.T) {
	t.Parallel()

	valid := ProofInput{
		ProtocolVersion: "device-pop-v1", Challenge: [32]byte{1}, GrantDigest: [32]byte{2}, SigningPublicKey: [32]byte{3},
		HPKEPublicKey: [32]byte{4}, Operation: "register_device", Audience: "https://api.example.test", RequestNonce: [32]byte{5},
	}
	invalidUTF8 := string([]byte{0xff})
	for _, input := range []ProofInput{
		func() ProofInput { value := valid; value.ProtocolVersion = ""; return value }(),
		func() ProofInput { value := valid; value.Operation = invalidUTF8; return value }(),
		func() ProofInput {
			value := valid
			value.Audience = string(bytes.Repeat([]byte{'a'}, 2049))
			return value
		}(),
	} {
		if proof := ProofBytes(input); proof != nil {
			t.Fatal("malformed transcript produced proof bytes")
		}
	}

	for _, audience := range []string{
		"http://api.example.test", "https://user@api.example.test", "https://api.example.test/path",
		"https://api.example.test?query=1", "https://api.example.test#fragment", "https:api.example.test",
	} {
		input := valid
		input.Audience = audience
		if err := verifyProof(input, [64]byte{1}, audience); !errors.Is(err, ErrInvalidProof) {
			t.Fatalf("audience %q error = %v, want fixed invalid-proof error", audience, err)
		}
	}
}

func TestTask18PublicOriginAcceptsControlledLoopbackHTTP(t *testing.T) {
	for _, origin := range []string{
		"http://localhost:8080",
		"http://LOCALHOST:1",
		"http://127.0.0.1:443",
		"http://127.255.255.254:65535",
		"http://[::1]:8080",
		"http://localhost:8080/",
		"https://api.example.test",
		"https://api.example.test/",
	} {
		t.Run(origin, func(t *testing.T) {
			if !validPublicOrigin(origin) {
				t.Fatalf("controlled public origin was rejected: %q", origin)
			}
		})
	}
}

func TestTask18PublicOriginRejectsUnsafeHTTPAndURLComponents(t *testing.T) {
	for _, origin := range []string{
		"http://example.test:8080",
		"http://localhost.:8080",
		"http://localhost.example:8080",
		"http://192.0.2.1:8080",
		"http://[2001:db8::1]:8080",
		"http://user@localhost:8080",
		"http://localhost:8080/path",
		"http://localhost:8080?source=CANARY",
		"http://localhost:8080?",
		"http://localhost:8080#CANARY",
		"http:localhost:8080",
		"http://localhost:",
		"http://localhost:0",
		"http://localhost:08080",
		"http://localhost:65536",
		"http://localhost:https",
	} {
		t.Run(origin, func(t *testing.T) {
			if validPublicOrigin(origin) {
				t.Fatalf("unsafe public origin was accepted: %q", origin)
			}
		})
	}
}
